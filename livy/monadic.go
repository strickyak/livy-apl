package livy

import (
	"log"
	"math"
)

type MonadicFunc func(*Context, Val) Val

var StandardMonadics = map[string]MonadicFunc{
	"asin": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Asin(b)
	})),
	"acos": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Acos(b)
	})),
	"atan": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Atan(b)
	})),
	"sin": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Sin(b)
	})),
	"cos": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Cos(b)
	})),
	"tan": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Tan(b)
	})),
	"asinh": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Asinh(b)
	})),
	"acosh": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Acosh(b)
	})),
	"atanh": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Atanh(b)
	})),
	"sinh": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Sinh(b)
	})),
	"cosh": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Cosh(b)
	})),
	"tanh": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Tanh(b)
	})),
	"exp": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Exp(b)
	})),
	"ln": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Log(b)
	})),
	"log10": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Log10(b)
	})),
	"ceil": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Ceil(b)
	})),
	"floor": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Floor(b)
	})),
	"cbrt": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Cbrt(b)
	})),
	"sqrt": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Sqrt(b)
	})),
	"double": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return b + b
	})),
	"square": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return b * b
	})),
	"iota": iotaMonadic,
	"rho":  rhoMonadic,
}

type FF func(float64) float64
type VV func(Val) Val

func WrapFloatMonadic(fn FF) VV {
	return func(b Val) Val {
		y := b.GetScalarFloat()
		return &Num{fn(y)}
	}
}

func doubleMonadic(c *Context, b Val) Val {
	switch y := b.(type) {
	case *Num:
		return &Num{2 * y.F}
	}
	log.Panicf("Wrong type for monadic `double`: %T %q", b, b)
	return nil
}

func iotaMonadic(c *Context, b Val) Val {
	n := b.GetScalarInt()
	vec := make([]Val, n)
	for i := 0; i < n; i++ {
		vec[i] = &Num{float64(i)}
	}
	return &Mat{
		M: vec,
		S: []int{n},
	}
}
func rhoMonadic(c *Context, b Val) Val {
	switch y := b.(type) {
	case *Mat:
		n := len(y.S)
		vec := make([]Val, n)
		for i := 0; i < n; i++ {
			vec[i] = &Num{float64(y.S[i])}
		}
		return &Mat{
			M: vec,
			S: []int{n},
		}
	default:
		return &Mat{
			M: nil,
			S: nil,
		}
	}
}

func WrapMatMonadic(fn VV) MonadicFunc {
	return func(c *Context, b Val) Val {
		switch y := b.(type) {
		case *Mat:
			n := len(y.M)
			vec := make([]Val, n)

			for i := 0; i < n; i++ {
				y1 := y.M[i].GetScalarOrNil()
				if y1 == nil {
					log.Panicf("arg not a scalar at matrix offset %d: %s", i, y1)
				}
				vec[i] = fn(y1)
			}

			return &Mat{M: vec, S: y.S}

		}

		ys := b.GetScalarOrNil()
		if ys == nil {
			log.Panicf("arg not scalar or matrix")
		}
		return fn(ys)
	}
}
