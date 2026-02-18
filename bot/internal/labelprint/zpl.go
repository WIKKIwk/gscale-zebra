package labelprint

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strings"
)

const monoThreshold = 180

func decodeImage(payload []byte) (image.Image, string, error) {
	img, format, err := image.Decode(bytes.NewReader(payload))
	if err != nil {
		return nil, "", err
	}
	return img, format, nil
}

func BuildImageLabelZPL(src image.Image, labelWidth, labelHeight int) ([]byte, error) {
	if src == nil {
		return nil, fmt.Errorf("rasm yo'q")
	}
	if labelWidth <= 0 || labelHeight <= 0 {
		return nil, fmt.Errorf("label o'lchami noto'g'ri")
	}

	fitted := fitImageToLabel(src, labelWidth, labelHeight)
	w := fitted.Bounds().Dx()
	h := fitted.Bounds().Dy()
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("rasm o'lchami noto'g'ri")
	}

	bitmap, rowBytes := imageToMonochrome(fitted)
	if len(bitmap) == 0 {
		return nil, fmt.Errorf("rasm bitmapga aylantirilmadi")
	}

	totalBytes := len(bitmap)
	x := (labelWidth - w) / 2
	if x < 0 {
		x = 0
	}
	y := (labelHeight - h) / 2
	if y < 0 {
		y = 0
	}

	hexData := strings.ToUpper(hex.EncodeToString(bitmap))
	zpl := fmt.Sprintf(
		"^XA\n^MMT\n^PW%d\n^LL%d\n^LH0,0\n^FO%d,%d^GFA,%d,%d,%d,%s^FS\n^PQ1,0,1,N\n^XZ\n",
		labelWidth,
		labelHeight,
		x,
		y,
		totalBytes,
		totalBytes,
		rowBytes,
		hexData,
	)
	return []byte(zpl), nil
}

func fitImageToLabel(src image.Image, maxW, maxH int) image.Image {
	srcB := src.Bounds()
	sw := srcB.Dx()
	sh := srcB.Dy()
	if sw <= 0 || sh <= 0 || maxW <= 0 || maxH <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	scale := math.Min(float64(maxW)/float64(sw), float64(maxH)/float64(sh))
	if scale <= 0 {
		scale = 1
	}

	dw := int(math.Round(float64(sw) * scale))
	dh := int(math.Round(float64(sh) * scale))
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}
	if dw > maxW {
		dw = maxW
	}
	if dh > maxH {
		dh = maxH
	}

	return resizeNearest(src, dw, dh)
}

func resizeNearest(src image.Image, dstW, dstH int) *image.RGBA {
	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}

	srcB := src.Bounds()
	sw := srcB.Dx()
	sh := srcB.Dy()
	if sw <= 0 || sh <= 0 {
		return image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		sy := srcB.Min.Y + (y*sh)/dstH
		if sy >= srcB.Max.Y {
			sy = srcB.Max.Y - 1
		}
		for x := 0; x < dstW; x++ {
			sx := srcB.Min.X + (x*sw)/dstW
			if sx >= srcB.Max.X {
				sx = srcB.Max.X - 1
			}
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func imageToMonochrome(img image.Image) ([]byte, int) {
	b := img.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w <= 0 || h <= 0 {
		return nil, 0
	}

	rowBytes := (w + 7) / 8
	out := make([]byte, rowBytes*h)
	threshold := uint32(monoThreshold)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, bl, a := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			if a < 0x4000 {
				continue
			}

			gray := ((299*(r>>8) + 587*(g>>8) + 114*(bl>>8)) / 1000)
			if gray >= threshold {
				continue
			}

			idx := y*rowBytes + (x / 8)
			bit := uint(7 - (x % 8))
			out[idx] |= 1 << bit
		}
	}

	return out, rowBytes
}
