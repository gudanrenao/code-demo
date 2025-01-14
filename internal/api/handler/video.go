package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"videodetection/internal/model"
)

// 抽帧选项
type ExtractOptions struct {
	Interval    *float64 `json:"interval,omitempty"`     // 时间间隔(秒)
	FrameCount  *int     `json:"frame_count,omitempty"`  // 总帧数
	ImageFormat string   `json:"image_format,omitempty"` // 图像格式
}

// 获取视频信息
func GetVideoInfo(c *gin.Context) {
	videoID := c.Param("id")
	videoPath := filepath.Join("uploads", videoID)

	// 检查文件是否存在
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "视频文件不存在",
		})
		return
	}

	info, err := taskProcessor.Video.GetVideoInfo(videoPath)
	if err != nil {
		log.Printf("获取视频信息失败 videoID=%s: %v", videoID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取视频信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

// 创建抽帧任务
func ExtractFrames(c *gin.Context) {
	videoID := c.Param("id")
	videoPath := filepath.Join("uploads", videoID)

	// 检查文件是否存在
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "视频文件不存在"})
		return
	}

	// 创建任务记录
	taskID := uuid.New().String()
	task := &model.Task{
		ID:       taskID,
		Filename: videoID,
		Status:   model.TaskStatusProcessing,
	}
	if err := taskProcessor.Storage.CreateTask(context.Background(), task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败"})
		return
	}

	var options ExtractOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	// 转换选项格式
	opts := make(map[string]interface{})
	if options.Interval != nil {
		opts["interval"] = *options.Interval
	}
	if options.FrameCount != nil {
		opts["frame_count"] = *options.FrameCount
	}
	if options.ImageFormat != "" {
		opts["image_format"] = options.ImageFormat
	}

	// 异步处理
	go func() {
		frames, err := taskProcessor.Video.ExtractFrames(videoPath, taskID, opts)
		if err != nil {
			log.Printf("抽帧失败 taskID=%s: %v", taskID, err)
			taskProcessor.Storage.UpdateTaskStatus(context.Background(), taskID, model.TaskStatusFailed)
			return
		}

		// 从文件名中提取时间信息
		results := make(model.DetectionResults, 0, len(frames))
		for _, frame := range frames {
			var index int
			var videoTime float64
			filename := filepath.Base(frame)
			// 从文件名中提取索引和时间，忽略扩展名
			base := strings.TrimSuffix(filename, filepath.Ext(filename))
			if _, err := fmt.Sscanf(base, "frame_%d_time_%f", &index, &videoTime); err != nil {
				log.Printf("解析帧文件名失败: %v", err)
				continue
			}

			results = append(results, &model.DetectionResult{
				Type:        "frame",
				Timestamp:   float64(index),
				VideoTime:   videoTime,
				ImageFormat: opts["image_format"].(string),
				Confidence:  1.0,
			})
		}

		if err := taskProcessor.Storage.UpdateTaskStatus(context.Background(), taskID, model.TaskStatusCompleted); err != nil {
			log.Printf("更新任务状态失败 taskID=%s: %v", taskID, err)
			return
		}

		if err := taskProcessor.Storage.UpdateTaskResults(context.Background(), taskID, results); err != nil {
			log.Printf("更新任务结果失败 taskID=%s: %v", taskID, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
	})
}

// 获取抽帧结果
func GetVideoFrames(c *gin.Context) {
	videoID := c.Param("id")
	framesDir := filepath.Join("frames", videoID)

	// 获取所有帧文件
	frames, err := filepath.Glob(filepath.Join(framesDir, "frame_*.jpg"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取帧列表失败",
		})
		return
	}

	// 提取每一帧的时间戳
	type Frame struct {
		Path      string  `json:"path"`
		Timestamp float64 `json:"timestamp"`
		VideoTime float64 `json:"videoTime"`
	}

	var frameList []Frame
	for _, frame := range frames {
		var index int
		var videoTime float64
		filename := filepath.Base(frame)
		if _, err := fmt.Sscanf(filename, "frame_%d_time_%f.jpg", &index, &videoTime); err != nil {
			continue
		}
		frameList = append(frameList, Frame{
			Path:      filename,
			Timestamp: float64(index),
			VideoTime: videoTime,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id": videoID,
		"results": frameList,
	})
}
