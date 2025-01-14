package handler

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func GetFile(c *gin.Context) {
    filename := c.Param("filename")
    filepath := filepath.Join("uploads", filename)
    
    c.File(filepath)
}

func GetFrame(c *gin.Context) {
    taskID := c.Param("taskId")
    filename := c.Param("filename")
    filepath := filepath.Join("frames", taskID, filename)
    
    c.File(filepath)
} 