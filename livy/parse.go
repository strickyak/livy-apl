package livy

// “Should array indices start at 0 or 1?
// My compromise of 0.5 was rejected without, I thought, proper consideration.”
//      — Stan Kelly-Bootle
//   — http://exple.tive.org/blarg/2013/10/22/citation-needed/

// "So let us let our ordinals start at zero: ..."
//   — Edsger W. Dijkstra

// I used to describe [Pascal] as a ‘fascist programming language’,
// because it is dictatorially rigid. …
// If Pascal is fascist, APL is anarchist.
//   — Brad McCormick
// http://www.computerhistory.org/atchm/the-apl-programming-language-source-code/#footnote-13
// http://www.users.cloud9.net/~bradmcc/APL.html

// APL is a mistake, carried through to perfection. It is the language
// of the future for the programming techniques of the past: it creates a
// new generation of coding bums.
//   — Edsger W. Dijkstra
// http://www.computerhistory.org/atchm/the-apl-programming-language-source-code/#footnote-14
// https://www.cs.virginia.edu/~evans/cs655/readings/ewd498.html

import (
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
)

const DefaultAxis = -1

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
	Token *Token
	Op    string
	B     Expression
	Axis  Expression
}

type Dyad struct {
	Token *Token
	A     Expression
	Op    string
	B     Expression
	Axis  Expression
}

type List struct {
	Vec []Expression
}

type Seq struct {
	Vec []Expression
}

type Def struct {
	Name   string
	Seq    *Seq
	Lhs    string
	Axis   string
	Rhs    string
	Locals []string
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
	axis := DefaultAxis
	if o.Axis != nil {
		axis = o.Axis.Eval(c).GetScalarInt()
	}
	z := fn(c, b, axis)
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
	var fn DyadicFunc
	switch o.Token.Type {
	case OperatorToken:
		fn1, ok := c.Dyadics[o.Op]
		if !ok {
			log.Panicf("No such dyadaic operator %q", o.Op)
		}
		fn = fn1
	case InnerProductToken:
		op1 := o.Token.Match[1]
		fn1, ok := c.Dyadics[op1]
		if !ok {
			log.Panicf("No such dyadaic operator %q", op1)
		}
		op2 := o.Token.Match[2]
		fn2, ok := c.Dyadics[op2]
		if !ok {
			log.Panicf("No such dyadaic operator %q", op2)
		}
		fn = MkInnerProduct(o.Token.Str, fn1, fn2)
	case OuterProductToken:
		op1 := o.Token.Match[1]
		fn1, ok := c.Dyadics[op1]
		if !ok {
			log.Panicf("No such dyadaic operator %q", op1)
		}
		fn = MkOuterProduct(o.Token.Str, fn1)
	}
	a := o.A.Eval(c)
	axis := DefaultAxis
	if o.Axis != nil {
		axis = o.Axis.Eval(c).GetScalarInt()
	}
	b := o.B.Eval(c)
	log.Printf("Dyad:Eval %s %s %s -> ?", a, o.Op, b)
	z := fn(c, a, b, axis)
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

func (o Def) Eval(c *Context) Val {
	fn := func(c *Context, a Val, b Val, axis int) Val {
		// We do the stupid thing where all variables
		// are used from c.Globals context, but we save and
		// restore global variables shadowed by local
		// variables on entry and exit to functions.
		localMap := make(map[string]Val)
		for _, lvar := range o.Locals {
			gval, _ := c.Globals[lvar]
			localMap[lvar] = gval
			c.Globals[lvar] = &Num{0}
		}
		c.LocalStack = append(c.LocalStack, localMap)
		defer func() {
			n_1 := len(c.LocalStack) - 1
			localMap = c.LocalStack[n_1]
			c.LocalStack = c.LocalStack[:n_1]
			for _, lvar := range o.Locals {
				saved := localMap[lvar]
				if saved == nil {
					delete(c.Globals, lvar)
				} else {
					c.Globals[lvar] = saved
				}
			}
		}()

		if o.Lhs != "" {
			c.Globals[o.Lhs] = a
		}
		if o.Axis != "" {
			c.Globals[o.Axis] = &Num{float64(axis)}
		}
		c.Globals[o.Rhs] = b
		return o.Seq.Eval(c)
	}

	if o.Lhs == "" {
		c.Monadics[o.Name] = func(c *Context, b Val, axis int) Val {
			return fn(c, nil, b, axis)
		}
	} else {
		c.Dyadics[o.Name] = func(c *Context, a Val, b Val, axis int) Val {
			return fn(c, a, b, axis)
		}
	}
	return &Box{"def"}
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
func (o Subscript) Assign(c *Context, a Val) Val {
	v := o.Var.Eval(c)
	mat, ok := v.(*Mat)
	if !ok {
		log.Panicf("Cannot subscript non-matrix: %s", v)
	}

	rank := len(mat.S)
	if len(o.Vec) != rank {
		log.Panicf("Number of subscripts %d does not match rank %d of matrix: %s", len(o.Vec), rank, v)
	}

	// Replace mat with a copy, that can be modified.
	matM := make([]Val, len(mat.M)) // Alloc new contents.
	copy(matM, mat.M)               // Copy the contents.
	mat = &Mat{matM, mat.S}         // New mat with new contents.  Shape is immutable and can be shared.

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

	// TODO zzzzzzzzzzzzzzzzz TODO

	/*
		newSize := mulReduce(newShape) // Now this is the size of a, that we assign from.
		newMat := &Mat{M: make([]Val, newSize), S: newShape}
		if len(newShape) > 0 {
			copyIntoSubscriptedMatrix(newShape, subscripts, 0, mat, mat.S, newMat.M, 0)
		}
	*/
	return a
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
	sub := ""
	if o.Axis != nil {
		sub = fmt.Sprintf("[%d]", o.Axis)
	}
	return fmt.Sprintf("Monad(%s%s %s)", o.Op, sub, o.B)
}

func (o Dyad) String() string {
	sub := ""
	if o.Axis != nil {
		sub = fmt.Sprintf("[%d]", o.Axis)
	}
	return fmt.Sprintf("Dyad(%s %s%s %s)", o.A, o.Op, sub, o.B)
}

func (o List) String() string {
	return fmt.Sprintf("List(%#v)", o.Vec)
}

func (o Def) String() string {
	return fmt.Sprintf("Def(%q %q [%q] %q %v %v)", o.Lhs, o.Name, o.Axis, o.Rhs, o.Locals, o.Seq)
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

func ParseSeq(lex *Lex, i int) (*Seq, int) {
	tt := lex.Tokens
	var vec []Expression
LOOP:
	for i < len(tt) && tt[i].Type != EndToken {
		log.Printf("ParseSeq: i=%d max=%d token=%s", i, len(tt), tt[i])
		b, j := ParseExpr(lex, i)
		log.Printf("ParseSeq: i=%d b=%s", i, b)
		vec = append(vec, b)
		i = j

		switch tt[i].Type {
		case EndToken, CloseCurlyToken:
			break LOOP
		case SemiToken:
			i++
			continue LOOP
		default:
			log.Fatalf("default: %d %s", i, tt[i])
		}
	}

	return &Seq{vec}, i
}

func ParseDef(lex *Lex, i int) (*Def, int) {
	var lhs, axis, rhs string
	var locals []string

	tt := lex.Tokens
	t := tt[i]
	// Expect operator.
	if t.Type == VariableToken {
		lhs = t.Str
		locals = append(locals, lhs)
		i++
		t = tt[i]
	}

	if t.Type != OperatorToken {
		log.Fatalf("expected operator after def, but got %v", t)
	}
	name := t.Str
	i++
	t = tt[i]
	if t.Type == BraToken {
		i++
		t = tt[i]
		if t.Type != VariableToken {
			log.Fatalf("expected AXIS variable after def operator open-bracket, but got %v", t)
		}
		axis = t.Str
		locals = append(locals, axis)
		i++
		t = tt[i]
		if t.Type != KetToken {
			log.Fatalf("expected close-bracket def operator open-bracket axis, but got %v", t)
		}
		i++
		t = tt[i]
	}
	if t.Type != VariableToken {
		log.Fatalf("expected RHS variable after def, but got %v", t)
	}
	rhs = t.Str
	locals = append(locals, rhs)
	i++
	t = tt[i]

	for t.Type == SemiToken {
		i++
		t = tt[i]
		// Allow extra semicolon before open curly.
		if t.Type == OpenCurlyToken {
			break
		}
		if t.Type != VariableToken {
			log.Fatalf("expected local variable name after def semicolon, but got %v", t)
		}
		locals = append(locals, t.Str)
		i++
		t = tt[i]
	}

	if t.Type != OpenCurlyToken {
		log.Fatalf("expected open-curly-brace after operator after def, but got %v", t)
	}
	i++

	seq, j := ParseSeq(lex, i)
	i = j
	t = tt[i]

	if t.Type != CloseCurlyToken {
		log.Fatalf("expected close-curly-brace after operator after def, but got %v", t)
	}
	i++
	return &Def{name, seq, lhs, axis, rhs, locals}, i
}

func ParseExpr(lex *Lex, i int) (z Expression, zi int) {
	tt := lex.Tokens
	var vec []Expression
LOOP:
	for {
		t := tt[i]
		switch t.Type {
		case KeywordToken:
			switch t.Str {
			case "def":
				def, j := ParseDef(lex, i+1)
				vec = append(vec, def)
				i = j
			default:
				panic("not yet")
			}
		case EndToken, CloseToken, KetToken, SemiToken, CloseCurlyToken:
			break LOOP
		case BraToken:
			log.Panicf("Unexpected `[` at position %d: %s", t.Pos, lex.Source)
		case OpenCurlyToken:
			log.Panicf("Unexpected `{` at position %d: %s", t.Pos, lex.Source)
		case InnerProductToken, OuterProductToken, OperatorToken:
			axis := Expression(nil)
			var j int
			if tt[i+1].Type == BraToken {
				log.Printf("Axis1")
				axis, j = ParseExpr(lex, i+2)
				log.Printf("Axis2 %d %s", j, axis)
				if tt[j].Type != KetToken {
					log.Panicf("Expected ']' but got %q after subscript", tt[i].Str)
				}
				i = j // Don't add 1 here; ParseExpr just below gets i+1.
			}

			b, j := ParseExpr(lex, i+1)
			switch len(vec) {
			case 0:
				return &Monad{t, t.Str, b, axis}, j
			case 1:
				return &Dyad{t, vec[0], t.Str, b, axis}, j
			default:
				return &Dyad{t, &List{vec}, t.Str, b, axis}, j
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
	log.Printf("VEC=%v", vec)
	return vec[0], i
}
