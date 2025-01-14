package task

import (
	"context"
	"fmt"

	"videodetection/internal/config"
	"videodetection/internal/model"
	"videodetection/internal/service/detection"
	"videodetection/internal/service/video"
	"videodetection/internal/storage"
)

type TaskProcessor struct {
	Config    *config.Config
	Video     *video.Processor
	detectors map[string]detection.Detector
	Storage   storage.Storage
}

func NewTaskProcessor(cfg *config.Config, storage storage.Storage) *TaskProcessor {
	return &TaskProcessor{
		Config:    cfg,
		Video:     video.NewProcessor(cfg),
		detectors: make(map[string]detection.Detector),
		Storage:   storage,
	}
}

func (p *TaskProcessor) ProcessVideo(taskID string, videoPath string, types []string, params map[string]float64) error {
	// 创建任务记录
	task := &model.Task{
		ID:       taskID,
		Status:   "processing",
		Progress: 0,
		Types:    types,
		Params:   params,
	}
	if err := p.Storage.CreateTask(context.Background(), task); err != nil {
		return fmt.Errorf("创建任务记录失败: %v", err)
	}

	go func() {
		defer func() {
			if err := p.Video.Cleanup(taskID); err != nil {
				fmt.Printf("清理任务文件失败: %v\n", err)
			}
		}()

		// 获取视频时长
		_, err := p.Video.GetVideoDuration(videoPath)
		if err != nil {
			p.updateTaskStatus(taskID, "failed", err)
			return
		}

		// 抽取视频帧
		frames, err := p.Video.ExtractFrames(videoPath, taskID, map[string]interface{}{
			"frame_count": p.Config.Detection.MaxFrames,
		})
		if err != nil {
			p.updateTaskStatus(taskID, "failed", err)
			return
		}

		// 处理每一帧
		results := make(model.DetectionResults, 0)
		for i, frame := range frames {
			for _, detectorType := range types {
				detector := p.getDetector(detectorType, params)
				if detector == nil {
					continue
				}

				result, err := detector.Detect(frame)
				if err != nil {
					fmt.Printf("检测失败 %s: %v\n", frame, err)
					continue
				}

				if result != nil {
					results = append(results, &model.DetectionResult{
						Type:       result.Type,
						Timestamp:  result.Timestamp,
						VideoTime:  result.VideoTime,
						Confidence: result.Confidence,
					})
				}
			}

			// 更新进度
			progress := float64(i+1) / float64(len(frames)) * 100
			p.updateTaskProgress(taskID, progress)
		}

		// 更新任务状态为完成
		ctx := context.Background()
		if err := p.Storage.UpdateTaskStatus(ctx, taskID, model.TaskStatusCompleted); err != nil {
			fmt.Printf("更新任务状态失败: %v\n", err)
			return
		}
		if err := p.Storage.UpdateTaskResults(ctx, taskID, results); err != nil {
			fmt.Printf("更新任务结果失败: %v\n", err)
		}
	}()

	return nil
}

func (p *TaskProcessor) GetTask(taskID string) *model.Task {
	task, err := p.Storage.GetTask(context.Background(), taskID)
	if err != nil {
		fmt.Printf("获取任务失败: %v\n", err)
		return nil
	}
	return task
}

func (p *TaskProcessor) updateTaskStatus(taskID string, status string, _ error) {
	if err := p.Storage.UpdateTaskStatus(context.Background(), taskID, model.TaskStatus(status)); err != nil {
		fmt.Printf("更新任务状态失败: %v\n", err)
	}
}

func (p *TaskProcessor) updateTaskProgress(taskID string, progress float64) {
	if err := p.Storage.UpdateTaskProgress(context.Background(), taskID, progress); err != nil {
		fmt.Printf("更新任务进度失败: %v\n", err)
	}
}

func (p *TaskProcessor) getDetector(detectorType string, params map[string]float64) detection.Detector {
	threshold := p.Config.Detection.DefaultThreshold
	if t, ok := params[detectorType]; ok {
		threshold = t
	}

	switch detectorType {
	case "black_screen":
		return detection.NewBlackScreenDetector(threshold)
	case "blue_screen":
		return detection.NewColorScreenDetector("blue_screen", threshold)
	case "green_screen":
		return detection.NewColorScreenDetector("green_screen", threshold)
	// 添加其他类型的检测器
	default:
		return nil
	}
}

func (p *TaskProcessor) ListTasks(ctx context.Context, limit, offset int) ([]*model.Task, error) {
	return p.Storage.ListTasks(ctx, limit, offset)
}
