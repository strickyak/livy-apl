package livy

import (
	"fmt"
	"log"
	"runtime/debug"
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
	F complex128
}

type Literal struct {
	V Val
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

type Cond struct {
	If   *Seq
	Then *Seq
	Else *Seq
}

type While struct {
	While *Seq
	Do    *Seq
}

type Subscript struct {
	Matrix Expression
	Vec    []Expression
}

func (o Variable) Eval(c *Context) Val {
	z, ok := c.Globals[o.S]
	if !ok {
		Log.Panicf("No such variable %q in Globals %#v", o.S, c.Globals)
	}
	return z
}
func (o Number) Eval(c *Context) Val {
	return &Num{o.F}
}
func (o Literal) Eval(c *Context) Val {
	return o.V
}
func (o Monad) Eval(c *Context) Val {
	var ok bool
	var fn MonadicFunc
	switch o.Token.Type {
	case OperatorToken:
		fn, ok = c.Monadics[o.Op]
		if !ok {
			Log.Panicf("No such monadaic operator %q", o.Op)
		}
	case ReduceToken:
		op1 := o.Token.Match[1]
		fn1, ok := c.Dyadics[op1]
		if !ok {
			Log.Panicf("Reduce syntax: No such dyadaic operator %q", op1)
		}
		identity, ok := IdentityValueOfDyadic[op1]
		if !ok {
			identity = Zero
		}
		fn = MkReduceOrScanOp(o.Token.Str, fn1, identity, false)
	case ScanToken:
		op1 := o.Token.Match[1]
		fn1, ok := c.Dyadics[op1]
		if !ok {
			Log.Panicf("Scan syntax: No such dyadaic operator %q", op1)
		}
		identity, ok := IdentityValueOfDyadic[op1]
		if !ok {
			identity = Zero
		}
		fn = MkReduceOrScanOp(o.Token.Str, fn1, identity, true)
	case EachToken:
		op1 := o.Token.Match[1]
		fn1, ok := c.Monadics[op1]
		if !ok {
			Log.Panicf("Each syntax: No such dyadaic operator %q", op1)
		}
		fn = MkEachOpMonadic(o.Token.Str, fn1)
	}

	b := EvalFor(c, o.B, "RHS of Monadic expression", o.Op)
	Log.Printf("Monad:Eval %s %s -> ?", o.Op, b)
	axis := DefaultAxis
	if o.Axis != nil {
		axis = EvalFor(c, o.Axis, "Axis of Monadic expression", o.Op).GetScalarInt()
	}
	z := CallFor(c, func() Val { return fn(c, b, axis) }, "evaluation of Monadic function", o.Op)
	Log.Printf("Monad:Eval %s %s -> %s", o.Op, b, z)
	return z
}

func (o Dyad) Assign(c *Context) Val {
	b := EvalFor(c, o.B, "Assignment", o.Op)
	switch t := o.A.(type) {
	case *Subscript:
		return t.Assign(c, b)

	case *Variable:
		c.Globals[t.S] = b
		Log.Printf("Assigning %s = %s", t.S, b)
		return b

	default:
		Log.Panicf("cannot assign to %s", o.A)
		panic(0)
	}
}
func (o Dyad) Eval(c *Context) Val {
	if o.Op == "=" {
		return o.Assign(c)
	}
	var fn DyadicFunc
	switch o.Token.Type {
	default:
		Log.Panicf("Default case: token %v", o.Token.Type)
	case OperatorToken:
		fn1, ok := c.Dyadics[o.Op]
		if !ok {
			Log.Panicf("No such dyadaic operator %q", o.Op)
		}
		fn = fn1
	case EachToken:
		op1 := o.Token.Match[1]
		fn1, ok := c.Dyadics[op1]
		if !ok {
			Log.Panicf("Each syntax: No such dyadaic operator %q", op1)
		}
		fn = MkEachOpDyadic(o.Token.Str, fn1)
	case InnerProductToken:
		op1 := o.Token.Match[1]
		fn1, ok := c.Dyadics[op1]
		if !ok {
			Log.Panicf("Inner product syntax: No such dyadaic operator %q", op1)
		}
		op2 := o.Token.Match[2]
		fn2, ok := c.Dyadics[op2]
		if !ok {
			Log.Panicf("Inner product syntax: No such dyadaic operator %q", op2)
		}
		fn = MkInnerProduct(o.Token.Str, fn1, fn2)
	case OuterProductToken:
		op1 := o.Token.Match[1]
		fn1, ok := c.Dyadics[op1]
		if !ok {
			Log.Panicf("Outer product syntax: No such dyadaic operator %q", op1)
		}
		fn = MkOuterProduct(o.Token.Str, fn1)
	}

	b := EvalFor(c, o.B, "RHS of Dyadic expression", o.Op)
	axis := DefaultAxis
	if o.Axis != nil {
		axis = EvalFor(c, o.Axis, "Axis of Dyadic expression", o.Op).GetScalarInt()
	}
	a := EvalFor(c, o.A, "LHS of Dyadic expression", o.Op)

	Log.Printf("Dyad:Eval %s %s %s -> ?", a, o.Op, b)
	z := CallFor(c, func() Val { return fn(c, a, b, axis) }, "evaluation of Dyadic function", o.Op)
	Log.Printf("Dyad:Eval %s %s %s -> %s", a, o.Op, b, z)

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

func (o Cond) Eval(c *Context) Val {
	cond := o.If.Eval(c)
	b := float2bool(cond.GetScalarFloat())
	if b {
		return o.Then.Eval(c)
	} else {
		return o.Else.Eval(c)
	}
}

type Break struct{}
type Continue struct{}

var BREAK Break
var CONTINUE Continue

func (Break) Eval(c *Context) Val {
	panic(BREAK)
}
func (Continue) Eval(c *Context) Val {
	panic(CONTINUE)
}
func (o While) Eval(c *Context) Val {
	var z []Val
	for {
		cond := o.While.Eval(c)
		b := float2bool(cond.GetScalarFloat())
		if !b {
			break
		}

		var doBreak bool
		item := func() Val {
			defer func() {
				r := recover()
				switch r.(type) {
				case nil:
					return
				case Break:
					doBreak = true
					return
				case Continue:
					return
				default:
					panic(r)
				}
			}()
			return o.Do.Eval(c)
		}()
		if doBreak {
			break
		}
		if item != nil {
			z = append(z, item)
		}
	}
	return &Mat{z, []int{len(z)}}
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
			c.Globals[o.Axis] = &Num{complex(float64(axis), 0)}
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

func (o Subscript) PreEval(c *Context) (mat *Mat, newShape []int, subscripts [][]int) {
	var ok bool
	lhs := o.Matrix.Eval(c)
	mat, ok = lhs.(*Mat)
	if !ok {
		Log.Panicf("Cannot subscript non-matrix: %s", lhs)
	}
	rank := len(mat.S)
	if len(o.Vec) != rank {
		Log.Panicf("Number of subscripts %d does not match rank %d of matrix: %s", len(o.Vec), rank, lhs)
	}

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
	return mat, newShape, subscripts
}
func (o Subscript) Eval(c *Context) Val {
	mat, newShape, subscripts := o.PreEval(c)
	newSize := Product(newShape)
	newMat := &Mat{M: make([]Val, newSize), S: newShape}
	if len(newShape) > 0 {
		copyIntoSubscriptedMatrix(newShape, subscripts, 0, mat, mat.S, newMat.M, 0)
	}
	return newMat
}
func (o Subscript) Assign(c *Context, b Val) Val {
	bmat, ok := b.(*Mat)
	if !ok {
		Log.Panicf("Cannot assign non-matrix to subscripted variable: %v", b)
	}

	avar, ok := o.Matrix.(*Variable)
	if !ok {
		Log.Panicf("Cannot assign to subscripted non-variable: %s", o.Matrix)
	}

	aval := avar.Eval(c)
	amat, ok := aval.(*Mat)
	if !ok {
		Log.Panicf("Cannot assign to subscripted non-matrix variable: %s", amat)
	}

	rank := len(amat.S)
	if len(o.Vec) != rank {
		Log.Panicf("Number of subscripts %d does not match rank %d of matrix: %s", len(o.Vec), rank, aval)
	}

	// Replace mat with a copy, that can be modified.
	matM := make([]Val, len(amat.M)) // Alloc new contents.
	copy(matM, amat.M)               // Copy the contents.
	mat := &Mat{matM, amat.S}        // New mat with newly copied contents.  Shape is immutable and can be shared.

	var subscripts [][]int
	for i, sub := range o.Vec {
		if sub == nil {
			// For missing subscripts, use entire range available in mat's shape.
			subscripts = append(subscripts, intRange(mat.S[i]))
		} else {
			r := EvalFor(c, sub, "subscripts in subscripted assignment of variable", "").Ravel()
			ints := make([]int, len(r))
			for i, e := range r {
				ints[i] = e.GetScalarInt()
			}
			subscripts = append(subscripts, ints)
		}
	}

	oldOff := 0
	var recurse func(subscripts [][]int, newShape []int, newOff int)
	recurse = func(subscripts [][]int, newShape []int, newOff int) {
		Log.Printf("subscripts=%v newShape=%v newOff=%d oldOff=%d", subscripts, newShape, newOff, oldOff)
		if len(newShape) == 0 {
			matM[newOff] = bmat.M[oldOff]
			oldOff++
			return
		}
		newStride := Product(newShape[1:])
		for _, j := range subscripts[0] {
			recurse(subscripts[1:], newShape[1:], newOff+j*newStride)
		}

	}
	recurse(subscripts, amat.S, 0)
	c.Globals[avar.S] = mat
	return b
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
			nextOffset := offset + Product(shape[1:])*i
			nextSubOffset := subOffset + Product(matShape[1:])*subscripts[0][i]
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

func Product(v []int) int {
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

func (o Literal) String() string {
	return fmt.Sprintf("(%q)", o.V.String())
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

func (o Cond) String() string {
	return fmt.Sprintf("Cond(if %v then %v else %v fi)", o.If, o.Then, o.Else)
}

func (o While) String() string {
	return fmt.Sprintf("While(while %v do %v done)", o.While, o.Do)
}

func (o Break) String() string {
	return fmt.Sprintf("Break")
}

func (o Continue) String() string {
	return fmt.Sprintf("Continue")
}

func (o Seq) String() string {
	return fmt.Sprintf("Seq(%#v)", o.Vec)
}

func (o Subscript) String() string {
	return fmt.Sprintf("Sub(%v [ %v ])", o.Matrix, o.Vec)
}

func float2bool(f float64) bool {
	if f == 1.0 {
		return true
	}
	if f == 0.0 {
		return false
	}
	Log.Panicf("Cannot use %.18g as a bool", f)
	panic(0)
}

func cx2bool(c complex128) bool {
	re, im := real(c), imag(c)
	if im != 0 {
		Log.Panicf("Cannot use %s as a bool", Cx2Str(c))
	}
	if re == 1.0 {
		return true
	}
	if re == 0.0 {
		return false
	}
	Log.Panicf("Cannot use %s as a bool", Cx2Str(c))
	panic(0)
}

func EvalFor(c *Context, expr Expression, why string, what string) Val {
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("... during: %s: %s", why, what)
			panic(r)
		}
	}()
	return expr.Eval(c)
}

func CallFor(c *Context, fn func() Val, why string, what string) Val {
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("... during: %s: %s", why, what)
			panic(r)
		}
	}()
	return fn()
}
