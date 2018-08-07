package livy

import (
	"bytes"
	"fmt"
	"log"
)

type Val interface {
	String() string
	GetScalarInt() int
	GetScalarFloat() float64
	GetScalarOrNil() Val
}

type Char struct {
	R rune
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
	return fmt.Sprintf("'%c' ", o.R)
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

func (o Char) GetScalarInt() int {
	log.Panicf("Char cannot be a Scalar Int: '%c'", o.R)
	panic(0)
}
func (o Num) GetScalarInt() int {
	a := int(o.F)
	if float64(a) != o.F {
		log.Panicf("Not an integer: %g", o.F)
	}
	return a
}
func (o Mat) GetScalarInt() int {
	if len(o.M) == 1 {
		return o.M[1].GetScalarInt()
	}
	log.Panicf("Matrix with %d entries cannot be a Scalar Int", len(o.M))
	panic(0)
}
func (o Box) GetScalarInt() int {
	log.Panicf("Box cannot be a Scalar Int")
	panic(0)
}

func (o Char) GetScalarFloat() float64 {
	log.Panicf("Char cannot be a Scalar Float: '%c'", o.R)
	panic(0)
}
func (o Num) GetScalarFloat() float64 {
	return o.F
}
func (o Mat) GetScalarFloat() float64 {
	if len(o.M) == 1 {
		return o.M[1].GetScalarFloat()
	}
	log.Panicf("Matrix with %d entries cannot be a Scalar Float", len(o.M))
	panic(0)
}
func (o Box) GetScalarFloat() float64 {
	log.Panicf("Box cannot be a Scalar Float")
	panic(0)
}

func (o Char) GetScalarOrNil() Val {
	return o
}
func (o Num) GetScalarOrNil() Val {
	return o
}
func (o Mat) GetScalarOrNil() Val {
	if len(o.M) == 1 {
		return o.M[0].GetScalarOrNil()
	}
	return nil
}
func (o Box) GetScalarOrNil() Val {
	return o
}
