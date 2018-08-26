package livy

import (
	"log"
	"math"
)

type MonadicFunc func(c *Context, b Val, dim int) Val

var StandardMonadics = map[string]MonadicFunc{
	"rot":   rotMonadic,
	"iota":  iotaMonadic,
	"iota1": iota1Monadic,
	"rho":   rhoMonadic,
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
}

type funcFloatFloat func(b float64) float64

func WrapFloatMonadic(fn funcFloatFloat) MonadicFunc {
	return func(c *Context, b Val, axis int) Val {
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

func iotaMonadic(c *Context, b Val, axis int) Val {
	return iotaK(c, b, 0)
}
func iota1Monadic(c *Context, b Val, axis int) Val {
	return iotaK(c, b, 1)
}
func iotaK(c *Context, b Val, k int) Val {
	n := b.GetScalarInt()
	vec := make([]Val, n)
	for i := 0; i < n; i++ {
		vec[i] = &Num{float64(i + k)}
	}
	return &Mat{
		M: vec,
		S: []int{n},
	}
}
func rhoMonadic(c *Context, b Val, axis int) Val {
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

func WrapMatMonadic(fn MonadicFunc) MonadicFunc {
	return func(c *Context, b Val, axis int) Val {
		switch y := b.(type) {
		case *Mat:
			n := len(y.M)
			vec := make([]Val, n)

			for i := 0; i < n; i++ {
				y1 := y.M[i].GetScalarOrNil()
				if y1 == nil {
					log.Panicf("arg not a scalar at matrix offset %d: %s", i, y1)
				}
				vec[i] = fn(c, y1, axis)
			}

			return &Mat{M: vec, S: y.S}

		}

		ys := b.GetScalarOrNil()
		if ys == nil {
			log.Panicf("arg not scalar or matrix")
		}
		return fn(c, ys, axis)
	}
}

func rotMonadic(c *Context, b Val, axis int) Val {
	mat, ok := b.(*Mat)
	if !ok {
		// scalar is like 1x1, whose rot or flip is itself.
		return b
	}

	shape := mat.S
	n := len(shape)
	if n < 1 {
		// rot or flip on Emptiness yields Emptiness.
		return b
	}
	axis = (((axis % n) + n) % n)

	axisLen := shape[axis]
	var reversed []Val
	for i := axisLen - 1; i >= 0; i-- {
		reversed = append(reversed, &Num{float64(i)})
	}
	return dyadicRot(c, &Mat{reversed, []int{axisLen}}, b, axis)
}
