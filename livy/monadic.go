package livy

import (
	"log"
)

type MonadicFunc func(*Context, Val) Val

var StandardMonadics = map[string]MonadicFunc{
	"double": doubleMonadic,
	"iota": iotaMonadic,
	"rho": rhoMonadic,
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
