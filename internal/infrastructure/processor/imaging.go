package processor

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	thumbWidth  = 150
	thumbHeight = 150
)

type ImageProcessor struct {
}

func New() *ImageProcessor {
	return &ImageProcessor{}
}

func (p *ImageProcessor) Resize(ctx context.Context, contentType string, data []byte, width, height int) ([]byte, error) {
	img, err := decodeImage(data)
	if err != nil {
		return nil, fmt.Errorf("ImageProcessor - Resize - decodeImage: %w", err)
	}

	resized := imaging.Resize(img, width, height, imaging.Lanczos)

	res, err := encodeImage(resized, contentType)
	if err != nil {
		return nil, fmt.Errorf("ImageProcessor - Resize - encodeImage: %w", err)
	}

	return res, nil
}

func (p *ImageProcessor) Thumbnail(ctx context.Context, contentType string, data []byte) ([]byte, error) {
	img, err := decodeImage(data)
	if err != nil {
		return nil, fmt.Errorf("ImageProcessor - Thumbnail - decodeImage: %w", err)
	}

	thumb := imaging.Thumbnail(img, thumbWidth, thumbHeight, imaging.Lanczos)

	res, err := encodeImage(thumb, contentType)
	if err != nil {
		return nil, fmt.Errorf("ImageProcessor - ResiThumbnailze - encodeImage: %w", err)
	}

	return res, nil
}

func (p *ImageProcessor) Watermark(ctx context.Context, contentType string, data []byte, text string) ([]byte, error) {
	img, err := decodeImage(data)
	if err != nil {
		return nil, fmt.Errorf("ImageProcessor - Watermark - decodeImage: %w", err)
	}

	rgba := imaging.Clone(img)

	d := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.White),
		Face: basicfont.Face7x13,
	}

	bounds := rgba.Bounds()
	textWidth := d.MeasureString(text).Round()

	d.Dot = fixed.P(
		bounds.Max.X-textWidth-10,
		bounds.Max.Y-20,
	)

	d.DrawString(text)

	res, err := encodeImage(rgba, contentType)
	if err != nil {
		return nil, fmt.Errorf("ImageProcessor - Watermark - encodeImage: %w", err)
	}

	return res, nil
}

func decodeImage(data []byte) (image.Image, error) {
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("ImageProcessor - decodeImage - imaging.Decode: %w", err)
	}

	return img, nil
}

func encodeImage(img image.Image, contentType string) ([]byte, error) {
	var buf bytes.Buffer
	var format imaging.Format

	switch contentType {
	case "image/jpeg", "image/jpg":
		format = imaging.JPEG
	case "image/png":
		format = imaging.PNG
	case "image/gif":
		format = imaging.GIF
	default:
		format = imaging.JPEG
	}

	err := imaging.Encode(&buf, img, format)
	if err != nil {
		return nil, fmt.Errorf("ImageProcessor - encodeImage - imaging.Encode: %w", err)
	}

	return buf.Bytes(), nil
}
