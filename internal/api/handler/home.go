package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func HomePage(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", gin.H{
        "title": "视频检测系统",
    })
} 