package cache

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// gradientPNG renders a w x h gradient and encodes it as PNG.
func gradientPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return buf.Bytes()
}

func TestResizeCoverDownscalesLongestEdge(t *testing.T) {
	got, ratio, err := ResizeCover(gradientPNG(t, 2000, 3000), 800)
	if err != nil {
		t.Fatalf("ResizeCover: %v", err)
	}
	img, format, err := image.Decode(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("output not decodable: %v", err)
	}
	if format != "webp" {
		t.Fatalf("output format %q, want webp", format)
	}
	if w, h := img.Bounds().Dx(), img.Bounds().Dy(); w != 533 || h != 800 {
		t.Fatalf("resized to %dx%d, want 533x800", w, h)
	}
	if want := float64(533) / 800; ratio != want {
		t.Fatalf("ratio %v, want %v", ratio, want)
	}
}

func TestResizeCoverKeepsImagesUnderMax(t *testing.T) {
	got, ratio, err := ResizeCover(gradientPNG(t, 400, 300), 800)
	if err != nil {
		t.Fatalf("ResizeCover: %v", err)
	}
	img, _, err := image.Decode(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("output not decodable: %v", err)
	}
	if w, h := img.Bounds().Dx(), img.Bounds().Dy(); w != 400 || h != 300 {
		t.Fatalf("kept at %dx%d, want 400x300", w, h)
	}
	if want := float64(400) / 300; ratio != want {
		t.Fatalf("ratio %v, want %v", ratio, want)
	}
}

func TestResizeCoverUndecodableInput(t *testing.T) {
	in := []byte("<html>not an image</html>")
	got, ratio, err := ResizeCover(in, 800)
	if err != nil {
		t.Fatalf("ResizeCover: %v", err)
	}
	if ratio != 0 {
		t.Fatalf("ratio %v, want 0", ratio)
	}
	if !bytes.Equal(got, in) {
		t.Fatal("input bytes not returned unchanged")
	}
}

func TestWriteRatio(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cover.webp")
	if err := WriteRatio(path, 0.75); err != nil {
		t.Fatalf("WriteRatio: %v", err)
	}
	b, err := os.ReadFile(path + ".json")
	if err != nil {
		t.Fatalf("read sidecar: %v", err)
	}
	var got sidecar
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("parse sidecar %s: %v", b, err)
	}
	if got.Ratio != 0.75 {
		t.Fatalf("sidecar ratio %v, want 0.75", got.Ratio)
	}
}
