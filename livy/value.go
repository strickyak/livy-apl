package livy

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
)

type StringExtensionFunc func(s string) Expression

var StringExtension StringExtensionFunc

type ValEnum int

const (
	CharVal ValEnum = iota + 1
	NumVal
	MatVal
	BoxVal
	ChirpVal // github.com/yak-libs/chirp-lang extension
)

type Val interface {
	Compare(Val) int
	ValEnum() ValEnum
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
	return fmt.Sprintf("(%v)", o.X)
}

func (o Char) Pretty() string {
	return fmt.Sprintf("'%c' ", o.R)
}
func (o Num) Pretty() string {
	return fmt.Sprintf("%.15g ", o.F)
}
func (o Mat) PrettyMatrix(vec []string) string {
	var bb bytes.Buffer
	rank := len(o.S)
	switch rank {
	case 0:
		panic("bad case")
	case 1:
		for _, s := range vec {
			bb.WriteString(s)
		}
	default:
		for i := 0; i < o.S[0]; i++ {
			begin := i * mulReduce(o.S[1:])
			end := (i + 1) * mulReduce(o.S[1:])
			bb.WriteString(Mat{M: o.M[begin:end], S: o.S[1:]}.PrettyMatrix(vec[begin:end]))
			bb.WriteString("\n")
		}
	}
	return bb.String()
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
	default:
		var ss []string
		// Get String of each matrix element.
		for _, x := range o.M {
			ss = append(ss, x.String())
		}
		// Get widest by last dimension.
		lastDim := o.S[len(o.S)-1]
		for j := 0; j < lastDim; j++ {
			w := 0
			for i := j; i < len(ss); i += lastDim {
				if len(ss[i]) > w {
					w = len(ss[i])
				}
			}
			for i := j; i < len(ss); i += lastDim {
				s := ss[i]
				for len(s) < w {
					s = " " + s
				}
				ss[i] = s
				log.Printf("%d/%d/%q", i, len(ss[i]), ss[i])
			}
		}

		return o.PrettyMatrix(ss)
	}
	return bb.String()
}
func (o Box) Pretty() string {
	return fmt.Sprintf("(Box of %T: %v) ", o.X, o.X)
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
		return o.M[0].GetScalarInt()
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

func (o Char) ValEnum() ValEnum {
	return CharVal
}
func (o Num) ValEnum() ValEnum {
	return NumVal
}
func (o Mat) ValEnum() ValEnum {
	return MatVal
}
func (o Box) ValEnum() ValEnum {
	return BoxVal
}

func (a Char) Compare(x Val) int {
	b, ok := x.(*Char)
	if !ok {
		log.Panicf("Char::Compare to not-a-Char: %#v", x)
	}
	switch {
	case a.R < b.R:
		return -1
	case a.R == b.R:
		return 0
	case a.R > b.R:
		return +1
	}
	panic("NOT_REACHED")
}
func (a Num) Compare(x Val) int {
	b, ok := x.(*Num)
	if !ok {
		log.Panicf("Num::Compare to not-a-Num: %v", x)
	}
	switch {
	case a.F < b.F:
		return -1
	case a.F == b.F:
		return 0
	case a.F > b.F:
		return +1
	}
	panic("NOT_REACHED")
}
func (a Mat) Compare(x Val) int {
	b, ok := x.(*Mat)
	if !ok {
		log.Panicf("Mat::Compare to not-a-Mat: %v", x)
	}
	switch {
	case len(a.S) < len(b.S):
		return -1
	case len(a.S) > len(b.S):
		return +1
	}
	for i := range a.S {
		switch {
		case a.S[i] < b.S[i]:
			return -1
		case a.S[i] > b.S[i]:
			return +1
		}
	}
	for i := range a.M {
		cmp := a.M[i].Compare(b.M[i])
		if cmp != 0 {
			return cmp
		}
	}
	return 0
}
func (a Box) Compare(x Val) int {
	b, ok := x.(*Box)
	if !ok {
		log.Panicf("Box::Compare to not-a-Box: %v", x)
	}
	aa := reflect.ValueOf(a).Pointer()
	bb := reflect.ValueOf(b).Pointer()
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

func Compare(a, b Val) int {
	ae := a.ValEnum()
	be := b.ValEnum()
	if ae < be {
		return -1
	}
	if ae > be {
		return +1
	}
	return a.Compare(b)
}
