package livy

import (
	"log"
	"math"
)

type DyadicFunc func(c *Context, a Val, b Val, dim int) Val

var StandardDyadics = map[string]DyadicFunc{
	"rho": dyadicRho,

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
	"/": WrapMatMatDyadic(WrapFloatDyadic(
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
var ReduceWithIdentity = map[string]Val{
	"+":   Zero,
	"-":   Zero,
	"or":  Zero,
	"!=":  Zero,
	"<":   Zero,
	">":   Zero,
	"*":   One,
	"/":   One,
	"**":  One,
	"and": One,
	"==":  One,
	">=":  One,
	"<=":  One,
}

func init() {
	for name, identity := range ReduceWithIdentity {
		fn, ok := StandardDyadics[name]
		if ok {
			scanName := name + "/"
			StandardMonadics[scanName] = mkReduceOp(scanName, fn, identity)
		}
	}
}

func mkReduceOp(name string, fn DyadicFunc, identity Val) MonadicFunc {
	return func(c *Context, a Val, dim int) Val {
		mat, ok := a.(*Mat)
		if !ok {
			log.Panicf("Cannot %s/ reduce on non-matrix: %s", name, a)
		}
		oldRank := len(mat.S)
		oldShape := mat.S
		if oldRank == 0 {
			log.Panicf("Cannot %s/ reduce on scalar: %s", name, mat)
		}
		if dim < 0 {
			dim += oldRank
		}
		if dim < 0 || dim > oldRank-1 {
			log.Panicf("Reduce dimension [%d] is bad for %s/ reduce of rank %d", dim, name, oldRank)
		}

		var newShape []int
		for i := 0; i < oldRank; i++ {
			if i == dim {
				// Skip
			} else {
				newShape = append(newShape, oldShape[i])
			}
		}

		newVecLen := mulReduce(newShape)
		newVec := make([]Val, newVecLen)
		oldVec := mat.M

		// reduceTarget := mulReduce(oldShape[:dim])
		reduceStride, reduceLen := mulReduce(oldShape[dim+1:]), oldShape[dim]
		log.Printf("Reduce Stride = %d", reduceStride)
		revdim := oldRank - dim

		var reduce func(oldShape []int, oldOffset int, newShape []int, newOffset int, reduceOffset int)
		reduce = func(oldShape []int, oldOffset int, newShape []int, newOffset int, reduceOffset int) {
			rank := len(oldShape)
			log.Printf("[[%d]] ;; old %d @ %v ;; new %d @ %v ;; {ro=%d,rs=%d,revdim=%d}", rank, oldOffset, oldShape, newOffset, newShape, reduceOffset, reduceStride, revdim)
			if rank == 0 {
				var reduction Val
				if reduceLen == 0 {
					reduction = identity
				} else if reduceLen == 1 {
					reduction = oldVec[oldOffset]
				} else if reduceLen > 0 {
					reduction = oldVec[oldOffset]
					for j := 1; j < reduceLen; j++ {
						log.Printf("...... %d [[%d; %s]]", j, newOffset, reduction)
						reduction = fn(c, reduction, oldVec[oldOffset+j*reduceStride], DefaultDim)
					}
				}
				newVec[newOffset] = reduction
				log.Printf("...... newVec: %v [[%d; %s]]", newVec, newOffset, reduction)
			} else if len(oldShape) == revdim {
				// This is the old dimension we reduce.
				// It exists in the oldShape but not in the newShape.
				// We do not iterate i here -- that will happen above when rank==0.
				if reduceStride != mulReduce(oldShape[1:]) {
					panic(0)
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

func dyadicRho(c *Context, a Val, b Val, dim int) Val {
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
	return func(c *Context, a, b Val, dim int) Val {
		x := a.GetScalarFloat()
		y := b.GetScalarFloat()
		return &Num{fn(x, y)}
	}
}

func WrapFloatBoolDyadic(fn FuncFloatFloatBool) DyadicFunc {
	return func(c *Context, a, b Val, dim int) Val {
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
	return func(c *Context, a, b Val, dim int) Val {
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
						vec[i] = fn(c, x1, y1, dim)
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
					vec[i] = fn(c, x1, ys, dim)
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
					vec[i] = fn(c, xs, y1, dim)
				}

				return &Mat{M: vec, S: y.S}
			}
		}

		ys := b.GetScalarOrNil()
		if ys == nil {
			log.Panicf("RHS neither matrix nor scalar: %s", b)
		}
		return fn(c, xs, ys, dim)
	}
}
