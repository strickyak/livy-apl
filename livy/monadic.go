package livy

import (
	"log"
)

type MonadicFunc func(*Context, Val) Val

var Monadics = map[string]MonadicFunc{
	"double": Double,
}

func Double(c *Context, b Val) Val {
	switch x := b.(type) {
	case *Num:
		return &Num{2 * x.F}
	default:
		log.Panicf("Wrong type for Double: %T %q", b, b)
	}
	return nil
}
