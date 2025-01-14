package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

type Task struct {
	ID        string           `json:"id" db:"id"`
	Filename  string           `json:"filename" db:"filename"`
	Types     DetectionTypes   `json:"types" db:"types"`
	Params    DetectionParams  `json:"params" db:"params"`
	Status    TaskStatus       `json:"status" db:"status"`
	Progress  float64          `json:"progress" db:"progress"`
	Results   DetectionResults `json:"results" db:"results"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt time.Time        `json:"updated_at" db:"updated_at"`
}

// 自定义类型用于JSON序列化和反序列化
type DetectionTypes []string
type DetectionParams map[string]float64
type DetectionResult struct {
	Type        string  `json:"type"`
	Timestamp   float64 `json:"timestamp"`
	Confidence  float64 `json:"confidence"`
	VideoTime   float64 `json:"videoTime"` // 视频实际时间点
	ImageFormat string  `json:"imageFormat"`
}

type DetectionResults []*DetectionResult

// Value 实现 driver.Valuer 接口
func (dt DetectionTypes) Value() (driver.Value, error) {
	return json.Marshal(dt)
}

func (dp DetectionParams) Value() (driver.Value, error) {
	return json.Marshal(dp)
}

func (dr DetectionResults) Value() (driver.Value, error) {
	return json.Marshal(dr)
}

// Scan 实现 sql.Scanner 接口
func (dt *DetectionTypes) Scan(value interface{}) error {
	if value == nil {
		*dt = DetectionTypes{}
		return nil
	}
	return json.Unmarshal(value.([]byte), dt)
}

func (dp *DetectionParams) Scan(value interface{}) error {
	if value == nil {
		*dp = DetectionParams{}
		return nil
	}
	return json.Unmarshal(value.([]byte), dp)
}

func (dr *DetectionResults) Scan(value interface{}) error {
	if value == nil {
		*dr = DetectionResults{}
		return nil
	}
	return json.Unmarshal(value.([]byte), dr)
}
