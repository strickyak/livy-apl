package livy

import (
	"log"
)

type DyadicFunc func(*Context, Val, Val) Val

var StandardDyadics = map[string]DyadicFunc{
	"+": add,
}

func add(c *Context, a, b Val) Val {
	switch x := a.(type) {
	case *Num:
		switch y := b.(type) {
		case *Num:
			return &Num{x.F + y.F}
		}
	}
	log.Panicf("Wrong types for `+`: %T, %T: %q, %q, ", a, b, a, b)
	return nil
}
