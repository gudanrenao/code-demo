package storage

import (
	"context"

	"videodetection/internal/model"
)

type Storage interface {
	CreateTask(ctx context.Context, task *model.Task) error
	GetTask(ctx context.Context, taskID string) (*model.Task, error)
	UpdateTaskStatus(ctx context.Context, taskID string, status model.TaskStatus) error
	UpdateTaskProgress(ctx context.Context, taskID string, progress float64) error
	UpdateTaskResults(ctx context.Context, taskID string, results model.DetectionResults) error
	ListTasks(ctx context.Context, limit, offset int) ([]*model.Task, error)
}
