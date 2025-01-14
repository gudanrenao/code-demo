package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"videodetection/internal/model"

	_ "github.com/go-sql-driver/mysql"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string) (*Storage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return &Storage{db: db}, nil
}

func (s *Storage) CreateTask(ctx context.Context, task *model.Task) error {
	query := `
		INSERT INTO tasks (id, filename, types, params, status, progress, results)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	typesJSON, err := task.Types.Value()
	if err != nil {
		return fmt.Errorf("序列化types失败: %v", err)
	}

	paramsJSON, err := task.Params.Value()
	if err != nil {
		return fmt.Errorf("序列化params失败: %v", err)
	}

	resultsJSON, err := task.Results.Value()
	if err != nil {
		return fmt.Errorf("序列化results失败: %v", err)
	}

	_, err = s.db.ExecContext(ctx, query,
		task.ID,
		task.Filename,
		typesJSON,
		paramsJSON,
		task.Status,
		task.Progress,
		resultsJSON,
	)
	if err != nil {
		return fmt.Errorf("创建任务失败: %v", err)
	}

	return nil
}

func (s *Storage) GetTask(ctx context.Context, taskID string) (*model.Task, error) {
	query := `SELECT * FROM tasks WHERE id = ?`

	var task model.Task
	var typesJSON, paramsJSON, resultsJSON []byte

	err := s.db.QueryRowContext(ctx, query, taskID).Scan(
		&task.ID,
		&task.Filename,
		&typesJSON,
		&paramsJSON,
		&task.Status,
		&task.Progress,
		&resultsJSON,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询任务失败: %v", err)
	}

	// 解析JSON字段
	if err := task.Types.Scan(typesJSON); err != nil {
		return nil, fmt.Errorf("解析types失败: %v", err)
	}
	if err := task.Params.Scan(paramsJSON); err != nil {
		return nil, fmt.Errorf("解析params失败: %v", err)
	}
	if err := task.Results.Scan(resultsJSON); err != nil {
		return nil, fmt.Errorf("解析results失败: %v", err)
	}

	return &task, nil
}

func (s *Storage) UpdateTaskStatus(ctx context.Context, taskID string, status model.TaskStatus) error {
	query := `UPDATE tasks SET status = ? WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, status, taskID)
	if err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	return nil
}

func (s *Storage) UpdateTaskProgress(ctx context.Context, taskID string, progress float64) error {
	query := `UPDATE tasks SET progress = ? WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, progress, taskID)
	if err != nil {
		return fmt.Errorf("更新任务进度失败: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	return nil
}

func (s *Storage) UpdateTaskResults(ctx context.Context, taskID string, results model.DetectionResults) error {
	query := `UPDATE tasks SET results = ? WHERE id = ?`

	resultsJSON, err := results.Value()
	if err != nil {
		return fmt.Errorf("序列化results失败: %v", err)
	}

	result, err := s.db.ExecContext(ctx, query, resultsJSON, taskID)
	if err != nil {
		return fmt.Errorf("更新任务结果失败: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	return nil
}

func (s *Storage) ListTasks(ctx context.Context, limit, offset int) ([]*model.Task, error) {
	query := `
		SELECT * FROM tasks 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询任务列表失败: %v", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		var task model.Task
		var typesJSON, paramsJSON, resultsJSON []byte

		err := rows.Scan(
			&task.ID,
			&task.Filename,
			&typesJSON,
			&paramsJSON,
			&task.Status,
			&task.Progress,
			&resultsJSON,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描任务数据失败: %v", err)
		}

		// 解析JSON字段
		if err := task.Types.Scan(typesJSON); err != nil {
			return nil, fmt.Errorf("解析types失败: %v", err)
		}
		if err := task.Params.Scan(paramsJSON); err != nil {
			return nil, fmt.Errorf("解析params失败: %v", err)
		}
		if err := task.Results.Scan(resultsJSON); err != nil {
			return nil, fmt.Errorf("解析results失败: %v", err)
		}

		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历任务数据失败: %v", err)
	}

	return tasks, nil
}
