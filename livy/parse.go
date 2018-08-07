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

type Cons struct {
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

func (o Cons) Eval(c *Context) Val {
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

func (o Cons) String() string {
	return fmt.Sprintf("C(%#v)", o.L)
}

func ParseDyadic(lex *Lex, i int) (Expression, int) {
	//log.Printf("PD <- %d", i)
	tt := lex.Tokens
	n := len(tt)
	if i+1 >= n {
		//log.Printf("PD: i+1 (%d) >= n (%d)", i+1, n)
		return nil, i
	}
	op := tt[i+1]
	if op.Type != OperatorToken {
		//log.Printf("PD: op.Type (%d) != OperatorToken (%d)", op.Type, OperatorToken)
		return nil, i
	}
	b, j := Parse(lex, i+2)
	//log.Printf("PD: Yes b=%s j=%d", b, j)
	return b, j
}

func ParseTail(lex *Lex, a Expression, i int) (Expression, int) {
	var list []Expression
	list = append(list, a)
Loop:
	for i+1 < len(lex.Tokens) {
		t := lex.Tokens[i+1]
		log.Printf("ParseTail...i+1=[%d]  %s", i+1, t)
		switch t.Type {
		case NumberToken:
			num, err := strconv.ParseFloat(t.Str, 64)
			if err != nil {
				log.Panicf("Error parsing number %q at position %d: %s", t.Str, t.Pos, lex.Source)
			}
			list = append(list, &Number{num})
			i++
		case VariableToken:
			list = append(list, &Variable{t.Str})
			i++
		case OpenToken:
			log.Printf("Open on token %s at i+1=%d", t, i+1)
			b, j := Parse(lex, i+1)
			log.Printf("Close %s returns j=%d", b, j)
			list = append(list, b)
			i = j + 1
		default:
			log.Printf("Break on token %s at i+1=%d", t, i+1)
			break Loop
		}
	}
	if len(list) > 1 {
		a = &Cons{list}
	}

	for {
		b, j := ParseDyadic(lex, i)
		if b != nil {
			a, i = &Dyad{a, lex.Tokens[i+1].Str, b}, j
			continue
		}
		i = j
		return a, i
	}
}

func Parse(lex *Lex, i int) (Expression, int) {
	tt := lex.Tokens
	n := len(tt)
	if i >= n {
		log.Panicf("Parse out of tokens: i %d >= len %d", i, n)
	}
	t := tt[i]

	//log.Printf("PARSE CONSIDER %d: %s", i, t)
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
	case CloseToken:
		log.Panicf("Close paren not expected at position %d: %s", t.Pos, lex.Source)
	}
	log.Fatalf("bad default: %d", t.Type)
	panic("not reached")
}
