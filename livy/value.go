package livy

import (
	"fmt"
	//"log"
	"bytes"
)

type Val interface {
	String() string
}

type Char struct {
	C rune
}

type Num struct {
	F float64
}

type Mat struct {
	M []Val
	S []int
}

type Box struct {
	X interface{}
}

func (o Char) String() string {
	return fmt.Sprintf("'%c' ", o.C)
}
func (o Num) String() string {
	return fmt.Sprintf("%g ", o.F)
}
func (o Mat) String() string {
	var bb bytes.Buffer
	bb.WriteString("[")
	for _, d := range o.S {
		fmt.Fprintf(&bb, "%d ", d)
	}
	bb.WriteString("]{")
	for _, v := range o.M {
		bb.WriteString(v.String())
	}
	bb.WriteString("} ")
	return bb.String()
}
func (o Box) String() string {
	return fmt.Sprintf("<Box> ")
}
