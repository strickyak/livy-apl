package extend

import (
	"fmt"
	"log"

	. "github.com/strickyak/livy-apl/lib"
	chirp "github.com/yak-labs/chirp-lang"
)

func MatToTcl(mat *Mat) chirp.T {
	switch len(mat.S) {
	case 0:
		return chirp.MkList(nil)
	case 1:
		var vec []chirp.T
		for i := 0; i < mat.S[0]; i++ {
			vec = append(vec, ValToTcl(mat.M[i]))
		}
		return chirp.MkList(vec)
	default:
		stride := Product(mat.S[1:])
		var vec []chirp.T
		for i := 0; i < mat.S[0]; i++ {
			vec = append(vec, MatToTcl(&Mat{M: mat.M[i*stride : (i+1)*stride], S: mat.S[1:]}))
		}
		return chirp.MkList(vec)
	}
}

func ValToTcl(v Val) chirp.T {
	switch t := v.(type) {
	case *ChirpBox:
		return t.X
	case *Num:
		Must(imag(t.F) == 0) // Until I figure out what to do with imaginary numbers.
		return chirp.MkFloat(real(t.F))
	case *Mat:
		return MatToTcl(t)
	case *Box:
		return chirp.MkT(t.X)
	default:
		panic("unknown case")
	}
}

func monadicTcl(c *Context, b Val, axis int) Val {
	tcl := ValToTcl(b)
	frame := c.Extra["chirp"].(*chirp.Frame)
	z := frame.Eval(tcl)
	return &ChirpBox{z}
}

func ChirpStringExtension(s string) Expression {
	return &Literal{&ChirpBox{chirp.MkString(s)}}
}

func Init(c *Context) {
	c.StringExtension = ChirpStringExtension
	c.Extra["chirp"] = chirp.NewInterpreter()
	c.Monadics["tcl"] = monadicTcl
}

type ChirpBox struct {
	X chirp.T
}

func (o ChirpBox) String() string {
	return fmt.Sprintf("%q ", o.X.String())
}

func (o ChirpBox) Pretty() string {
	return fmt.Sprintf("%q ", o.X.String())
}

func (o ChirpBox) GetScalarInt() int {
	// Convert int64 to int.
	long := o.X.Int()
	i := int(long)
	if int64(i) != long {
		log.Panicf("ChirpBox int64 too big for Livy int: %d", long)
	}
	return i
}

func (o ChirpBox) GetScalarFloat() float64 {
	return o.X.Float()
}

func (o ChirpBox) GetScalarCx() complex128 {
	return complex(o.X.Float(), 0)
}

func (o ChirpBox) GetScalarOrNil() Val {
	return o
}

func (o ChirpBox) Size() int {
	return 1
}

func (o ChirpBox) Shape() []int {
	return nil
}

func (o ChirpBox) Ravel() []Val {
	return []Val{o}
}

func (o ChirpBox) ValEnum() ValEnum {
	return ChirpVal
}

func (a ChirpBox) Compare(x Val) int {
	aa := a.X.String()
	bb := x.String()
	switch {
	case aa < bb:
		return -1
	case aa == bb:
		return 0
	case aa > bb:
		return +1
	}
	panic("NOT_REACHED")
}
