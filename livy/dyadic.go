package livy

import (
	"log"
)

type DyadicFunc func(*Context, Val, Val) Val

var StandardDyadics = map[string]DyadicFunc{
	"+": addDyadic,
}

func addDyadic(c *Context, a, b Val) Val {
	switch x := a.(type) {
	case *Mat:
		switch y := b.(type) {
		case *Mat:
			n := len(x.M)
			vec := make([]Val, n)

			if len(x.S) != len(y.S) {
				goto TryScalarY
			}
			for i := 0; i < len(x.S); i++ {
				if x.S[i] != y.S[i] {
					goto TryScalarY
				}
			}
			for i := 0; i < n; i++ {
				z := x.M[i].GetScalarFloat() + y.M[i].GetScalarFloat()
				vec[i] = &Num{z}
			}

			return &Mat{
				M: vec,
				S: x.S,
			}
		TryScalarY:
			panic(0)
		}

	case *Num:
	       switch y := b.(type) {
               case *Num:
                       return &Num{x.F + y.F}
	      }

	}
	log.Panicf("Wrong args for `+`: %T, %T: %q, %q, ", a, b, a, b)
	return nil
}
