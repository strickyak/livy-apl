package livy

import (
	"log"
	"math"
)

type DyadicFunc func(*Context, Val, Val) Val

var StandardDyadics = map[string]DyadicFunc{
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

type FFF func(float64, float64) float64
type VVV func(Val, Val) Val

func WrapFloatDyadic(fn FFF) VVV {
	return func(a, b Val) Val {
		x := a.GetScalarFloat()
		y := b.GetScalarFloat()
		return &Num{fn(x, y)}
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

func WrapMatMatDyadic(fn VVV) DyadicFunc {
	return func(c *Context, a, b Val) Val {
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
						vec[i] = fn(x1, y1)
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
					vec[i] = fn(x1, ys)
				}

				return &Mat{M: vec, S: x.S}
			}
		}

		log.Printf("ONE %s", a)
		xs := a.GetScalarOrNil()
		log.Printf("TWO %s", xs)
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
					vec[i] = fn(xs, y1)
				}

				return &Mat{M: vec, S: y.S}
			}
		}

		ys := b.GetScalarOrNil()
		if ys == nil {
			log.Panicf("RHS neither matrix nor scalar: %s", b)
		}
		return fn(xs, ys)
	}
}
