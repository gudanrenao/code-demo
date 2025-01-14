package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TaskDetail 渲染任务详情页面
func TaskDetail(c *gin.Context) {
	taskID := c.Param("id")
	task := taskProcessor.GetTask(taskID)
	if task == nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "任务不存在",
		})
		return
	}

	c.HTML(http.StatusOK, "detail.html", gin.H{
		"title": "任务详情",
		"task":  task,
	})
}
