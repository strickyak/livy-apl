package fft

import (
	"github.com/mjibson/go-dsp/fft"
	. "github.com/strickyak/livy-apl/lib"
	"log"
)

func monadicFFT(c *Context, b Val, dim int) Val {
	return monadicCxSliceToCxSlice(c, b, dim, fft.FFT, "fft")
}

func monadicIFFT(c *Context, b Val, dim int) Val {
	return monadicCxSliceToCxSlice(c, b, dim, fft.IFFT, "ifft")
}

func monadicCxSliceToCxSlice(c *Context, b Val, dim int, fn func([]complex128) []complex128, name string) Val {
	Must(dim == -1) // For now.
	m, ok := b.(*Mat)
	if !ok {
		log.Panicf("%v expected arg of type (*Mat), got type (%T)", name, b)
	}
	rank := len(m.S)
	if rank != 1 {
		log.Panicf("%v expected arg of rank 1, got rank %d", name, rank)
	}
	vec := make([]complex128, m.S[0])
	for i := 0; i < m.S[0]; i++ {
		vec[i] = m.M[i].GetScalarCx()
	}
	out := fn(vec)
	zz := make([]Val, len(out))
	for i := 0; i < len(vec); i++ {
		zz[i] = CxNum(out[i])
	}
	return &Mat{M: zz, S: []int{len(zz)}}
}

func init() {
	StandardMonadics["fft"] = monadicFFT
	StandardMonadics["ifft"] = monadicIFFT
}
