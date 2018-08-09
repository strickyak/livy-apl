package livy

import (
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
)

var _ = debug.PrintStack

type Expression interface {
	Eval(*Context) Val
	String() string
}

type Variable struct {
	S string
}

type Number struct {
	F float64
}

type Monad struct {
	Op string
	B  Expression
}

type Dyad struct {
	A  Expression
	Op string
	B  Expression
}

type List struct {
	L []Expression
}

func (o Variable) Eval(c *Context) Val {
	z, ok := c.Globals[o.S]
	if !ok {
		log.Panicf("No such variable %q in Globals %#v", o.S, c.Globals)
	}
	return z
}
func (o Number) Eval(c *Context) Val {
	return &Num{o.F}
}
func (o Monad) Eval(c *Context) Val {
	fn, ok := c.Monadics[o.Op]
	if !ok {
		log.Panicf("No such monadaic operator %q", o.Op)
	}
	b := o.B.Eval(c)
	log.Printf("Monad:Eval %s %s -> ?", o.Op, b)
	z := fn(c, b)
	log.Printf("Monad:Eval %s %s -> %s", o.Op, b, z)
	return z
}
func (o Dyad) Assign(c *Context) Val {
	avar, ok := o.A.(*Variable)
	if !ok {
		log.Panicf("cannot assign to %s", o.A)
	}
	b := o.B.Eval(c)
	c.Globals[avar.S] = b
	log.Printf("Assigning %s = %s", avar.S, b)
	return b
}
func (o Dyad) Eval(c *Context) Val {
	if o.Op == "=" {
		return o.Assign(c)
	}
	fn, ok := c.Dyadics[o.Op]
	if !ok {
		log.Panicf("No such dyadaic operator %q", o.Op)
	}
	a := o.A.Eval(c)
	b := o.B.Eval(c)
	log.Printf("Dyad:Eval %s %s %s -> ?", a, o.Op, b)
	z := fn(c, a, b)
	log.Printf("Dyad:Eval %s %s %s -> %s", a, o.Op, b, z)
	return z
}

func (o List) Eval(c *Context) Val {
	var vec []Val
	for _, expr := range o.L {
		e := expr.Eval(c)
		vec = append(vec, e)
	}
	return &Mat{M: vec, S: []int{len(vec)}}
}

func (o Variable) String() string {
	return fmt.Sprintf("[%s]", o.S)
}

func (o Number) String() string {
	return fmt.Sprintf("[#%g]", o.F)
}

func (o Monad) String() string {
	return fmt.Sprintf("M(%s %s)", o.Op, o.B)
}

func (o Dyad) String() string {
	return fmt.Sprintf("D(%s %s %s)", o.A, o.Op, o.B)
}

func (o List) String() string {
	return fmt.Sprintf("C(%#v)", o.L)
}

func ParseExpr(lex *Lex, i int) (z Expression, zi int) {
/*
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("ParseExpr EXCEPTION %s >>> %s ,%d", r, z, zi)
			debug.PrintStack()
			panic(r)
		}
	}()
*/
	tt := lex.Tokens
	var vec []Expression
LOOP:
	for {
		t := tt[i]
		switch t.Type {
		case EndToken, CloseToken:
			break LOOP
		case OperatorToken:
			b, j := ParseExpr(lex, i+1)
			switch len(vec) {
			case 0:
				return &Monad{t.Str, b}, j
			case 1:
				return &Dyad{vec[0], t.Str, b}, j
			default:
				return &Dyad{&List{vec}, t.Str, b}, j
			}
		case NumberToken:
			num, err := strconv.ParseFloat(t.Str, 64)
			if err != nil {
				log.Panicf("Error parsing number %q at position %d: %s", t.Str, t.Pos, lex.Source)
			}
			vec = append(vec, &Number{num})
			i++
		case VariableToken:
			vec = append(vec, &Variable{t.Str})
			i++
		case OpenToken:
			b, j := ParseExpr(lex, i+1)
			vec = append(vec, b)
			i = j + 1
		default:
			log.Fatalf("bad default: %d", t.Type)
		}
	}

	if len(vec) > 1 {
		return &List{vec}, i
	} else if len(vec) == 1 {
		return vec[0], i
	}
	panic(0)
}
