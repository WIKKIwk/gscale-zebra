package labelprint

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestBuildImageLabelZPL_ContainsCoreCommands(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.Black)
	img.Set(1, 0, color.White)
	img.Set(0, 1, color.White)
	img.Set(1, 1, color.Black)

	zpl, err := BuildImageLabelZPL(img, 100, 60)
	if err != nil {
		t.Fatalf("BuildImageLabelZPL error: %v", err)
	}

	s := string(zpl)
	checks := []string{
		"^XA",
		"^PW100",
		"^LL60",
		"^GFA",
		"^XZ",
	}
	for _, c := range checks {
		if !strings.Contains(s, c) {
			t.Fatalf("zpl missing %q: %s", c, s)
		}
	}
}

func TestFitImageToLabel_NoOverflow(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	fitted := fitImageToLabel(img, 560, 320)

	if fitted.Bounds().Dx() > 560 {
		t.Fatalf("width overflow: %d", fitted.Bounds().Dx())
	}
	if fitted.Bounds().Dy() > 320 {
		t.Fatalf("height overflow: %d", fitted.Bounds().Dy())
	}
}
