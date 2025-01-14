package detection

type DetectionResult struct {
	Type       string  `json:"type"`
	Timestamp  float64 `json:"timestamp"`
	Confidence float64 `json:"confidence"`
	VideoTime  float64 `json:"videoTime"`
}

type Detector interface {
	Detect(imagePath string) (*DetectionResult, error)
}
