package detection

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"os"
	"path/filepath"
)

type ColorScreenDetector struct {
	threshold float64
	colorType string
	targetHSV HSVRange
}

type HSVRange struct {
	HueMin, HueMax float64
	SatMin, SatMax float64
	ValMin, ValMax float64
}

// 预定义的颜色范围
var (
	BlueScreenHSV = HSVRange{
		HueMin: 0.55, HueMax: 0.7, // 蓝色色调范围
		SatMin: 0.4, SatMax: 1.0, // 饱和度范围
		ValMin: 0.4, ValMax: 1.0, // 亮度范围
	}

	GreenScreenHSV = HSVRange{
		HueMin: 0.25, HueMax: 0.4, // 绿色色调范围
		SatMin: 0.4, SatMax: 1.0,
		ValMin: 0.4, ValMax: 1.0,
	}
)

func NewColorScreenDetector(colorType string, threshold float64) *ColorScreenDetector {
	var targetHSV HSVRange
	switch colorType {
	case "blue_screen":
		targetHSV = BlueScreenHSV
	case "green_screen":
		targetHSV = GreenScreenHSV
	default:
		return nil
	}

	return &ColorScreenDetector{
		threshold: threshold,
		colorType: colorType,
		targetHSV: targetHSV,
	}
}

// RGB转HSV
func rgbToHSV(r, g, b uint8) (h, s, v float64) {
	red := float64(r) / 255
	green := float64(g) / 255
	blue := float64(b) / 255

	max := red
	if green > max {
		max = green
	}
	if blue > max {
		max = blue
	}

	min := red
	if green < min {
		min = green
	}
	if blue < min {
		min = blue
	}

	v = max
	delta := max - min

	if max != 0 {
		s = delta / max
	}

	if delta == 0 {
		h = 0
	} else {
		switch max {
		case red:
			h = (green - blue) / delta
			if green < blue {
				h += 6
			}
		case green:
			h = 2 + (blue-red)/delta
		case blue:
			h = 4 + (red-green)/delta
		}
		h /= 6
	}

	return h, s, v
}

func (d *ColorScreenDetector) Detect(imagePath string) (*DetectionResult, error) {
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

	bounds := img.Bounds()
	matchingPixels := 0
	totalPixels := bounds.Dx() * bounds.Dy()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			h, s, v := rgbToHSV(uint8(r>>8), uint8(g>>8), uint8(b>>8))

			// 检查像素是否在目标HSV范围内
			if h >= d.targetHSV.HueMin && h <= d.targetHSV.HueMax &&
				s >= d.targetHSV.SatMin && s <= d.targetHSV.SatMax &&
				v >= d.targetHSV.ValMin && v <= d.targetHSV.ValMax {
				matchingPixels++
			}
		}
	}

	ratio := float64(matchingPixels) / float64(totalPixels)

	if ratio > d.threshold {
		filename := filepath.Base(imagePath)
		var timestamp float64
		fmt.Sscanf(filename, "frame_%f.jpg", &timestamp)

		return &DetectionResult{
			Type:       d.colorType,
			Timestamp:  timestamp,
			Confidence: ratio,
		}, nil
	}

	return nil, nil
}
