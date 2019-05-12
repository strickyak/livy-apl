package image

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	. "github.com/strickyak/livy-apl/lib"
)

func loadImage(filename string) (image.Image, int, int) {
	r, err := os.Open(filename)
	Check(err)
	defer func() { r.Close() }()
	im, _, err := image.Decode(r)
	Check(err)
	b := im.Bounds()
	Must(b.Min.X == 0)
	Must(b.Min.Y == 0)
	return im, b.Max.X, b.Max.Y
}

func loadImageToVal(filename string) Val {
	im, mx, my := loadImage(filename)
	var vec []Val
	for x := 0; x < mx; x++ {
		for y := 0; y < my; y++ {
			_r, _b, _g, _a := im.At(x, y).RGBA()
			vec = append(vec, FloatNum(float64(_r)/0xFFFF), FloatNum(float64(_b)/0xFFFF), FloatNum(float64(_g)/0xFFFF), FloatNum(float64(_a)/0xFFFF))
		}
	}
	return &Mat{vec, []int{mx, my, 4}}
}

func monadicImage(c *Context, b Val, dim int) Val {
	return loadImageToVal("/tmp/image")
}

func init() {
	StandardMonadics["image"] = monadicImage
}
