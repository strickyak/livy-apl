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
	Vec []Expression
}

type Seq struct {
	Vec []Expression
}

type Subscript struct {
	Var *Variable
	Vec []Expression
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
	for _, expr := range o.Vec {
		e := expr.Eval(c)
		vec = append(vec, e)
	}
	return &Mat{M: vec, S: []int{len(vec)}}
}

func (o Seq) Eval(c *Context) Val {
	if len(o.Vec) == 0 {
		panic("cant")
	}
	var z Val
	for _, expr := range o.Vec {
		z = expr.Eval(c)
	}
	return z
}

func (o Subscript) Eval(c *Context) Val {
	v := o.Var.Eval(c)
	mat, ok := v.(*Mat)
	if !ok {
		log.Panicf("Cannot subscript non-matrix: %s", v)
	}
	rank := len(mat.S)
	if len(o.Vec) != rank {
		log.Panicf("Number of subscripts %d does not match rank %d of matrix: %s", len(o.Vec), rank, v)
	}

	var newShape []int
	var subscripts [][]int
	for i, sub := range o.Vec {
		if sub == nil {
			// For missing subscripts, use entire range available in mat's shape.
			subscripts = append(subscripts, intRange(mat.S[i]))
			newShape = append(newShape, mat.S[i])
		} else {
			r := sub.Eval(c).Ravel()
			newShape = append(newShape, len(r))
			ints := make([]int, len(r))
			for i, e := range r {
				ints[i] = e.GetScalarInt()
			}
			subscripts = append(subscripts, ints)
		}
	}

	newSize := mulReduce(newShape)
	newMat := &Mat{M: make([]Val, newSize), S: newShape}
	if len(newShape) > 0 {
		copyIntoSubscriptedMatrix(newShape, subscripts, 0, mat, mat.S, newMat.M, 0)
	}
	return newMat
}

func copyIntoSubscriptedMatrix(shape []int, subscripts [][]int, subOffset int, mat *Mat, matShape []int, z []Val, offset int) {
	if shape[0] == 0 {
		return
	}
	if len(shape) == 1 {
		for i := 0; i < shape[0]; i++ {
			z[offset+i] = mat.M[subOffset+subscripts[0][i]]
		}
	} else {
		for i := 0; i < shape[0]; i++ {
			nextOffset := offset + mulReduce(shape[1:])*i
			nextSubOffset := subOffset + mulReduce(matShape[1:])*subscripts[0][i]
			copyIntoSubscriptedMatrix(shape[1:], subscripts[1:], nextSubOffset, mat, matShape[1:], z, nextOffset)
		}
	}
}

func intRange(n int) []int {
	z := make([]int, n)
	for i := 0; i < n; i++ {
		z[i] = i
	}
	return z
}

func mulReduce(v []int) int {
	z := 1
	for _, e := range v {
		z *= e
	}
	return z
}

func (o Variable) String() string {
	return fmt.Sprintf("(%s)", o.S)
}

func (o Number) String() string {
	return fmt.Sprintf("(#%g)", o.F)
}

func (o Monad) String() string {
	return fmt.Sprintf("Monad(%s %s)", o.Op, o.B)
}

func (o Dyad) String() string {
	return fmt.Sprintf("Dyad(%s %s %s)", o.A, o.Op, o.B)
}

func (o List) String() string {
	return fmt.Sprintf("List(%#v)", o.Vec)
}

func (o Seq) String() string {
	return fmt.Sprintf("Seq(%#v)", o.Vec)
}

func (o Subscript) String() string {
	return fmt.Sprintf("Sub(%s [ %#v ])", o.Var.S, o.Vec)
}

func ParseBracket(lex *Lex, i int) ([]Expression, int) {
	i++
	var vec []Expression
	var tmp Expression
	for {
		switch lex.Tokens[i].Type {
		case KetToken:
			vec = append(vec, tmp)
			i++
			return vec, i
		case SemiToken:
			i++
			vec = append(vec, tmp)
			tmp = nil
		default:
			// This is a bit weak.
			tmp, i = ParseExpr(lex, i)
		}
	}
}

func ParseSeq(lex *Lex) *Seq {
	tt := lex.Tokens
	var vec []Expression
	i := 0
LOOP:
	for i < len(tt) && tt[i].Type != EndToken {
		log.Printf("ParseSeq: i=%d max=%d token=%s", i, len(tt), tt[i])
		b, j := ParseExpr(lex, i)
		log.Printf("ParseSeq: i=%d b=%s", i, b)
		vec = append(vec, b)
		i = j

		switch tt[i].Type {
		case EndToken:
			break LOOP
		case SemiToken:
			i++
			continue LOOP
		default:
			log.Fatalf("default: %d %s", i, tt[i])
		}
	}

	return &Seq{vec}
}

func ParseExpr(lex *Lex, i int) (z Expression, zi int) {
	tt := lex.Tokens
	var vec []Expression
LOOP:
	for {
		t := tt[i]
		switch t.Type {
		case EndToken, CloseToken, KetToken, SemiToken:
			break LOOP
		case BraToken:
			log.Panicf("Unexpected `[` at position %d: %s", t.Pos, lex.Source)
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
			variable := &Variable{t.Str}
			i++
			if tt[i].Type == BraToken {
				log.Printf("B1")
				v, j := ParseBracket(lex, i)
				log.Printf("B2 %d %s", j, v)
				vec = append(vec, &Subscript{variable, v})
				log.Printf("B3 %s", vec)
				i = j
			} else {
				vec = append(vec, variable)
			}
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
	}
	return vec[0], i
}
