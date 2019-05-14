package livy

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
)

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
	GetScalarCx() complex128
	GetScalarOrNil() Val
}

/*
type Char struct {
	R rune
}
*/

type Num struct {
	F complex128
}

type Mat struct {
	M []Val
	S []int
}

type Box struct {
	X interface{}
}

var CX_REGEXP = `([-+0-9.eE]+)(([-+])j([-+0-9.eE]+))?`
var MatchCx = regexp.MustCompile(CX_REGEXP)

func ParseCx(s string) complex128 {
	m := MatchCx.FindStringSubmatch(s)
	if m == nil {
		log.Panicf("cannot parse complex number: %q", s)
	}
	_re, _sign, _im := m[1], m[3], m[4]
	re, err := strconv.ParseFloat(_re, 64)
	if err != nil {
		log.Panicf("cannot parse complex number: %q", s)
	}
	if len(_sign) == 1 {
		im, err := strconv.ParseFloat(_im, 64)
		if err != nil {
			log.Panicf("cannot parse complex number: %q", s)
		}
		if _sign == "-" {
			return complex(re, -im)
		} else {
			return complex(re, +im)
		}
	} else {
		return complex(re, 0)
	}
}

func Cx2Str(c complex128) string {
	return Num{c}.String()
}

/*
func (o Char) String() string {
	return fmt.Sprintf("'%c' ", o.R)
}
*/

func (o Num) String() string {
	rl, im := real(o.F), imag(o.F)
	if im == 0 {
		return fmt.Sprintf("%g ", rl)
	} else if rl == 0 && im > 0 {
		return fmt.Sprintf("+j%g ", im)
	} else if rl == 0 && im < 0 {
		return fmt.Sprintf("-j%g ", -im)
	} else if im < 0 {
		return fmt.Sprintf("%g-j%g ", rl, -im)
	} else {
		return fmt.Sprintf("%g+j%g ", rl, +im)
	}
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
	return fmt.Sprintf("Box(%v) ", o.X)
}

/*
func (o Char) Pretty() string {
	return fmt.Sprintf("'%c' ", o.R)
}
*/
func (o Num) Pretty() string {
	return fmt.Sprintf("%s ", o)
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
			begin := i * Product(o.S[1:])
			end := (i + 1) * Product(o.S[1:])
			bb.WriteString(Mat{M: o.M[begin:end], S: o.S[1:]}.PrettyMatrix(vec[begin:end]))
			bb.WriteString("\n")
		}
	}
	return bb.String()
}

func BoundingBoxString(s string) (w int, h int) {
	w, h = 1, 1
	c := 0 // cursor
	for _, r := range s {
		if r == '\n' {
			if c > w {
				w = c
			}
			c = 0
			h++
		} else {
			c++
		}
	}
	if c > w {
		w = c
	}
	//log.Printf("BBS: %q -> w=%d, h=%d", s, w, h)
	return
}

func FillString(w, h int, s string) []string {
	var zz []string

	// Convert string to runes.
	var rr []rune
	for _, r := range s {
		rr = append(rr, r)
	}
	i, rrlen := 0, len(rr)

	for y := 0; y < h; y++ {
		var bb bytes.Buffer
		for x := 0; x < w; x++ {
			if i >= rrlen {
				bb.WriteByte(' ')
			} else if rr[i] == '\n' {
				bb.WriteByte(' ')
			} else {
				bb.WriteRune(rr[i])
				i++
			}
		}
		zz = append(zz, bb.String())
		i++
	}
	//log.Printf("FillString(%d, %d): <- %q", w, h, s)
	//for i, z := range zz {
	//log.Printf("FillString[ %d ] : -> %q", i, z)
	//}
	return zz
}

func RenderBoxString(sss [][]string, border int) string {
	var z bytes.Buffer
	var maxw, maxh []int

	//for y, ss := range sss {
	//for x, s := range ss {
	//log.Printf("Render (%d, %d): %q", x, y, s)
	//}
	//}

	for y, ss := range sss {
		for x, s := range ss {
			w, h := BoundingBoxString(s)
			if len(maxw) <= x {
				maxw = append(maxw, w)
			} else {
				if maxw[x] < w {
					maxw[x] = w
				}
			}
			if len(maxh) <= y {
				maxh = append(maxh, h)
			} else {
				if maxh[y] < h {
					maxh[y] = h
				}
			}
		}
	}

	for y, ss := range sss {
		var ff [][]string
		for x, s := range ss {
			f := FillString(maxw[x], maxh[y], s)
			ff = append(ff, f)
		}

		for i := 0; i < maxh[y]; i++ {
			for _, f := range ff {
				z.WriteString(f[i])
			}
			z.WriteByte('\n')
		}
		for i := 0; i < border; i++ {
			z.WriteByte('\n')
		}
	}

	s := z.String()

	// Trim trailing newlines.
	n := len(s)
	for n > 0 && s[n-1] == '\n' {
		n--
	}
	return s[:n]
}

func RenderPrettyMatrix(mat Mat) string {
	var hologram [][]string // as if it were 2d.

	in := mat.M
	lastLen := mat.S[len(mat.S)-1]

	var recurse func(shape []int, p int, last bool)
	recurse = func(shape []int, p int, last bool) {
		switch len(shape) {
		case 0:
			panic("impossible")
		case 1:
			var ss []string
			for i := 0; i < shape[0]; i++ {
				s := in[p+i].Pretty()
				ss = append(ss, s)
			}
			hologram = append(hologram, ss)
		default:
			stride := Product(shape[1:])
			for i := 0; i < shape[0]; i++ {
				recurse(shape[1:], p+i*stride, last && i == shape[0]-1)
			}

			if !last {
				hologram = append(hologram, make([]string, lastLen))
			}
		}
	}
	recurse(mat.S, 0, true)
	z := RenderBoxString(hologram, 0)
	return z
}

func (o Mat) Pretty() string {
	var bb bytes.Buffer
	rank := len(o.S)
	switch rank {
	case 0:
		log.Panicf("(* OH NO: Mat rank 0 *) ")
	case 1:
		return RenderPrettyMatrix(o)
		//TODO
		if len(o.M) != o.S[0] {
			log.Panicf("matrix shape %v but contains %d elements: %#v", o.S, len(o.M), o)
		}
		for _, v := range o.M {
			bb.WriteString(v.String())
		}
	default:
		return RenderPrettyMatrix(o)
		//TODO

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
				//Log.Printf("%d/%d/%q", i, len(ss[i]), ss[i])
			}
		}

		return o.PrettyMatrix(ss)
	}
	return bb.String()
}

const box_h = 0x2550
const box_v = 0x2551
const box_nw = 0x2554
const box_ne = 0x2557
const box_sw = 0x255A
const box_se = 0x255D

func (o Box) Pretty() string {
	var s string
	switch t := o.X.(type) {
	case Val:
		s = t.Pretty()
	case string:
		return fmt.Sprintf("%q ", t) // Omit drawing a box.
	case fmt.Stringer:
		s = fmt.Sprintf("%q ", t.String())
	default:
		s = fmt.Sprintf("(Box of %T: %v) ", o.X, o.X)
	}

	w, h := BoundingBoxString(s)
	ff := FillString(w, h, s)
	var bb bytes.Buffer
	bb.WriteRune(box_nw)
	for i := 0; i < w; i++ {
		bb.WriteRune(box_h)
	}
	bb.WriteRune(box_ne)
	bb.WriteByte(' ')
	bb.WriteByte('\n')
	for _, f := range ff {
		bb.WriteRune(box_v)
		bb.WriteString(f)
		bb.WriteRune(box_v)
		bb.WriteByte(' ')
		bb.WriteByte('\n')
	}
	bb.WriteRune(box_sw)
	for i := 0; i < w; i++ {
		bb.WriteRune(box_h)
	}
	bb.WriteRune(box_se)
	bb.WriteByte(' ')
	return bb.String()
}

/*
func (o Char) GetScalarInt() int {
	Log.Panicf("Char cannot be a Scalar Int: '%c'", o.R)
	panic(0)
}
*/
func (o Num) GetScalarInt() int {
	re, im := real(o.F), imag(o.F)
	if im != 0 {
		log.Panicf("Number has imag part, cannot be used as integer: %s", Cx2Str(o.F))
	}
	a := int(re)
	if float64(a) != re {
		Log.Panicf("Not an integer: %s", Cx2Str(o.F))
	}
	return a
}
func (o Mat) GetScalarInt() int {
	if len(o.M) == 1 {
		return o.M[0].GetScalarInt()
	}
	Log.Panicf("Matrix with %d entries cannot be a Scalar Int", len(o.M))
	panic(0)
}
func (o Box) GetScalarInt() int {
	Log.Panicf("Box cannot be a Scalar Int")
	panic(0)
}

/*
func (o Char) GetScalarCx() complex128 {
	Log.Panicf("Char cannot be a Scalar Complex: '%c'", o.R)
	panic(0)
}
func (o Char) GetScalarFloat() float64 {
	Log.Panicf("Char cannot be a Scalar Float: '%c'", o.R)
	panic(0)
}
*/
func (o Num) GetScalarCx() complex128 {
	return o.F
}
func (o Num) GetScalarFloat() float64 {
	if imag(o.F) != 0 {
		log.Panicf("Number has imag part, cannot be used as real: %s", Cx2Str(o.F))
	}
	return real(o.F)
}
func (o Mat) GetScalarCx() complex128 {
	if len(o.M) == 1 {
		return o.M[1].GetScalarCx()
	}
	Log.Panicf("Matrix with %d entries cannot be a Scalar Complex", len(o.M))
	panic(0)
}
func (o Mat) GetScalarFloat() float64 {
	if len(o.M) == 1 {
		return o.M[1].GetScalarFloat()
	}
	Log.Panicf("Matrix with %d entries cannot be a Scalar Float", len(o.M))
	panic(0)
}
func (o Box) GetScalarCx() complex128 {
	Log.Panicf("Box cannot be a Scalar Complex")
	panic(0)
}

func (o Box) GetScalarFloat() float64 {
	Log.Panicf("Box cannot be a Scalar Float")
	panic(0)
}

/*
func (o Char) GetScalarOrNil() Val {
	return o
}
*/
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

/*
func (o Char) Size() int {
	return 1
}
*/
func (o Num) Size() int {
	return 1
}
func (o Mat) Size() int {
	return len(o.M)
}
func (o Box) Size() int {
	return 1
}

/*
func (o Char) Shape() []int {
	return nil
}
*/
func (o Num) Shape() []int {
	return nil
}
func (o Mat) Shape() []int {
	return o.S
}
func (o Box) Shape() []int {
	return nil
}

/*
func (o Char) Ravel() []Val {
	return []Val{o}
}
*/
func (o Num) Ravel() []Val {
	return []Val{o}
}
func (o Mat) Ravel() []Val {
	return o.M
}
func (o Box) Ravel() []Val {
	return []Val{o}
}

/*
func (o Char) ValEnum() ValEnum {
	return CharVal
}
*/
func (o Num) ValEnum() ValEnum {
	return NumVal
}
func (o Mat) ValEnum() ValEnum {
	return MatVal
}
func (o Box) ValEnum() ValEnum {
	return BoxVal
}

/*
func (a Char) Compare(x Val) int {
	b, ok := x.(*Char)
	if !ok {
		Log.Panicf("Char::Compare to not-a-Char: %#v", x)
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
*/
func (a Num) Compare(x Val) int {
	fa := a.GetScalarFloat()
	fx := x.GetScalarFloat()
	switch {
	case fa < fx:
		return -1
	case fa == fx:
		return 0
	case fa > fx:
		return +1
	}
	panic("NOT_REACHED")
}
func (a Mat) Compare(x Val) int {
	b, ok := x.(*Mat)
	if !ok {
		Log.Panicf("Mat::Compare to not-a-Mat: %v", x)
	}
	switch {
	case len(a.S) < len(b.S):
		return -1
	case len(a.S) > len(b.S):
		return +1
	}
	for i := range a.S {
		fa := a.S[i]
		fb := b.S[i]
		switch {
		case fa < fb:
			return -1
		case fa > fb:
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
		Log.Panicf("Box::Compare to not-a-Box: %v", x)
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

func Bool2Cx(b bool) complex128 {
	if b {
		return 1
	} else {
		return 0
	}
}
func BoolNum(b bool) *Num {
	if b {
		return One
	} else {
		return Zero
	}
}
func CxNum(f complex128) *Num {
	return &Num{f}
}

func FloatNum(f float64) *Num {
	return &Num{complex(f, 0.0)}
}

func IntNum(i int) *Num {
	return &Num{complex(float64(i), 0.0)}
}

func Must(predicate bool) {
	if !predicate {
		panic("FAILED: Must()")
	}
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}
