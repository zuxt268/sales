package usecase

import (
	"context"

	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/util"
)

type TaskUsecase interface {
	ExecuteTask(ctx context.Context, id int) (*domain.Task, error)
	ExecuteTasks(ctx context.Context) error
	GetTasks(ctx context.Context) ([]domain.Task, error)
	CreateTask(ctx context.Context, req *domain.CreateTaskRequest) (*domain.Task, error)
	UpdateTask(ctx context.Context, id int, req *domain.UpdateTaskRequest) (*domain.Task, error)
	DeleteTask(ctx context.Context, id int) error
}

type taskUsecase struct {
	baseRepo         repository.BaseRepository
	taskRepo         repository.TaskRepository
	taskQueueAdapter adapter.TaskQueueAdapter
}

func NewTaskUsecase(
	baseRepo repository.BaseRepository,
	taskRepo repository.TaskRepository,
	taskQueueAdapter adapter.TaskQueueAdapter,
) TaskUsecase {
	return &taskUsecase{
		baseRepo:         baseRepo,
		taskRepo:         taskRepo,
		taskQueueAdapter: taskQueueAdapter,
	}
}

func (u *taskUsecase) GetTasks(ctx context.Context) ([]domain.Task, error) {
	return u.taskRepo.FindAll(ctx, repository.TaskFilter{})
}

func (u *taskUsecase) ExecuteTask(ctx context.Context, id int) (*domain.Task, error) {

	var task domain.Task
	err := u.baseRepo.WithTransaction(ctx, func(ctx context.Context) error {
		var err error
		task, err = u.taskRepo.GetForUpdate(ctx, repository.TaskFilter{
			ID: &id,
		})
		if err != nil {
			return err
		}
		task.Status = domain.TaskStatusRunning
		err = u.taskQueueAdapter.Enqueue(ctx, task)
		if err != nil {
			return err
		}
		err = u.taskRepo.Save(ctx, &task)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (u *taskUsecase) ExecuteTasks(ctx context.Context) error {
	tasks, err := u.taskRepo.FindAll(ctx, repository.TaskFilter{
		Status: util.Pointer(domain.TaskStatusPending),
	})
	if err != nil {
		return err
	}
	for _, task := range tasks {
		err := u.baseRepo.WithTransaction(ctx, func(ctx context.Context) error {
			if err := u.taskQueueAdapter.Enqueue(ctx, task); err != nil {
				return err
			}
			task.Status = domain.TaskStatusRunning
			if err := u.taskRepo.Save(ctx, &task); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *taskUsecase) CreateTask(ctx context.Context, req *domain.CreateTaskRequest) (*domain.Task, error) {

	task := &domain.Task{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	}
	err := u.taskRepo.Save(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (u *taskUsecase) UpdateTask(ctx context.Context, id int, req *domain.UpdateTaskRequest) (*domain.Task, error) {
	before, err := u.taskRepo.GetForUpdate(ctx, repository.TaskFilter{
		ID: util.Pointer(id),
	})
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		before.Name = *req.Name
	}
	if req.Description != nil {
		before.Description = *req.Description
	}
	if req.Status != nil {
		before.Status = *req.Status
	}
	err = u.taskRepo.Save(ctx, &before)
	if err != nil {
		return nil, err
	}
	return &before, nil
}

func (u *taskUsecase) DeleteTask(ctx context.Context, id int) error {
	return u.taskRepo.Delete(ctx, repository.TaskFilter{
		ID: util.Pointer(id),
	})
}
