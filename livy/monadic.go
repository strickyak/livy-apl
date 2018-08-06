package livy

import (
	"log"
)

type MonadicFunc func(*Context, Val) Val

var StandardMonadics = map[string]MonadicFunc{
	"double": double,
}

func double(c *Context, b Val) Val {
	switch y := b.(type) {
	case *Num:
		return &Num{2 * y.F}
	}
	log.Panicf("Wrong type for monadic `double`: %T %q", b, b)
	return nil
}
