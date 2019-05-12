package livy

import (
	"math"
	"math/cmplx"
	"sort"
)

type DyadicFunc func(c *Context, a Val, b Val, axis int) Val

var StandardDyadics = map[string]DyadicFunc{
	"member": dyadicMember,
	"e":      dyadicMember,
	"j":      WrapMatMatDyadic(WrapCxDyadic(cxcxJ)),
	"rect":   WrapMatMatDyadic(WrapCxDyadic(cxcxRect)),

	"rho": dyadicRho,
	"p":   dyadicRho,

	"transpose": dyadicTranspose,
	",":         dyadicCatenate,
	"laminate":  dyadicLaminate,
	"rot":       dyadicRot,
	"take":      dyadicTake,
	"drop":      dyadicDrop,
	`/`:         dyadicCompress,
	`\`:         dyadicExpand,

	"==": WrapMatMatDyadic(WrapCxBoolDyadic(
		func(a, b complex128) bool { return a == b })),
	"!=": WrapMatMatDyadic(WrapCxBoolDyadic(
		func(a, b complex128) bool { return a != b })),
	"<": WrapMatMatDyadic(WrapFloatBoolDyadic(
		func(a, b float64) bool { return a < b })),
	">": WrapMatMatDyadic(WrapFloatBoolDyadic(
		func(a, b float64) bool { return a > b })),
	"<=": WrapMatMatDyadic(WrapFloatBoolDyadic(
		func(a, b float64) bool { return a <= b })),
	">=": WrapMatMatDyadic(WrapFloatBoolDyadic(
		func(a, b float64) bool { return a >= b })),
	"and": WrapMatMatDyadic(WrapFloatBoolDyadic(ffand)),
	"or":  WrapMatMatDyadic(WrapFloatBoolDyadic(ffor)),
	"xor": WrapMatMatDyadic(WrapFloatBoolDyadic(ffxor)),

	"+": WrapMatMatDyadic(WrapCxDyadic(
		func(a, b complex128) complex128 { return a + b })),
	"-": WrapMatMatDyadic(WrapCxDyadic(
		func(a, b complex128) complex128 { return a - b })),
	"*": WrapMatMatDyadic(WrapCxDyadic(
		func(a, b complex128) complex128 { return a * b })),
	"div": WrapMatMatDyadic(WrapCxDyadic(
		func(a, b complex128) complex128 { return a / b })),
	"**": WrapMatMatDyadic(WrapCxDyadic(
		func(a, b complex128) complex128 { return cmplx.Pow(a, b) })),
	"remainder": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return math.Remainder(a, b) })),
	"mod": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return math.Mod(a, b) })),
	"atan": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return math.Atan2(a, b) })),
	"copysign": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return math.Copysign(a, b) })),
	"dim": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return math.Dim(a, b) })),
	"hypot": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return math.Hypot(a, b) })),
	"isInf": WrapMatMatDyadic(WrapFloatBoolDyadic(
		func(a, b float64) bool { return math.IsInf(a, int(b)) })),
	"jn": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return math.Jn(int(a), b) })),
	"yn": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return math.Yn(int(a), b) })),
}

var Zero = &Num{0.0}
var One = &Num{1.0}

// If not listed, we will use Zero.
// TODO: have a way to user-define an identity operator (into the Context).
var IdentityValueOfDyadic = map[string]Val{
	"+":   Zero,
	"-":   Zero,
	"or":  Zero,
	"!=":  Zero,
	"<":   Zero,
	">":   Zero,
	"*":   One,
	"div": One,
	"**":  One,
	"and": One,
	"==":  One,
	">=":  One,
	"<=":  One,
}

// mod forcing positive result, since I don't actually know what Go does.
func mod(x int, modulus int) int {
	if modulus < 1 {
		Log.Panicf("Nonpositive modulus: %d", modulus)
	}
	return ((x % modulus) + modulus) % modulus
}

func ffand(a, b float64) bool {
	x := float2bool(a)
	y := float2bool(b)
	return x && y
}
func ffor(a, b float64) bool {
	x := float2bool(a)
	y := float2bool(b)
	return x || y
}
func ffxor(a, b float64) bool {
	x := float2bool(a)
	y := float2bool(b)
	return xor(x, y)
	if x {
		return !y
	}
	return y
}

func MkOuterProduct(name string, fn DyadicFunc) DyadicFunc {
	return func(c *Context, a Val, b Val, axis int) Val {
		aa := GetVectorOfScalarVals(a)
		bb := GetVectorOfScalarVals(b)
		sz := len(aa) * len(bb)
		vec := make([]Val, sz)
		for ia, fa := range aa {
			for ib, fb := range bb {
				x := fn(c, fa, fb, -1)
				vec[ia*len(bb)+ib] = x
			}
		}
		return &Mat{M: vec, S: []int{len(aa), len(bb)}}
	}
}

func MkInnerProduct(name string, fn1, fn2 DyadicFunc) DyadicFunc {
	return func(c *Context, a Val, b Val, axis int) Val {
		mat1, ok := a.(*Mat)
		if !ok {
			Log.Panicf("LHS of inner product %q not a matrix: %v", name, a)
		}

		mat2, ok := b.(*Mat)
		if !ok {
			Log.Panicf("RHS of inner product %q not a matrix: %v", name, b)
		}

		vec1, vec2 := mat1.M, mat2.M
		shape1, shape2 := mat1.S, mat2.S
		rank1, rank2 := len(shape1), len(shape2)
		if rank1 < 1 {
			Log.Panicf("LHS of inner product %q has rank 0: %v", name, a)
		}
		if rank2 < 1 {
			Log.Panicf("RHS of inner product %q has rank 0: %v", name, b)
		}
		if shape1[rank1-1] != shape2[0] {
			Log.Panicf("Dimension conflict in inner product %q: LHS is shape %v; RHS is shape %v", name, shape1, shape2)
		}

		var outShape []int
		for _, sz := range shape1[:rank1-1] {
			outShape = append(outShape, sz)
		}
		for _, sz := range shape2[1:] {
			outShape = append(outShape, sz)
		}
		outVec := make([]Val, MulReduce(outShape))
		innerStride1 := 1
		innerStride2 := MulReduce(shape2[1:])
		innerLength := shape2[0]
		Log.Printf("innerStride 1:%d 2:%d innerLength:%d", innerStride1, innerStride2, innerLength)

		var recurse func(shape1, shape2 []int, off1, off2 int, outShape []int, outOff int)
		recurse = func(shape1, shape2 []int, off1, off2 int, outShape []int, outOff int) {
			rank1, rank2, outRank := len(shape1), len(shape2), len(outShape)
			Log.Printf("Rank(%d,%d -> %d) : shape( %v , %v -> %v ) : off (%d,%d -> %d)", rank1, rank2, outRank, shape1, shape2, outShape, off1, off2, outOff)
			if outRank == 0 {
				j := innerLength - 1
				rhs := fn2(c, vec1[off1+j*innerStride1], vec2[off2+j*innerStride2], -1)
				for i := innerLength - 2; i >= 0; i-- {
					Log.Printf(" rhs=%v [i=%d] vec1[%d] vec2[%d]", rhs, i, off1+i*innerStride1, off2+i*innerStride2)
					lhs := fn2(c, vec1[off1+i*innerStride1], vec2[off2+i*innerStride2], -1)
					rhs = fn1(c, lhs, rhs, -1)
				}
				outVec[outOff] = rhs
			} else if rank1 == 1 {
				// Stop using shape1 and start using shape2 when rank1 == 1.
				stride2, outStride := MulReduce(shape2[1:]), MulReduce(outShape[1:])
				for i := 0; i < shape2[0]; i++ {
					recurse(shape1, shape2[1:], off1, off2+i*stride2, outShape[1:], outOff+i*outStride)
				}
			} else {
				stride1, outStride := MulReduce(shape1[1:]), MulReduce(outShape[1:])
				for i := 0; i < shape1[0]; i++ {
					recurse(shape1[1:], shape2, off1+i*stride1, off2, outShape[1:], outOff+i*outStride)
				}
			}
		}
		recurse(shape1, shape2[1:], 0, 0, outShape, 0)
		if len(outShape) == 0 {
			return outVec[0] // Return scalar.
		} else {
			return &Mat{M: outVec, S: outShape}
		}
	}
}

func MkEachOpDyadic(name string, fn DyadicFunc) DyadicFunc {
	return func(c *Context, a, b Val, axis int) Val {
		if axis != -1 {
			Log.Panicf("dyadic ~ op: cannot use axis: %d", axis)
		}
		amat, aok := a.(*Mat)
		bmat, bok := b.(*Mat)
		var vec []Val
		switch {
		case aok && bok:
			if len(amat.S) != len(bmat.S) {
				Log.Panicf("left and right matrix need same rank, but got shapes %v and %v", amat.S, bmat.S)
			}
			// TODO EQ
			if MulReduce(amat.S) != MulReduce(bmat.S) {
				Log.Panicf("left and right matrix need same shape, but got shapes %v and %v", amat.S, bmat.S)
			}
			for i, e := range amat.M {
				x := fn(c, e, bmat.M[i], -1)
				vec = append(vec, x)
			}
			return &Mat{vec, amat.S}
		case aok:
			for _, e := range amat.M {
				x := fn(c, e, b, -1)
				vec = append(vec, x)
			}
			return &Mat{vec, amat.S}
		case bok:
			for _, e := range bmat.M {
				x := fn(c, a, e, -1)
				vec = append(vec, x)
			}
			return &Mat{vec, bmat.S}
		default:
			return fn(c, a, b, -1)
		}
	}
}

func MkReduceOrScanOp(name string, fn DyadicFunc, identity Val, toScan bool) MonadicFunc {
	verb := "reduce"
	if toScan {
		verb = "scan"
	}
	return func(c *Context, a Val, axis int) Val {
		mat, ok := a.(*Mat)
		if !ok {
			Log.Panicf("Cannot %s %s on non-matrix: %s", name, verb, a)
		}
		oldRank := len(mat.S)
		oldShape := mat.S
		Log.Printf("oldShape %v oldRank %v for matrix %v", oldShape, oldRank, mat)
		if oldRank == 0 {
			Log.Panicf("Cannot %s %s on scalar: %s", name, verb, mat)
		}
		if axis < 0 {
			axis += oldRank
		}
		if axis < 0 || axis > oldRank-1 {
			Log.Panicf("Reduce axis [%d] is bad for %s %s of rank %d", axis, name, verb, oldRank)
		}

		var newShape []int
		for i := 0; i < oldRank; i++ {
			if i == axis {
				if toScan {
					newShape = append(newShape, oldShape[i])
				}
			} else {
				newShape = append(newShape, oldShape[i])
			}
		}

		newVecLen := MulReduce(newShape)
		newVec := make([]Val, newVecLen)
		oldVec := mat.M

		reduceStride, reduceLen := MulReduce(oldShape[axis+1:]), oldShape[axis]
		Log.Printf("Reduce Stride = %d", reduceStride)
		revAxis := oldRank - axis

		var reduce func(oldShape []int, oldOffset int, newShape []int, newOffset int, reduceOffset int)
		reduce = func(oldShape []int, oldOffset int, newShape []int, newOffset int, reduceOffset int) {
			rank := len(oldShape)
			Log.Printf("[[%d]] ;; old %d @ %v ;; new %d @ %v ;; {ro=%d,rs=%d,revAxis=%d}", rank, oldOffset, oldShape, newOffset, newShape, reduceOffset, reduceStride, revAxis)
			if rank == 0 {
				var reduction Val

				if reduceLen == 0 {
					reduction = identity
				} else {
					// j is 0:
					reduction = oldVec[oldOffset]
					if toScan {
						newVec[newOffset] = reduction
						Log.Printf("Scan...0  newVec: %v [[%d; %s]]", newVec, newOffset, reduction)
					}
					// other j's:
					for j := 1; j < reduceLen; j++ {
						Log.Printf("...... %d [[%d; %s]]", j, newOffset, reduction)
						reduction = fn(c, reduction, oldVec[oldOffset+j*reduceStride], DefaultAxis)
						if toScan {
							newVec[newOffset+j*reduceStride] = reduction
							Log.Printf("Scan...%d  newVec: %v [[%d; %s]]", j, newVec, newOffset+j*reduceStride, reduction)
						}
					}
				}

				if !toScan {
					newVec[newOffset] = reduction
					Log.Printf("Reduction...  newVec: %v [[%d; %s]]", newVec, newOffset, reduction)
				}
			} else if len(oldShape) == revAxis {
				// This is the old dimension we reduce.
				// It exists in the oldShape but not in the newShape, if reducing.
				// We do not iterate i here -- that will happen above when rank==0.
				if reduceStride != MulReduce(oldShape[1:]) {
					panic(0)
				}
				if toScan {
					newShape = newShape[1:]
				}
				Log.Printf("111 reduceOffset = %d ;; old %d @ %v ;; new %d @ %v -->", oldOffset, reduceOffset, oldShape, newOffset, newShape)
				reduce(oldShape[1:], oldOffset, newShape, newOffset, oldOffset)
			} else {
				for i := 0; i < oldShape[0]; i++ {
					oldStride := MulReduce(oldShape[1:])
					newStride := MulReduce(newShape[1:])
					reduce(oldShape[1:], oldOffset+i*oldStride, newShape[1:], newOffset+i*newStride, reduceOffset)
				}
			}
		}

		reduce(oldShape, 0 /*oldOffset*/, newShape, 0 /*newOffset*/, -1000000000 /* reduceOffset: cause panic if used in any reasonable test */)
		if len(newShape) == 0 {
			return newVec[0]
		} else {
			return &Mat{newVec, newShape}
		}
	}
}

func asMat(a Val) *Mat {
	mat, ok := a.(*Mat)
	if !ok {
		sc := a.GetScalarOrNil()
		if sc == nil {
			return nil
		}
		return &Mat{M: []Val{sc}, S: []int{1}}
	}
	return mat
}

func cxcxJ(ca, cb complex128) complex128 {
	return ca + complex(0.0, 1.0)*cb
}
func cxcxRect(ca, cb complex128) complex128 {
	Must(imag(ca) == 0)
	Must(imag(cb) == 0)
	return cmplx.Rect(real(ca), real(cb))
}

func dyadicRho(c *Context, a Val, b Val, axis int) Val {
	spec := GetVectorOfScalarInts(a)
	outSize := MulReduce(spec)
	bm := asMat(b)

	if outSize > 0 && bm == nil {
		Log.Panicf("Cannot resize empty matrix to shape %v", spec)
	}

	if len(spec) == 0 {
		return bm.M[0]
	}

	var vec []Val
	vec, _ = recursiveFill(spec, bm.M, vec, 0)
	return &Mat{M: vec, S: spec}
}

func recursiveFill(shape []int, source []Val, vec []Val, i int) ([]Val, int) {
	modulus := len(source)
	head := shape[0]
	tail := shape[1:]
	if len(tail) > 0 {
		for j := 0; j < head; j++ {
			vec, i = recursiveFill(tail, source, vec, i)
		}
	} else {
		for j := 0; j < head; j++ {
			vec = append(vec, source[i])
			i = (i + 1) % modulus
		}
	}
	return vec, i
}

type FuncFloatFloatBool func(float64, float64) bool
type FuncCxCxBool func(complex128, complex128) bool
type FuncFloatFloatFloat func(float64, float64) float64
type FuncCxCxCx func(complex128, complex128) complex128

func WrapFloatDyadic(fn FuncFloatFloatFloat) DyadicFunc {
	return func(c *Context, a, b Val, axis int) Val {
		x := a.GetScalarFloat()
		y := b.GetScalarFloat()
		return &Num{complex(fn(x, y), 0)}
	}
}

func WrapCxDyadic(fn FuncCxCxCx) DyadicFunc {
	return func(c *Context, a, b Val, axis int) Val {
		x := a.GetScalarCx()
		y := b.GetScalarCx()
		return &Num{fn(x, y)}
	}
}

func WrapFloatBoolDyadic(fn FuncFloatFloatBool) DyadicFunc {
	return func(c *Context, a, b Val, axis int) Val {
		x := a.GetScalarFloat()
		y := b.GetScalarFloat()
		if fn(x, y) {
			return &Num{complex(1.0, 0)}
		} else {
			return &Num{complex(0.0, 0)}
		}
	}
}
func WrapCxBoolDyadic(fn FuncCxCxBool) DyadicFunc {
	return func(c *Context, a, b Val, axis int) Val {
		x := a.GetScalarCx()
		y := b.GetScalarCx()
		if fn(x, y) {
			return &Num{complex(1.0, 0)}
		} else {
			return &Num{complex(0.0, 0)}
		}
	}
}

func SameRank(a, b *Mat) bool {
	return len(a.S) == len(b.S)
}
func SameShape(a, b *Mat) bool {
	if !SameRank(a, b) {
		return false
	}
	for i := 0; i < len(a.S); i++ {
		if a.S[i] != b.S[i] {
			return false
		}
	}
	return true
}

func WrapMatMatDyadic(fn DyadicFunc) DyadicFunc {
	return func(c *Context, a, b Val, axis int) Val {
		switch x := a.(type) {
		case *Mat:
			switch y := b.(type) {
			case *Mat:
				// same shape or wrong.
				if SameShape(x, y) {

					n := len(x.M)
					vec := make([]Val, n)

					for i := 0; i < n; i++ {
						x1 := x.M[i].GetScalarOrNil()
						if x1 == nil {
							Log.Panicf("LHS not a scalar at matrix offset %d: %s", i, x1)
						}
						y1 := y.M[i].GetScalarOrNil()
						if y1 == nil {
							Log.Panicf("RHS not a scalar at matrix offset %d: %s", i, y1)
						}
						vec[i] = fn(c, x1, y1, axis)
					}

					return &Mat{M: vec, S: x.S}
				}
			}
			ys := b.GetScalarOrNil()
			if ys != nil {

				n := len(x.M)
				vec := make([]Val, n)
				for i := 0; i < n; i++ {
					x1 := x.M[i].GetScalarOrNil()
					if x1 == nil {
						Log.Panicf("LHS not a scalar at matrix offset %d: %s", i, x1)
					}
					vec[i] = fn(c, x1, ys, axis)
				}

				return &Mat{M: vec, S: x.S}
			}
		}

		//Log.Printf("ONE %s", a)
		xs := a.GetScalarOrNil()
		//Log.Printf("TWO %s", xs)
		if xs == nil {
			Log.Panicf("LHS neither matching matrix nor scalar: %s", a)
		}

		switch y := b.(type) {
		case *Mat:
			if xs != nil {

				n := len(y.M)
				vec := make([]Val, n)
				for i := 0; i < n; i++ {
					y1 := y.M[i].GetScalarOrNil()
					if y1 == nil {
						Log.Panicf("RHS not a scalar at matrix offset %d: %s", i, y1)
					}
					vec[i] = fn(c, xs, y1, axis)
				}

				return &Mat{M: vec, S: y.S}
			}
		}

		ys := b.GetScalarOrNil()
		if ys == nil {
			Log.Panicf("RHS neither matrix nor scalar: %s", b)
		}
		return fn(c, xs, ys, axis)
	}
}

func GetVectorOfScalarVals(a Val) []Val {
	var z []Val

	mat, ok := a.(*Mat)
	if !ok {
		// degenerate vector from scalar.
		y := a.GetScalarOrNil()
		if y == nil {
			Log.Panicf("GetVectorOfScalarVals: neither vector nor scalar: %v", a)
		}
		z = append(z, y)
	} else {
		for _, x := range mat.M {
			y := x.GetScalarOrNil()
			if y == nil {
				Log.Panicf("GetVectorOfScalarVals: item not scalar")
			}
			z = append(z, y)
		}
	}
	return z
}

func GetVectorOfScalarFloats(a Val) []float64 {
	var z []float64

	mat, ok := a.(*Mat)
	if !ok {
		// degenerate vector from scalar.
		z = append(z, a.GetScalarFloat())
	} else {
		// convert vector to float64s.
		for _, x := range mat.M {
			z = append(z, x.GetScalarFloat())
		}
	}
	return z
}

func GetVectorOfScalarInts(a Val) []int {
	var z []int

	mat, ok := a.(*Mat)
	if !ok {
		// degenerate vector from scalar.
		z = append(z, a.GetScalarInt())
	} else {
		// convert vector to ints.
		for _, x := range mat.M {
			z = append(z, x.GetScalarInt())
		}
	}
	return z
}

func dyadicRot(c *Context, a Val, b Val, axis int) Val {
	mat, bok := b.(*Mat)
	if !bok {
		Log.Panicf("Cannot rotate a non-matrix")
	}
	shape := mat.S     // in & out shape.
	rank := len(shape) // in & out rank.
	if rank == 0 {
		return b
	}
	axis = mod(axis, rank)
	revaxis := rank - axis

	// spec is the rearrangement specification.
	var spec []int
	var specShape []int
	amat, aok := a.(*Mat)
	if aok {
		if len(amat.S)+1 != len(mat.S) {
			Log.Panicf("rotate: LHS has shape %v; RHS has shape %v; axis is %d; shape of LHS should be 1 shorter than shape of RHS", amat.S, mat.S, axis)
		}
		spec = GetVectorOfScalarInts(a)
		j := 0
		for i, e := range mat.S {
			if i == axis {
				// skip chosen axis.
			} else {
				if amat.S[j] != e {
					Log.Panicf("rotate: LHS has shape %v; RHS has shape %v; axis is %d; dim %d of LHS should match dim %d of RHS", amat.S, mat.S, axis, j, i)
				}
				specShape = append(specShape, e)
				j++
			}
		}
	} else {
		r := a.GetScalarInt()
		for i, e := range mat.S {
			if i != axis {
				specShape = append(specShape, e)
			}
		}
		n := MulReduce(specShape)
		for i := 0; i < n; i++ {
			spec = append(spec, r)
		}
	}

	inVec := mat.M
	var outVec []Val

	var recurse func(shape, specShape []int, inOff, outOff int, spec []int, deferStart, deferStride, deferLen int)
	recurse = func(shape, specShape []int, inOff, outOff int, spec []int, deferStart, deferStride, deferLen int) {
		switch len(shape) {
		case 0:
			{
				// Log.Printf("[rot] in=%d out=%d spec=%d start=%d deferStride=%d deferLen=%d", inOff, outOff, spec, deferStart, deferStride, deferLen)
				r := mod(deferStart+spec[0], deferLen)
				x := inVec[inOff+r*deferStride]
				outVec = append(outVec, x)
			}
		case revaxis:
			{
				stride := MulReduce(shape[1:])
				for j := 0; j < shape[0]; j++ {
					recurse(shape[1:], specShape, inOff, outOff+j*stride, spec, j, stride, shape[0])
				}
			}
		default:
			{
				stride := MulReduce(shape[1:])
				specStride := MulReduce(specShape[1:])
				for j := 0; j < shape[0]; j++ {
					// Log.Printf("default j=%d (in=%d off=%d)  shape=%v   stride=%d spec=%v", j, inOff, outOff, shape, stride, spec)
					recurse(shape[1:], specShape[1:], inOff+j*stride, outOff+j*stride, spec[j*specStride:], deferStart, deferStride, deferLen)
				}
			}
		}
	}
	recurse(shape, specShape, 0, 0, spec, 0, 0, 0)
	return &Mat{M: outVec, S: shape}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func xor(a, b bool) bool {
	if a {
		return !b
	}
	return b
}
func dyadicTake(c *Context, a Val, b Val, axis int) Val {
	return dyadicTakeOrDrop(c, a, b, axis, false)
}
func dyadicDrop(c *Context, a Val, b Val, axis int) Val {
	return dyadicTakeOrDrop(c, a, b, axis, true)
}
func dyadicTakeOrDrop(c *Context, a Val, b Val, axis int, dropping bool) Val {
	if axis != -1 {
		Log.Panicf("Cannot specify axis for take or drop: %d", axis)
	}
	spec := GetVectorOfScalarInts(a)
	mat, ok := b.(*Mat)
	if !ok {
		Log.Panicf("Dyadic Take wants matrix on right, but got %#v", b)
	}
	inVec := mat.M
	inShape := mat.S
	if len(spec) != len(inShape) {
		Log.Panicf("Dyadic Take wants them to be the same, but len(LHS) == %d and len(shape(RHS)) == %d", len(spec), len(inShape))
	}

	// Figure out the outShape (how many to copy) and the inStart (where to start copying from).
	var prePad []int
	var postPad []int
	var outShape []int
	var inStart []int
	for i, sz := range inShape {
		pre, post := 0, 0 // padding
		k := abs(spec[i])
		if k > sz {
			if dropping {
				// TODO
				Log.Panicf("Dyadic Drop LHS[%d] abs too big, is %d; RHS shape is %v", i, spec[i], inShape)
			} else {
				if spec[i] > 0 {
					post = k - sz
				} else {
					pre = k - sz
				}
				k = sz
			}
		}
		if dropping {
			k = sz - k // k is how many to keep.
		}
		outShape = append(outShape, pre+k+post)
		if xor(dropping, spec[i] < 0) {
			inStart = append(inStart, sz-k)
		} else {
			inStart = append(inStart, 0)
		}
		prePad = append(prePad, pre)
		postPad = append(postPad, post)
	}
	outVec := make([]Val, MulReduce(outShape))

	Log.Printf("inStart %v", inStart)
	Log.Printf("inShape %v", inShape)
	Log.Printf("outShape %v", outShape)

	var recurse func(inStart []int, inShape []int, inOff int, outShape []int, outOff int, pre, post []int, zeroing bool)
	recurse = func(inStart []int, inShape []int, inOff int, outShape []int, outOff int, pre, post []int, zeroing bool) {
		if len(inStart) == 0 {
			Log.Printf("CP %d <= %d", outOff, inOff)
			if zeroing {
				outVec[outOff] = Zero
			} else {
				outVec[outOff] = inVec[inOff]
			}
			return
		}
		inStride := MulReduce(inShape[1:])
		outStride := MulReduce(outShape[1:])
		for i := 0; i < outShape[0]; i++ {
			nextInOff := inOff + (inStart[0]+i-pre[0])*inStride
			nextOutOff := outOff + i*outStride
			switch {
			case i < pre[0]:
				recurse(inStart[1:], inShape[1:], nextInOff, outShape[1:], nextOutOff, prePad[1:], postPad[1:], true)
			case outShape[0]-1-i < post[0]:
				recurse(inStart[1:], inShape[1:], nextInOff, outShape[1:], nextOutOff, prePad[1:], postPad[1:], true)
			default:
				recurse(inStart[1:], inShape[1:], nextInOff, outShape[1:], nextOutOff, prePad[1:], postPad[1:], zeroing)
			}
		}
	}
	recurse(inStart, inShape, 0, outShape, 0, prePad, postPad, false)
	return &Mat{M: outVec, S: outShape}
}

func dyadicExpand(c *Context, a Val, b Val, axis int) Val {
	return dyadicExpandOrCompress(c, a, b, axis, false, `\`)
}
func dyadicCompress(c *Context, a Val, b Val, axis int) Val {
	return dyadicExpandOrCompress(c, a, b, axis, true, `/`)
}
func dyadicExpandOrCompress(c *Context, a Val, b Val, axis int, compressing bool, name string) Val {
	mat, ok := b.(*Mat)
	if !ok {
		Log.Panicf("dyadic %s wants matrix on right, but got %#v", name, b)
	}
	inVec := mat.M
	inShape := mat.S
	origInRank := len(inShape)
	axis = mod(axis, origInRank)
	srcAxisShape := inShape[axis]

	spec := GetVectorOfScalarInts(a)
	// In the plan, 0 upwards mean copy over that source position.
	// So these special negative numbers can mean drop (for compress) & insert (for expand).
	const kDrop = -1
	const kInsert = -2
	var plan []int
	srcPos := 0 // Counts source positions, advances on 1's.
	destLen := 0
	for _, a := range spec {
		switch a {
		case 0:
			if compressing {
				plan = append(plan, kDrop)
				srcPos++
			} else {
				plan = append(plan, kInsert)
				destLen++
			}
		case 1:
			if srcPos == srcAxisShape {
				Log.Panicf("Dyadic %s axis is not wide enough: got LHS == %v; RHS shape is %v", name, spec, inShape)
			}
			plan = append(plan, srcPos)
			srcPos++
			destLen++
		default:
			Log.Panicf("dyadic %s has non-boolean element on LHS: %v", name, spec)
		}
	}

	var outShape []int
	for i, a := range inShape {
		if i == axis {
			outShape = append(outShape, destLen)
		} else {
			outShape = append(outShape, a)
		}
	}
	outVec := make([]Val, MulReduce(outShape))

	var recurse func(inShape []int, inOff int, outShape []int, outOff int)
	recurse = func(inShape []int, inOff int, outShape []int, outOff int) {
		if len(outShape) == 0 {
			if inOff == -1 {
				Log.Printf("ZERO %d", outOff)
				outVec[outOff] = &Num{0.0}
			} else {
				Log.Printf("CP %d <= %d", outOff, inOff)
				outVec[outOff] = inVec[inOff]
			}
			return
		}
		inStride := MulReduce(inShape[1:])
		outStride := MulReduce(outShape[1:])
		if len(inShape)+axis == origInRank {
			j := 0 // out index
			for _, p := range plan {
				switch p {
				case kDrop:
					break
				case kInsert:
					recurse(inShape[1:], -1, outShape[1:], outOff+j*outStride)
					j++
				default:
					recurse(inShape[1:], inOff+p*inStride, outShape[1:], outOff+j*outStride)
					j++
				}
			}
		} else {
			for i := 0; i < inShape[0]; i++ {
				if inOff == -1 {
					recurse(inShape[1:], -1, outShape[1:], outOff+i*outStride)
				} else {
					recurse(inShape[1:], inOff+i*inStride, outShape[1:], outOff+i*outStride)
				}
			}
		}
	}
	recurse(inShape, 0, outShape, 0)
	return &Mat{M: outVec, S: outShape}
}

func dyadicLaminate(c *Context, a Val, b Val, axis int) Val {
	ma, ok := a.(*Mat)
	if !ok {
		Log.Panicf("Dyadic `laminate` wants matrix on left, but got %#v", a)
	}

	mb, ok := b.(*Mat)
	if !ok {
		Log.Panicf("Dyadic `laminate` wants matrix on right, but got %#v", b)
	}

	aVec := ma.M
	aShape := ma.S
	aRank := len(aShape)
	bVec := mb.M
	bShape := mb.S
	bRank := len(bShape)

	if aRank != bRank {
		Log.Panicf("Dyadic `,` wants same shape, but left shape is %v and right shape is %v", aShape, bShape)
	}
	for i := 0; i < aRank; i++ {
		if aShape[i] != bShape[i] {
			Log.Panicf("Dyadic `,` wants same shape, but left shape is %v and right shape is %v", aShape, bShape)
		}
	}
	// axis is the newly created dimension, from 0 to aRank+1 (incl).
	axis = mod(axis, aRank+1)

	var outShape []int
	for i := 0; i < aRank+1; i++ {
		if i < axis {
			outShape = append(outShape, aShape[i])
		} else if i == axis {
			outShape = append(outShape, 2) // New dimension has length 2.
		} else {
			outShape = append(outShape, aShape[i-1])
		}
	}
	Log.Printf("Laminate: lhs %v rhs %v out %v", aShape, bShape, outShape)

	outVec := make([]Val, MulReduce(outShape))
	newDimStride := MulReduce(aShape[axis:])

	var recurse func(inShape []int, inOff int, outShape []int, outOff int)
	recurse = func(inShape []int, inOff int, outShape []int, outOff int) {
		if len(inShape) == 0 {
			outVec[outOff] = aVec[inOff]
			outVec[outOff+newDimStride] = bVec[inOff]
			return
		}

		if len(inShape)+axis == aRank {
			outShape = outShape[1:] // Skip the outShape when on the new axis.
		}

		inStride := MulReduce(inShape[1:])
		outStride := MulReduce(outShape[1:])
		for i := 0; i < inShape[0]; i++ {
			recurse(inShape[1:], inOff+i*inStride, outShape[1:], outOff+i*outStride)
		}
	}

	recurse(aShape, 0, outShape, 0)

	return &Mat{M: outVec, S: outShape}
}

func dyadicCatenate(c *Context, a Val, b Val, axis int) Val {
	ma, aok := a.(*Mat)
	mb, bok := b.(*Mat)
	if !aok && !bok {
		// Concat two scalars into a pair.
		return &Mat{[]Val{a, b}, []int{2}}
	}

	if !aok {
		n := len(mb.S)
		newShape := make([]int, n)
		copy(newShape, mb.S)
		axis = mod(axis, n)
		newShape[axis] = 1
		ma = &Mat{RepeatVal(a, MulReduce(newShape)), newShape}
	}

	if !bok {
		n := len(ma.S)
		newShape := make([]int, n)
		copy(newShape, ma.S)
		axis = mod(axis, n)
		newShape[axis] = 1
		mb = &Mat{RepeatVal(b, MulReduce(newShape)), newShape}
	}

	aVec := ma.M
	aShape := ma.S
	aRank := len(aShape)
	bVec := mb.M
	bShape := mb.S
	bRank := len(bShape)

	if aRank != bRank {
		Log.Panicf("Dyadic `,` wants same rank, but left shape is %v and right shape is %v", aShape, bShape)
	}
	axis = mod(axis, aRank)

	var outShape []int
	for i := 0; i < aRank; i++ {
		if i == axis {
			outShape = append(outShape, aShape[i]+bShape[i])
		} else {
			if aShape[i] != bShape[i] {
				Log.Panicf("Dyadic `,` wants same shape except for axis dimension %d, but left shape is %v and right shape is %v", axis, aShape, bShape)
			}
			outShape = append(outShape, aShape[i])
		}
	}
	Log.Printf("Concatenate: lhs %v rhs %v out %v", aShape, bShape, outShape)

	outVec := make([]Val, MulReduce(outShape))
	var inVec []Val

	var recurse func(inShape []int, inOff int, outShape []int, outOff int)
	recurse = func(inShape []int, inOff int, outShape []int, outOff int) {
		if len(inShape) == 0 {
			outVec[outOff] = inVec[inOff]
			return
		}

		inStride := MulReduce(inShape[1:])
		outStride := MulReduce(outShape[1:])
		for i := 0; i < inShape[0]; i++ {
			recurse(inShape[1:], inOff+i*inStride, outShape[1:], outOff+i*outStride)
		}
	}

	inVec = aVec
	recurse(aShape, 0, outShape, 0)

	inVec = bVec
	recurse(bShape, 0, outShape, aShape[axis]*MulReduce(outShape[axis+1:]))

	return &Mat{M: outVec, S: outShape}
}

func dyadicTranspose(c *Context, a Val, b Val, axis int) Val {
	mat, ok := b.(*Mat)
	if !ok {
		Log.Panicf("Dyadic `transpose` wants matrix on right, but got %#v", b)
	}

	inVec := mat.M
	inShape := mat.S
	inRank := len(inShape)

	spec := GetVectorOfScalarInts(a)
	if len(spec) != inRank {
		Log.Panicf("Dyadic `transpose` wants length of lhs %d to match rank of rhs %d", len(spec), inRank)
	}
	for i := range spec {
		spec[i] = mod(spec[i], inRank)
	}

	outRank := 0
	for _, e := range spec {
		if e < 0 || e >= len(spec) {
			Log.Panicf("Dyadic `transpose` finds %d on lhs, not a valid dimensio in lhs %v", e, spec)
		}
		if e+1 > outRank {
			outRank = e + 1
		}
	}
	outShape := make([]int, outRank)
	stride := make([]int, outRank)
	for i, e := range spec {
		//Log.Printf("i=%d e=%d, outShape<<%v", i, e, outShape)
		outShape[e] = inShape[i]
		// If this happens more than once, it is a diagonal:
		stride[e] += MulReduce(inShape[i+1:])
		//Log.Printf("outShape>>%v; stride>>%v", outShape, stride)
	}

	var outVec []Val

	var recurse func(outShape []int, inOff int, stride []int)
	recurse = func(outShape []int, inOff int, stride []int) {
		if len(outShape) == 0 {
			outVec = append(outVec, inVec[inOff])
			//Log.Printf("inOff %d outVec %v", inOff, outVec)
			return
		}

		for i := 0; i < outShape[0]; i++ {
			//Log.Printf("outShape=%v inOff=%d stride=%v i=%d", outShape, inOff, stride, i)
			recurse(outShape[1:], inOff+i*stride[0], stride[1:])
		}
	}
	recurse(outShape, 0, stride)

	return &Mat{M: outVec, S: outShape}
}

/*
type ValSlice []*Val

func (p ValSlice) Len() int {
	return len(p)
}

func (p ValSlice) Less(i, j int) bool {
	na, ok := a.(*Num)
	if !ok {
		Log.Panicf("Expected a number, but got %#v", a)
	}

	nb, ok := b.(*Num)
	if !ok {
		Log.Panicf("Expected a number, but got %#v", b)
	}

	return na.F < nb.F
}

func (p ValSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
*/

func dyadicMember(c *Context, a Val, b Val, axis int) Val {
	// TODO: allow scalar, particularly on LHS.
	// TODO: value.go should compare any Val, not just *Num.
	mata, ok := a.(*Mat)
	if !ok {
		Log.Panicf("Dyadic `member` wants matrix on left, but got %#v", a)
	}

	matb, ok := b.(*Mat)
	if !ok {
		Log.Panicf("Dyadic `member` wants matrix on right, but got %#v", b)
	}

	aVec := mata.M
	aShape := mata.S
	outShape := aShape
	outVec := make([]Val, len(aVec))

	bVec := matb.M
	floats := make(sort.Float64Slice, len(bVec))
	for i, e := range bVec {
		num, ok := e.(*Num)
		if !ok {
			Log.Panicf("Dyadic `member` RHS element @%d not a number: %v", i, e)
		}
		floats[i] = num.GetScalarFloat()
	}
	floats.Sort()

	for i, e := range aVec {
		num, ok := e.(*Num)
		if !ok {
			Log.Panicf("Dyadic `member` RHS element @%d not a number: %v", i, e)
		}
		j := sort.SearchFloat64s(floats, num.GetScalarFloat())

		found := (j < len(floats) && floats[j] == num.GetScalarFloat())
		outVec[i] = &Num{complex(boolf(found), 0)}
	}

	return &Mat{outVec, outShape}
}

func RepeatVal(a Val, n int) []Val {
	z := make([]Val, n)
	for i, _ := range z {
		z[i] = a
	}
	return z
}
