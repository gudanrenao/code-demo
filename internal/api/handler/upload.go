package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件上传失败",
		})
		return
	}

	// 检查文件大小
	if file.Size > int64(taskProcessor.Config.Detection.MaxFileSize)*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("文件大小超过限制(%dMB)", taskProcessor.Config.Detection.MaxFileSize),
		})
		return
	}

	// 检查文件类型
	ext := filepath.Ext(file.Filename)
	if !isAllowedFileType(ext) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不支持的文件类型",
		})
		return
	}

	// 生成唯一文件名
	filename := uuid.New().String() + ext
	// 确保文件名安全
	filename = filepath.Clean(filename)
	if strings.Contains(filename, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件名"})
		return
	}
	dst := filepath.Join("uploads", filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存文件失败",
		})
		return
	}

	// 如果是视频文件，创建抽帧任务
	if isVideoFile(ext) {
		c.JSON(http.StatusOK, gin.H{
			"filename": filename,
			"size":     file.Size,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filename": filename,
		"size":     file.Size,
	})
}

func isVideoFile(ext string) bool {
	videoTypes := map[string]bool{
		".mp4": true,
		".mov": true,
		".avi": true,
		".wmv": true,
		".flv": true,
		".mkv": true,
	}
	return videoTypes[ext]
}

func isAllowedFileType(ext string) bool {
	allowedTypes := map[string]bool{
		".mp4":  true,
		".mov":  true,
		".flv":  true,
		".avi":  true,
		".wmv":  true,
		".rmvb": true,
		".ts":   true,
		".3gp":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	return allowedTypes[ext]
}
