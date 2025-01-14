package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListTasks 渲染任务列表页面
func ListTasks(c *gin.Context) {
	c.HTML(http.StatusOK, "tasks.html", gin.H{
		"title": "检测任务列表",
	})
}

// GetTaskList 获取任务列表API
func GetTaskList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	tasks, err := taskProcessor.ListTasks(c, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取任务列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"pagination": gin.H{
			"page":     page,
			"pageSize": pageSize,
		},
	})
}
