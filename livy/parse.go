package livy

import (
	"fmt"
	"log"
	"strconv"
)

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

func (o Variable) Eval(c *Context) Val {
	z, ok := c.Globals[o.S]
	if !ok {
		log.Panicf("No such variable %q in Globals %#v", o.S, c.Globals)
	}
	return z
}
func (o Number) Eval(c *Context) Val {
	return Num{o.F}
}
func (o Monad) Eval(c *Context) Val {
	m, ok := c.Monadics[o.Op]
	if !ok {
		log.Panicf("No such mondaic operator %q", o.Op)
	}
	b := o.B.Eval(c)
	return m(c, b)
}
func (o Dyad) Eval(c *Context) Val {
	return o.B.Eval(c)
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

func Parse(lex *Lex, i int) Expression {
	tt := lex.Tokens
	n := len(tt)
	if i >= n {
		log.Panicf("Parse out of tokens: i %d >= len %d", i, n)
	}
	t := tt[i]

	switch t.Type {
	case EndToken:
		log.Panicf("Parse EndToken")
	case NumberToken:
		num, err := strconv.ParseFloat(t.Str, 64)
		if err != nil {
			log.Panicf("Error parsing number %q at position %d: %s", t.Str, t.Pos, lex.Source)
		}
		return &Number{num}
	case VariableToken:
		return &Variable{t.Str}
	case OperatorToken:
		b := Parse(lex, i+1)
		return &Monad{t.Str, b}
	default:
		log.Fatalf("wut")
	}
	return nil
}
