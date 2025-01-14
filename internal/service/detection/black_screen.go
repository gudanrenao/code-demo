package detection

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"os"
	"path/filepath"
)

type BlackScreenDetector struct {
	threshold float64
}

func NewBlackScreenDetector(threshold float64) *BlackScreenDetector {
	return &BlackScreenDetector{
		threshold: threshold,
	}
}

func (d *BlackScreenDetector) Detect(imagePath string) (*DetectionResult, error) {
	// 读取图片
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开图片: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("无法解码图片: %v", err)
	}

	// 计算平均亮度
	bounds := img.Bounds()
	var totalBrightness float64
	pixels := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			totalBrightness += float64(c.Y)
			pixels++
		}
	}

	brightness := totalBrightness / float64(pixels) / 255.0

	// 如果平均亮度低于阈值，认为是黑屏
	if brightness < d.threshold {
		// 从文件名中提取时间戳
		filename := filepath.Base(imagePath)
		var timestamp float64
		fmt.Sscanf(filename, "frame_%f.jpg", &timestamp)

		return &DetectionResult{
			Type:       "black_screen",
			Timestamp:  timestamp,
			Confidence: 1.0 - (brightness / d.threshold),
		}, nil
	}

	return nil, nil
}
