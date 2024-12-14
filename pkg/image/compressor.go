package image

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

const (
	maxWidth  = 800
	maxHeight = 600
	quality   = 75
)


func CompressImage(imageURL string) ([]byte, error) {
	
	resp, err := http.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %v", err)
	}


	format := getImageFormat(imageURL)
	src, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}
	resizedImg := resizeImage(src)
	return compressImage(resizedImg, format)
}


func resizeImage(src image.Image) image.Image {
	srcBounds := src.Bounds()
	srcWidth, srcHeight := srcBounds.Dx(), srcBounds.Dy()

	
	scaleX := float64(maxWidth) / float64(srcWidth)
	scaleY := float64(maxHeight) / float64(srcHeight)
	scale := min(scaleX, scaleY)

	
	if scale >= 1 {
		return src
	}
	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, srcBounds, draw.Over, nil)

	return dst
}

// compresses the image to the specified format
func compressImage(img image.Image, format string) ([]byte, error) {
	var buf bytes.Buffer

	switch strings.ToLower(format) {
	case "png":
		err := png.Encode(&buf, img)
		if err != nil {
			return nil, fmt.Errorf("failed to compress PNG: %v", err)
		}
	case "jpg", "jpeg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, fmt.Errorf("failed to compress JPEG: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	return buf.Bytes(), nil
}

// determines the image format from the URL
func getImageFormat(imageURL string) string {
	ext := strings.ToLower(filepath.Ext(imageURL))
	switch ext {
	case ".png":
		return "png"
	case ".jpg", ".jpeg":
		return "jpg"
	default:
		log.Printf("Unknown image format for URL: %s, defaulting to JPEG", imageURL)
		return "jpg"
	}
}

// get minimum of two float val
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}