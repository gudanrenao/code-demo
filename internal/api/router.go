package api

import (
	"videodetection/internal/api/handler"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// 首页
	r.GET("/", handler.HomePage)
	r.GET("/tasks", handler.ListTasks)
	r.GET("/tasks/:id", handler.TaskDetail)

	// 文件上传相关
	r.POST("/api/upload", handler.UploadFile)
	r.GET("/api/files/:filename", handler.GetFile)

	// 检测任务相关
	r.POST("/api/tasks", handler.CreateTask)
	r.GET("/api/tasks/:id", handler.GetTaskStatus)
	r.GET("/api/tasks/:id/results", handler.GetTaskResults)
	r.GET("/api/tasks", handler.GetTaskList)

	// 图片访问
	r.GET("/api/frames/:taskId/:filename", handler.GetFrame)

	// 视频抽帧相关
	r.GET("/api/video/:id/info", handler.GetVideoInfo)
	r.GET("/api/video/:id/frames", handler.GetVideoFrames)
	r.POST("/api/video/:id/extract", handler.ExtractFrames)
}
