package livy

import (
	"math"
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

	"asin": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Asin(b)
	})),
	"acos": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Acos(b)
	})),
	"atan": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Atan(b)
	})),
	"sin": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Sin(b)
	})),
	"cos": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Cos(b)
	})),
	"tan": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Tan(b)
	})),
	"asinh": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Asinh(b)
	})),
	"acosh": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Acosh(b)
	})),
	"atanh": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Atanh(b)
	})),
	"sinh": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Sinh(b)
	})),
	"cosh": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Cosh(b)
	})),
	"tanh": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Tanh(b)
	})),
	"exp": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Exp(b)
	})),
	"exp2": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Exp2(b)
	})),
	"expm1": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Expm1(b)
	})),
	"Log": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Log(b)
	})),
	"log10": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Log10(b)
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
	"round": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Round(b)
	})),
	"roundToEven": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.RoundToEven(b)
	})),
	"cbrt": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Cbrt(b)
	})),
	"sqrt": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Sqrt(b)
	})),
	"double": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return b + b
	})),
	"square": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
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
	"abs": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Abs(b)
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
	"isNaN": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return boolf(math.IsNaN(b))
	})),
	"y0": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Y0(b)
	})),
	"y1": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return math.Y1(b)
	})),
	"neg": WrapMatMonadic(WrapFloatMonadic(func(b float64) float64 {
		return -b
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

func MkEachOp(name string, fn MonadicFunc) MonadicFunc {
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

func WrapFloatMonadic(fn funcFloatFloat) MonadicFunc {
	return func(c *Context, b Val, axis int) Val {
		y := b.GetScalarFloat()
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
		vec[i] = &Num{float64(i + k)}
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
			vec[i] = &Num{float64(y.S[i])}
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
		reversed = append(reversed, &Num{float64(i)})
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
		switch mod(i, rank) {
		case mod(rank+axis, rank):
			spec = append(spec, &Num{float64(mod(i-1, rank))})
		case mod(rank+axis-1, rank):
			spec = append(spec, &Num{float64(mod(i+1, rank))})
		default:
			spec = append(spec, &Num{float64(mod(i, rank))})
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
		outVec = append(outVec, &Num{float64(ints[j])})
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
