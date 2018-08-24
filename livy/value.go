package livy

import (
	"bytes"
	"fmt"
	"log"
)

type Val interface {
	String() string
	Pretty() string
	Size() int
	Shape() []int
	Ravel() []Val
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

func (o Char) Pretty() string {
	return fmt.Sprintf("'%c' ", o.R)
}
func (o Num) Pretty() string {
	return fmt.Sprintf("%g ", o.F)
}
func (o Mat) Pretty() string {
	var bb bytes.Buffer
	rank := len(o.S)
	switch rank {
	case 0:
		return "(* TODO: Mat rank 0 *) " /* + o.M[0].Pretty() */
	case 1:
		if len(o.M) != o.S[0] {
			log.Panicf("matrix shape %v but contains %d elements: %#v", o.S, len(o.M), o)
		}
		for _, v := range o.M {
			bb.WriteString(v.String())
		}
	case 2:
		for i := 0; i < o.S[0]; i++ {
			begin := i * o.S[1]
			end := (i + 1) * o.S[1]
			bb.WriteString(Mat{M: o.M[begin:end], S: o.S[1:]}.Pretty())
			bb.WriteString("\n")
		}
	default:
		for i := 0; i < o.S[0]; i++ {
			begin := i * mulReduce(o.S[1:])
			end := (i + 1) * mulReduce(o.S[1:])
			bb.WriteString(Mat{M: o.M[begin:end], S: o.S[1:]}.Pretty())
			bb.WriteString("\n")
		}
	}
	return bb.String()
}
func (o Box) Pretty() string {
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

func (o Char) Size() int {
	return 1
}
func (o Num) Size() int {
	return 1
}
func (o Mat) Size() int {
	return len(o.M)
}
func (o Box) Size() int {
	return 1
}

func (o Char) Shape() []int {
	return nil
}
func (o Num) Shape() []int {
	return nil
}
func (o Mat) Shape() []int {
	return o.S
}
func (o Box) Shape() []int {
	return nil
}

func (o Char) Ravel() []Val {
	return []Val{o}
}
func (o Num) Ravel() []Val {
	return []Val{o}
}
func (o Mat) Ravel() []Val {
	return o.M
}
func (o Box) Ravel() []Val {
	return []Val{o}
}
