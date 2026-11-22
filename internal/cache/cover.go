// Package cache stores fetched images on disk behind a byte-budgeted LRU and
// transcodes them to the configured image format.
package cache

import (
	"bytes"
	"encoding/json"
	"image"
	// ponytail: decoders are jpeg/png/webp; add gif/bmp when a plugin needs them
	_ "image/jpeg" // register decoder: covers arrive as jpg
	_ "image/png"  // register decoder: covers arrive as png
	"os"

	"github.com/gen2brain/webp"
	"golang.org/x/image/draw"
)

// ResizeCover downscales data so its longest edge is at most max pixels while
// preserving the aspect ratio, encodes the result as lossy WebP and returns
// the encoded bytes with the width/height ratio of the result. Input that
// image.Decode cannot parse is returned unchanged with ratio 0 and no error;
// max <= 0 skips the downscale.
func ResizeCover(data []byte, max int) ([]byte, float64, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, 0, nil
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := src
	longest := w
	if h > w {
		longest = h
	}
	if max > 0 && longest > max {
		resized := image.NewRGBA(image.Rect(0, 0, w*max/longest, h*max/longest))
		draw.CatmullRom.Scale(resized, resized.Bounds(), src, src.Bounds(), draw.Src, nil)
		dst = resized
	}
	var out bytes.Buffer
	if err := webp.Encode(&out, dst); err != nil {
		return nil, 0, err
	}
	bounds := dst.Bounds()
	return out.Bytes(), float64(bounds.Dx()) / float64(bounds.Dy()), nil
}

// sidecar is the JSON payload written next to a cached image.
type sidecar struct {
	Ratio float64 `json:"ratio"`
}

// WriteRatio stores ratio in the "<path>.json" sidecar belonging to the image
// at path, replacing any previous content.
func WriteRatio(path string, ratio float64) error {
	b, err := json.Marshal(sidecar{Ratio: ratio})
	if err != nil {
		return err
	}
	return os.WriteFile(path+".json", b, 0o644)
}
