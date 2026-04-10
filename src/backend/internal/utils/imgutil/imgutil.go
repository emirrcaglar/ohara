package imgutil

import (
	"image"
	"image/jpeg"
	_ "image/png"
	"io"

	"golang.org/x/image/draw"
)

func Compress(r io.Reader, w io.Writer, maxWidth, quality int) error {
	src, _, err := image.Decode(r)
	if err != nil {
		return err
	}

	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	var out image.Image = src
	if srcW > maxWidth {
		newH := int(float64(srcH) * float64(maxWidth) / float64(srcW))
		dst := image.NewRGBA(image.Rect(0, 0, maxWidth, newH))
		draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)
		out = dst
	}

	return jpeg.Encode(w, out, &jpeg.Options{Quality: quality})
}
