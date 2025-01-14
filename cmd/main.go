package main

import (
	"html/template"
	"log"
	"os"

	"videodetection/internal/api"
	"videodetection/internal/api/handler"
	"videodetection/internal/config"
	"videodetection/internal/service/task"
	"videodetection/internal/storage/mysql"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建必要的目录
	dirs := []string{"uploads", "frames"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("创建目录失败 %s: %v", dir, err)
		}
	}

	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化MySQL存储
	storage, err := mysql.New(cfg.MySQL.DSN)
	if err != nil {
		log.Fatalf("初始化MySQL失败: %v", err)
	}

	// 初始化任务处理器
	taskProcessor := task.NewTaskProcessor(cfg, storage)
	handler.SetTaskProcessor(taskProcessor)

	// 创建 Gin 引擎
	r := gin.Default()

	// 注册模板函数
	r.SetFuncMap(template.FuncMap{
		"mul": func(a, b float64) float64 {
			return a * b
		},
	})

	// 设置静态文件路径
	r.Static("/static", "/Users/zhangwen/cursor/code-demo/web/static")
	r.LoadHTMLGlob("/Users/zhangwen/cursor/code-demo/web/templates/*")

	// 设置路由
	api.SetupRoutes(r)

	// 启动服务器
	log.Printf("服务器启动在 %s", cfg.Server.Address)
	if err := r.Run(cfg.Server.Address); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
