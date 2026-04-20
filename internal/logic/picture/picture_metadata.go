package picture

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"

	commonresponse "photo-album/internal/common/response"
)

func extractPictureMetadata(filePath, originalFilename string) (pictureMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return pictureMetadata{}, commonresponse.InternalServerError("读取图片失败")
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return pictureMetadata{}, commonresponse.InternalServerError("读取图片信息失败")
	}

	header := make([]byte, 64)
	n, _ := file.Read(header)
	header = header[:n]

	format := detectPictureFormat(header, originalFilename)
	if format == "" {
		return pictureMetadata{}, commonresponse.BadRequest("仅支持 jpg、jpeg、png、webp 图片")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return pictureMetadata{}, commonresponse.InternalServerError("读取图片失败")
	}

	metadata := pictureMetadata{
		Size:   info.Size(),
		Format: format,
	}

	switch format {
	case "jpg", "jpeg", "png":
		cfg, _, err := image.DecodeConfig(file)
		if err != nil {
			return pictureMetadata{}, commonresponse.BadRequest("无法解析图片尺寸")
		}
		metadata.Width = int64(cfg.Width)
		metadata.Height = int64(cfg.Height)
		if metadata.Height > 0 {
			metadata.Scale = float64(metadata.Width) / float64(metadata.Height)
		}

		if colorValue, err := extractDominantColor(filePath); err == nil {
			metadata.DominantColor = colorValue
		}
	case "webp":
		width, height, err := extractWebPDimensions(header)
		if err != nil {
			return pictureMetadata{}, commonresponse.BadRequest("无法解析 webp 图片尺寸")
		}
		metadata.Width = width
		metadata.Height = height
		if metadata.Height > 0 {
			metadata.Scale = float64(metadata.Width) / float64(metadata.Height)
		}
	}

	return metadata, nil
}

func detectPictureFormat(header []byte, originalFilename string) string {
	switch {
	case len(header) >= 3 && bytes.Equal(header[:3], []byte{0xFF, 0xD8, 0xFF}):
		return "jpg"
	case len(header) >= 8 && bytes.Equal(header[:8], []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}):
		return "png"
	case len(header) >= 12 && bytes.Equal(header[:4], []byte("RIFF")) && bytes.Equal(header[8:12], []byte("WEBP")):
		return "webp"
	default:
		ext := normalizeExtension(originalFilename)
		if ext == "jpeg" {
			return "jpeg"
		}
		if ext == "jpg" || ext == "png" || ext == "webp" {
			return ext
		}
		return ""
	}
}

func extractDominantColor(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width == 0 || height == 0 {
		return "", errors.New("invalid image bounds")
	}

	stepX := maxInt(width/32, 1)
	stepY := maxInt(height/32, 1)

	var totalR, totalG, totalB, count uint64
	for y := bounds.Min.Y; y < bounds.Max.Y; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			r, g, b, _ := img.At(x, y).RGBA()
			totalR += uint64(r >> 8)
			totalG += uint64(g >> 8)
			totalB += uint64(b >> 8)
			count++
		}
	}

	if count == 0 {
		return "", errors.New("empty color sample")
	}

	return fmt.Sprintf("#%02X%02X%02X", totalR/count, totalG/count, totalB/count), nil
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
