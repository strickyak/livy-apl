package livy

import (
	"log"
	"math"
)

type DyadicFunc func(c *Context, a Val, b Val, axis int) Val

var StandardDyadics = map[string]DyadicFunc{
	"rho":  dyadicRho,
	"rot":  dyadicRot,
	"take": dyadicTake,
	"drop": dyadicDrop,

	"==": WrapMatMatDyadic(WrapFloatBoolDyadic(
		func(a, b float64) bool { return a == b })),
	"!=": WrapMatMatDyadic(WrapFloatBoolDyadic(
		func(a, b float64) bool { return a != b })),
	"<": WrapMatMatDyadic(WrapFloatBoolDyadic(
		func(a, b float64) bool { return a < b })),
	">": WrapMatMatDyadic(WrapFloatBoolDyadic(
		func(a, b float64) bool { return a > b })),
	"<=": WrapMatMatDyadic(WrapFloatBoolDyadic(
		func(a, b float64) bool { return a <= b })),
	">=": WrapMatMatDyadic(WrapFloatBoolDyadic(
		func(a, b float64) bool { return a >= b })),

	"+": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return a + b })),
	"-": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return a - b })),
	"*": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return a * b })),
	"div": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return a / b })),
	"**": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return math.Pow(a, b) })),
	"remainder": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return math.Remainder(a, b) })),
	"mod": WrapMatMatDyadic(WrapFloatDyadic(
		func(a, b float64) float64 { return math.Mod(a, b) })),
}

var Zero = &Num{0.0}
var One = &Num{1.0}
var ScanAndReduceWithIdentity = map[string]Val{
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

func init() {
	for name, identity := range ScanAndReduceWithIdentity {
		fn, ok := StandardDyadics[name]
		if ok {
			reduceName := name + `/`
			StandardMonadics[reduceName] = MkReduceOrScanOp(reduceName, fn, identity, false)

			scanName := name + `\`
			StandardMonadics[scanName] = MkReduceOrScanOp(scanName, fn, identity, true)

			outerProductName := "@." + name
			StandardDyadics[outerProductName] = MkOuterProduct(outerProductName, fn)
		}
	}
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

func MkReduceOrScanOp(name string, fn DyadicFunc, identity Val, toScan bool) MonadicFunc {
	verb := "reduce"
	if toScan {
		verb = "scan"
	}
	return func(c *Context, a Val, axis int) Val {
		mat, ok := a.(*Mat)
		if !ok {
			log.Panicf("Cannot %s %s on non-matrix: %s", name, verb, a)
		}
		oldRank := len(mat.S)
		oldShape := mat.S
		log.Printf("oldShape %v oldRank %v for matrix %v", oldShape, oldRank, mat)
		if oldRank == 0 {
			log.Panicf("Cannot %s %s on scalar: %s", name, verb, mat)
		}
		if axis < 0 {
			axis += oldRank
		}
		if axis < 0 || axis > oldRank-1 {
			log.Panicf("Reduce axis [%d] is bad for %s %s of rank %d", axis, name, verb, oldRank)
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

		newVecLen := mulReduce(newShape)
		newVec := make([]Val, newVecLen)
		oldVec := mat.M

		reduceStride, reduceLen := mulReduce(oldShape[axis+1:]), oldShape[axis]
		log.Printf("Reduce Stride = %d", reduceStride)
		revAxis := oldRank - axis

		var reduce func(oldShape []int, oldOffset int, newShape []int, newOffset int, reduceOffset int)
		reduce = func(oldShape []int, oldOffset int, newShape []int, newOffset int, reduceOffset int) {
			rank := len(oldShape)
			log.Printf("[[%d]] ;; old %d @ %v ;; new %d @ %v ;; {ro=%d,rs=%d,revAxis=%d}", rank, oldOffset, oldShape, newOffset, newShape, reduceOffset, reduceStride, revAxis)
			if rank == 0 {
				var reduction Val

				if reduceLen == 0 {
					reduction = identity
				} else {
					// j is 0:
					reduction = oldVec[oldOffset]
					if toScan {
						newVec[newOffset] = reduction
						log.Printf("Scan...0  newVec: %v [[%d; %s]]", newVec, newOffset, reduction)
					}
					// other j's:
					for j := 1; j < reduceLen; j++ {
						log.Printf("...... %d [[%d; %s]]", j, newOffset, reduction)
						reduction = fn(c, reduction, oldVec[oldOffset+j*reduceStride], DefaultAxis)
						if toScan {
							newVec[newOffset+j*reduceStride] = reduction
							log.Printf("Scan...%d  newVec: %v [[%d; %s]]", j, newVec, newOffset+j*reduceStride, reduction)
						}
					}
				}

				if !toScan {
					newVec[newOffset] = reduction
					log.Printf("Reduction...  newVec: %v [[%d; %s]]", newVec, newOffset, reduction)
				}
			} else if len(oldShape) == revAxis {
				// This is the old dimension we reduce.
				// It exists in the oldShape but not in the newShape, if reducing.
				// We do not iterate i here -- that will happen above when rank==0.
				if reduceStride != mulReduce(oldShape[1:]) {
					panic(0)
				}
				if toScan {
					newShape = newShape[1:]
				}
				log.Printf("111 reduceOffset = %d ;; old %d @ %v ;; new %d @ %v -->", oldOffset, reduceOffset, oldShape, newOffset, newShape)
				reduce(oldShape[1:], oldOffset, newShape, newOffset, oldOffset)
			} else {
				for i := 0; i < oldShape[0]; i++ {
					oldStride := mulReduce(oldShape[1:])
					newStride := mulReduce(newShape[1:])
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

func dyadicRho(c *Context, a Val, b Val, axis int) Val {
	var shape []int

	am := asMat(a)
	bm := asMat(b)
	if am == nil || bm == nil {
		panic("BAD")
	}

	for _, e := range am.M {
		ei := e.GetScalarInt()
		shape = append(shape, ei)
	}

	if len(shape) == 0 {
		return &Mat{nil, nil}
	}

	var vec []Val
	vec, _ = recursiveFill(shape, bm.M, vec, 0)
	return &Mat{M: vec, S: shape}
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
type FuncFloatFloatFloat func(float64, float64) float64

func WrapFloatDyadic(fn FuncFloatFloatFloat) DyadicFunc {
	return func(c *Context, a, b Val, axis int) Val {
		x := a.GetScalarFloat()
		y := b.GetScalarFloat()
		return &Num{fn(x, y)}
	}
}

func WrapFloatBoolDyadic(fn FuncFloatFloatBool) DyadicFunc {
	return func(c *Context, a, b Val, axis int) Val {
		x := a.GetScalarFloat()
		y := b.GetScalarFloat()
		if fn(x, y) {
			return &Num{1.0}
		} else {
			return &Num{0.0}
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
							log.Panicf("LHS not a scalar at matrix offset %d: %s", i, x1)
						}
						y1 := y.M[i].GetScalarOrNil()
						if y1 == nil {
							log.Panicf("RHS not a scalar at matrix offset %d: %s", i, y1)
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
						log.Panicf("LHS not a scalar at matrix offset %d: %s", i, x1)
					}
					vec[i] = fn(c, x1, ys, axis)
				}

				return &Mat{M: vec, S: x.S}
			}
		}

		//log.Printf("ONE %s", a)
		xs := a.GetScalarOrNil()
		//log.Printf("TWO %s", xs)
		if xs == nil {
			log.Panicf("LHS neither matching matrix nor scalar: %s", a)
		}

		switch y := b.(type) {
		case *Mat:
			if xs != nil {

				n := len(y.M)
				vec := make([]Val, n)
				for i := 0; i < n; i++ {
					y1 := y.M[i].GetScalarOrNil()
					if y1 == nil {
						log.Panicf("RHS not a scalar at matrix offset %d: %s", i, y1)
					}
					vec[i] = fn(c, xs, y1, axis)
				}

				return &Mat{M: vec, S: y.S}
			}
		}

		ys := b.GetScalarOrNil()
		if ys == nil {
			log.Panicf("RHS neither matrix nor scalar: %s", b)
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
			log.Panicf("GetVectorOfScalarVals: neither vector nor scalar: %v", a)
		}
		z = append(z, y)
	} else {
		for _, x := range mat.M {
			y := x.GetScalarOrNil()
			if y == nil {
				log.Panicf("GetVectorOfScalarVals: item not scalar")
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
	// spec is the rearrangement specification.
	var spec []int
	spec = GetVectorOfScalarInts(a)
	/*
		specMat, ok := a.(*Mat)
		if !ok {
			// degenerate vector from scalar.
			spec = append(spec, a.GetScalarInt())
		} else {
			// convert vector to ints.
			for _, x := range specMat.M {
				println(x)
				spec = append(spec, x.GetScalarInt())
			}
		}
	*/
	mat, ok := b.(*Mat)
	if !ok {
		// scalar is like 1x1, whose rot or flip is itself.
		return b
	}
	inVec := mat.M
	inShape := mat.S
	n := len(inShape)
	if n < 1 {
		// rot or flip on Emptiness yields Emptiness.
		return b
	}
	axis = ((axis % n) + n) % n
	revaxis := n - axis

	// Result shape
	var outShape []int
	for i, sz := range inShape {
		if i == axis {
			outShape = append(outShape, len(spec))
		} else {
			outShape = append(outShape, sz)
		}
	}
	outSize := mulReduce(outShape)
	if outSize < 1 {
		return &Mat{M: nil, S: outShape}
	}
	outVec := make([]Val, outSize)

	var recurse func(inShape []int, inOff int, outShape []int, outOff int)
	recurse = func(inShape []int, inOff int, outShape []int, outOff int) {
		switch len(outShape) {
		case 0:
			// log.Printf("Assign out [%d] <- in [%d]", outOff, inOff)
			outVec[outOff] = inVec[inOff]
		case revaxis:
			inStride := mulReduce(inShape[1:])
			outStride := mulReduce(outShape[1:])
			for o := 0; o < outShape[0]; o++ {
				i := spec[o]
				i = ((i % inShape[0]) + inShape[0]) % inShape[0]
				recurse(inShape[1:], inOff+i*inStride,
					outShape[1:], outOff+o*outStride)
			}
		default:
			inStride := mulReduce(inShape[1:])
			outStride := mulReduce(outShape[1:])
			for o := 0; o < outShape[0]; o++ {
				i := o
				recurse(inShape[1:], inOff+i*inStride,
					outShape[1:], outOff+o*outStride)
			}
		}
	}
	recurse(inShape, 0, outShape, 0)
	return &Mat{M: outVec, S: outShape}
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
func dyadicTakeOrDrop(c *Context, a Val, b Val, axis int, doBeDropping bool) Val {
	if axis != -1 {
		log.Panicf("Cannot specify axis for take or drop: %d", axis)
	}
	spec := GetVectorOfScalarInts(a)
	mat, ok := b.(*Mat)
	if !ok {
		log.Panicf("Dyadic Take wants matrix on right, but got %#v", b)
	}
	inVec := mat.M
	inShape := mat.S
	if len(spec) != len(inShape) {
		log.Panicf("Dyadic Take wants them to be the same, but len(LHS) == %d and len(shape(RHS)) == %d", len(spec), len(inShape))
	}

	// Figure out the outShape (how many to copy) and the inStart (where to start copying from).
	var outShape []int
	var inStart []int
	for i, sz := range inShape {
		k := abs(spec[i])
		if k > sz {
			log.Panicf("Dyadic Take LHS[%d] abs too big, is %d; RHS shape is %v", i, spec[i], i, inShape)
		}
		if doBeDropping {
			k = sz - k // k is how many to keep.
		}
		outShape = append(outShape, k)
		if xor(doBeDropping, spec[i] < 0) {
			inStart = append(inStart, sz-k)
		} else {
			inStart = append(inStart, 0)
		}
	}
	outVec := make([]Val, mulReduce(outShape))

	log.Printf("inStart %v", inStart)
	log.Printf("inShape %v", inShape)
	log.Printf("outShape %v", outShape)

	var recurse func(inStart []int, inShape []int, inOff int, outShape []int, outOff int)
	recurse = func(inStart []int, inShape []int, inOff int, outShape []int, outOff int) {
		if len(inStart) == 0 {
			log.Printf("CP %d <= %d", outOff, inOff)
			outVec[outOff] = inVec[inOff]
			return
		}
		inStride := mulReduce(inShape[1:])
		outStride := mulReduce(outShape[1:])
		for i := 0; i < outShape[0]; i++ {
			recurse(inStart[1:], inShape[1:], inOff+(inStart[0]+i)*inStride, outShape[1:], outOff+i*outStride)
		}
	}
	recurse(inStart, inShape, 0, outShape, 0)
	return &Mat{M: outVec, S: outShape}
}
