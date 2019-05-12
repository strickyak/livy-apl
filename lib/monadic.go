package livy

import (
	"math"
	"math/cmplx"
	"sort"
)

type MonadicFunc func(c *Context, b Val, dim int) Val

var StandardMonadics = map[string]MonadicFunc{
	"box":   monadicBox,
	"unbox": monadicUnbox,
	"b":     monadicBox,
	"u":     monadicUnbox,

	"up":        monadicUp,
	"down":      monadicDown,
	"transpose": transposeMonadic,
	",":         ravelMonadic,
	"rot":       rotMonadic,
	"iota":      iotaMonadic,
	"iota1":     iota1Monadic,
	"rho":       rhoMonadic,

	// Abbreviations for iota & rho are i & p
	"i":  iotaMonadic,
	"i1": iota1Monadic,
	"p":  rhoMonadic,
	"j": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(0.0, 1.0) * b
	})),
	"real": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(real(b), 0.0)
	})),
	"imag": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(imag(b), 0.0)
	})),
	"rect": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		Must(imag(b) == 0)
		return cmplx.Rect(1.0, real(b))
	})),
	"isInf": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return Bool2Cx(cmplx.IsInf(b))
	})),
	"isNaN": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return Bool2Cx(cmplx.IsNaN(b))
	})),
	"asin": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Asin(b)
	})),
	"acos": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Acos(b)
	})),
	"atan": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Atan(b)
	})),
	"sin": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Sin(b)
	})),
	"cos": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Cos(b)
	})),
	"tan": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Tan(b)
	})),
	"asinh": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Asinh(b)
	})),
	"acosh": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Acosh(b)
	})),
	"atanh": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Atanh(b)
	})),
	"sinh": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Sinh(b)
	})),
	"cosh": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Cosh(b)
	})),
	"tanh": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Tanh(b)
	})),
	"exp": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Exp(b)
	})),
	"exp2": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Exp2(b)
	})),
	"expm1": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Expm1(b)
	})),
	"log": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Log(b)
	})),
	"log10": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Log10(b)
	})),
	"log2": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Log2(b)
	})),
	"log1p": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Log1p(b)
	})),
	"ceil": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Ceil(b)
	})),
	"floor": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Floor(b)
	})),
	"round": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(math.Round(real(b)), math.Round(imag(b)))
	})),
	"round1": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(math.Round(real(10*b))/10, math.Round(imag(10*b))/10)
	})),
	"round2": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(math.Round(real(100*b))/100, math.Round(imag(100*b))/100)
	})),
	"round3": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(math.Round(real(1000*b))/1000, math.Round(imag(1000*b))/1000)
	})),
	"round4": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(math.Round(real(10000*b))/10000, math.Round(imag(10000*b))/10000)
	})),
	"round5": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(math.Round(real(100000*b))/100000, math.Round(imag(100000*b))/100000)
	})),
	"round6": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(math.Round(real(1000000*b))/1000000, math.Round(imag(1000000*b))/10000000)
	})),
	"round7": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(math.Round(real(10000000*b))/10000000, math.Round(imag(10000000*b))/100000000)
	})),
	"round8": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(math.Round(real(100000000*b))/100000000, math.Round(imag(100000000*b))/1000000000)
	})),
	"round9": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(math.Round(real(1000000000*b))/1000000000, math.Round(imag(1000000000*b))/10000000000)
	})),
	"roundToEven": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.RoundToEven(b)
	})),
	"ki":     WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1024 * b })),
	"mi":     WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1024 * 1024 * b })),
	"gi":     WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1024 * 1024 * 1024 * b })),
	"ti":     WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1024 * 1024 * 1024 * 1024 * b })),
	"pi":     WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1024 * 1024 * 1024 * 1024 * 1024 * b })),
	"ei":     WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * b })),
	"ks":     WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1000 * b })),
	"ms":     WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1000 * 1000 * b })),
	"gs":     WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1000 * 1000 * 1000 * b })),
	"ts":     WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1000 * 1000 * 1000 * 1000 * b })),
	"ps":     WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1000 * 1000 * 1000 * 1000 * 1000 * b })),
	"es":     WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1000 * 1000 * 1000 * 1000 * 1000 * 1000 * b })),
	"millis": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return b / 1000 })),
	"micros": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return b / 1000 / 1000 })),
	"nanos":  WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return b / 1000 / 1000 / 1000 })),
	"picos":  WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return b / 1000 / 1000 / 1000 / 1000 })),
	"div":    WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 { return 1 / b })),
	"cbrt": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		if imag(b) == 0 {
			return complex(math.Cbrt(real(b)), 0)
		} else {
			return cmplx.Pow(b, complex(1.0/3.0, 0))
		}
	})),
	"sqrt": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return cmplx.Sqrt(b)
	})),
	"double": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return b + b
	})),
	"square": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return b * b
	})),
	"sgn": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		if b < 0 {
			return -1
		} else if b > 0 {
			return +1
		} else if b == 0 {
			return 0
		} else {
			panic("cannot sgn")
		}
	})),
	"abs": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(cmplx.Abs(b), 0)
	})),
	"phase": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(cmplx.Phase(b), 0.0)
	})),
	"erf": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Erf(b)
	})),
	"erfc": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Erfc(b)
	})),
	"erfinv": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Erfinv(b)
	})),
	"erfcinv": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Erfcinv(b)
	})),
	"gamma": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Gamma(b)
	})),
	"inf": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Inf(int(b))
	})),
	"y0": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Y0(b)
	})),
	"y1": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Y1(b)
	})),
	"neg": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return -b
	})),
	"-": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return -b
	})),
	"+": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return +b
	})),
	"conjugate": WrapMatMonadic(WrapCxMonadic(func(b complex128) complex128 {
		return complex(real(b), -imag(b))
	})),
	"not": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		x := float2bool(b)
		return boolf(!x)
	})),
}

func boolf(a bool) float64 {
	if a {
		return 1.0
	} else {
		return 0.0
	}
}

func MkEachOpMonadic(name string, fn MonadicFunc) MonadicFunc {
	return func(c *Context, b Val, axis int) Val {
		mat, ok := b.(*Mat)
		if !ok {
			Log.Panicf("Each operator %s~ expects matrix argument, got %s", name, b)
		}
		vec := make([]Val, len(mat.M))
		for i, x := range mat.M {
			vec[i] = fn(c, x, axis)
		}
		return &Mat{vec, mat.S}
	}
}

type funcFloatFloat func(b float64) float64
type funcCxCx func(b complex128) complex128

func WrapFloatMonadic(fn funcFloatFloat) MonadicFunc {
	return func(c *Context, b Val, axis int) Val {
		y := b.GetScalarFloat()
		return &Num{complex(fn(y), 0)}
	}
}

func WrapCxMonadic(fn funcCxCx) MonadicFunc {
	return func(c *Context, b Val, axis int) Val {
		y := b.GetScalarCx()
		return &Num{fn(y)}
	}
}

func doubleMonadic(c *Context, b Val) Val {
	switch y := b.(type) {
	case *Num:
		return &Num{2 * y.F}
	}
	Log.Panicf("Wrong type for monadic `double`: %T %q", b, b)
	return nil
}

func ravelMonadic(c *Context, b Val, axis int) Val {
	mat, ok := b.(*Mat)
	if ok {
		// Grab ravelled guts from the matrix.
		return &Mat{mat.M, []int{len(mat.M)}}
	}
	// Singleton vector.
	return &Mat{[]Val{b}, []int{1}}
}

func iotaMonadic(c *Context, b Val, axis int) Val {
	return iotaK(c, b, 0)
}
func iota1Monadic(c *Context, b Val, axis int) Val {
	return iotaK(c, b, 1)
}
func iotaK(c *Context, b Val, k int) Val {
	n := b.GetScalarInt()
	vec := make([]Val, n)
	for i := 0; i < n; i++ {
		vec[i] = &Num{complex(float64(i+k), 0)}
	}
	return &Mat{
		M: vec,
		S: []int{n},
	}
}

func rhoMonadic(c *Context, b Val, axis int) Val {
	switch y := b.(type) {
	case *Mat:
		n := len(y.S)
		vec := make([]Val, n)
		for i := 0; i < n; i++ {
			vec[i] = &Num{complex(float64(y.S[i]), 0)}
		}
		return &Mat{
			M: vec,
			S: []int{n},
		}
	default:
		return &Mat{
			M: nil,
			S: nil,
		}
	}
}

func WrapMatMonadic(fn MonadicFunc) MonadicFunc {
	return func(c *Context, b Val, axis int) Val {
		switch y := b.(type) {
		case *Mat:
			n := len(y.M)
			vec := make([]Val, n)

			for i := 0; i < n; i++ {
				y1 := y.M[i].GetScalarOrNil()
				if y1 == nil {
					Log.Panicf("arg not a scalar at matrix offset %d: %s", i, y1)
				}
				vec[i] = fn(c, y1, axis)
			}

			return &Mat{M: vec, S: y.S}

		}

		ys := b.GetScalarOrNil()
		if ys == nil {
			Log.Panicf("arg not scalar or matrix")
		}
		return fn(c, ys, axis)
	}
}

func rotMonadic(c *Context, b Val, axis int) Val {
	mat, ok := b.(*Mat)
	if !ok {
		// scalar is like 1x1, whose rot or flip is itself.
		return b
	}

	shape := mat.S
	n := len(shape)
	if n < 1 {
		// rot or flip on Emptiness yields Emptiness.
		return b
	}
	axis = (((axis % n) + n) % n)

	axisLen := shape[axis]
	var reversed []Val
	for i := axisLen - 1; i >= 0; i-- {
		reversed = append(reversed, &Num{complex(float64(i), 0)})
	}
	return dyadicRot(c, &Mat{reversed, []int{axisLen}}, b, axis)
}

func transposeMonadic(c *Context, b Val, axis int) Val {
	mat, ok := b.(*Mat)
	if !ok {
		Log.Panicf("Monadic `transpose` needs matrix on right, but got %v", b)
	}

	shape := mat.S
	rank := len(shape)
	if rank < 2 {
		Log.Panicf("Monadic `transpose` needs matrix with rank >= 2, but got shape %v", shape)
	}

	var spec []Val
	for i := 0; i < rank; i++ {
		switch Mod(i, rank) {
		case Mod(rank+axis, rank):
			spec = append(spec, &Num{complex(float64(Mod(i-1, rank)), 0)})
		case Mod(rank+axis-1, rank):
			spec = append(spec, &Num{complex(float64(Mod(i+1, rank)), 0)})
		default:
			spec = append(spec, &Num{complex(float64(Mod(i, rank)), 0)})
		}
	}
	lhs := &Mat{spec, []int{len(spec)}}

	return dyadicTranspose(c, lhs, b, -1)
}

type IndexedValSlice struct {
	Vals []Val
	Ints []int
}

func (p IndexedValSlice) Len() int {
	return len(p.Vals)
}

func (p IndexedValSlice) Less(i, j int) bool {
	return Compare(p.Vals[p.Ints[i]], p.Vals[p.Ints[j]]) < 0
}

func (p *IndexedValSlice) Swap(i, j int) {
	p.Ints[i], p.Ints[j] = p.Ints[j], p.Ints[i]
}

func monadicUp(c *Context, b Val, axis int) Val {
	return monadicUpDown(c, b, false, "up")
}
func monadicDown(c *Context, b Val, axis int) Val {
	return monadicUpDown(c, b, true, "down")
}
func monadicUpDown(c *Context, b Val, reverse bool, name string) Val {
	mat, ok := b.(*Mat)
	if !ok {
		Log.Panicf("monadic `%s` wants matrix, got %v", name, b)
	}
	if len(mat.S) != 1 {
		Log.Panicf("monadic `%s` wants matrix of rank 1, got %v", name, b)
	}

	n := mat.S[0]
	ints := make([]int, n)
	for i := range ints {
		ints[i] = i
	}

	sorter := &IndexedValSlice{Vals: mat.M, Ints: ints}
	sort.Sort(sorter)

	var outVec []Val
	for i := 0; i < n; i++ {
		j := i
		if reverse {
			j = n - 1 - i
		}
		outVec = append(outVec, &Num{complex(float64(ints[j]), 0)})
	}

	return &Mat{outVec, []int{n}}
}
func monadicBox(c *Context, b Val, axis int) Val {
	return &Box{b}
}
func monadicUnbox(c *Context, b Val, axis int) Val {
	box, ok := b.(*Box)
	if !ok {
		Log.Panicf("In unbox, not a box: %T: %v", b, b)
	}
	val, ok := box.X.(Val)
	if !ok {
		Log.Panicf("In unbox, not an apl value in the box: %T: %v", box.X, box.X)
	}
	return val
}
