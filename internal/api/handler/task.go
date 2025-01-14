package handler

import (
	"net/http"
	"path/filepath"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"videodetection/internal/service/task"
)

var taskProcessor *task.TaskProcessor

func SetTaskProcessor(p *task.TaskProcessor) {
	taskProcessor = p
}

func CreateTask(c *gin.Context) {
	var req struct {
		Filename string             `json:"filename"`
		Types    []string           `json:"types"`
		Params   map[string]float64 `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 生成任务ID
	taskID := uuid.New().String()

	// 获取视频文件路径
	videoPath := filepath.Join("uploads", req.Filename)

	// 启动处理任务
	if err := taskProcessor.ProcessVideo(taskID, videoPath, req.Types, req.Params); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"status":  "pending",
	})
}

func GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")

	task := taskProcessor.GetTask(taskID)
	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":  task.ID,
		"status":   task.Status,
		"progress": task.Progress,
	})
}

func GetTaskResults(c *gin.Context) {
	taskID := c.Param("id")

	task := taskProcessor.GetTask(taskID)
	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	if task.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务尚未完成"})
		return
	}

	// 按时间戳排序
	sort.Slice(task.Results, func(i, j int) bool {
		return task.Results[i].Timestamp < task.Results[j].Timestamp
	})

	c.JSON(http.StatusOK, gin.H{
		"task_id": task.ID,
		"results": task.Results,
	})
}
