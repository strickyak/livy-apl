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
func (o Dyad) Eval(c *Context) Val {
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

func ParseDyadic(lex *Lex, i int) (Expression, int) {
	println("PD <-", i)
	tt := lex.Tokens
	n := len(tt)
	if i+1 >= n {
		log.Printf("PD: i+1 (%d) >= n (%d)", i+1, n)
		return nil, i
	}
	op := tt[i+1]
	if op.Type != OperatorToken {
		log.Printf("PD: op.Type (%d) != OperatorToken (%d)", op.Type, OperatorToken)
		return nil, i
	}
	b, j := Parse(lex, i+2)
	log.Printf("PD: Yes b=%s j=%d", b, j)
	return b, j
}

func ParseTail(lex *Lex, a Expression, i int) (Expression, int) {
	for {
		b, j := ParseDyadic(lex, i)
		if b != nil {
			a, i = &Dyad{a, lex.Tokens[i+1].Str, b}, j
			continue
		}
		return a, j
	}
}

func Parse(lex *Lex, i int) (Expression, int) {
	tt := lex.Tokens
	n := len(tt)
	if i >= n {
		log.Panicf("Parse out of tokens: i %d >= len %d", i, n)
	}
	t := tt[i]

	log.Printf("PARSE CONSIDER %d: %s", i, t)
	switch t.Type {
	case EndToken:
		log.Panicf("Parse EndToken")
	case NumberToken:
		num, err := strconv.ParseFloat(t.Str, 64)
		if err != nil {
			log.Panicf("Error parsing number %q at position %d: %s", t.Str, t.Pos, lex.Source)
		}
		var a Expression = &Number{num}
		return ParseTail(lex, a, i)
	case VariableToken:
		var a Expression = &Variable{t.Str}
		return ParseTail(lex, a, i)
	case OperatorToken:
		b, j := Parse(lex, i+1)
		return &Monad{t.Str, b}, j
	case OpenToken:
		a, j := Parse(lex, i+1)
		if lex.Tokens[j+1].Type != CloseToken {
			log.Panicf("Expected close paren at position %d: %s", lex.Tokens[j].Pos, lex.Source)
		}
		i = j + 1
		return ParseTail(lex, a, i)
	}
	log.Fatalf("bad default: %d", t.Type)
	panic("not reached")
}
