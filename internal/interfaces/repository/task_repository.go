package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/zuxt268/sales/internal/entity"
	"github.com/zuxt268/sales/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TaskRepository interface {
	Exists(ctx context.Context, f TaskFilter) (bool, error)
	Get(ctx context.Context, f TaskFilter) (model.Task, error)
	GetForUpdate(ctx context.Context, f TaskFilter) (model.Task, error)
	FindAll(ctx context.Context, f TaskFilter) ([]model.Task, error)
	Save(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, f TaskFilter) error
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{
		db: db,
	}
}

func (r *taskRepository) Exists(ctx context.Context, f TaskFilter) (bool, error) {
	var tasks []model.Task
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Find(&tasks).Error
	if err != nil {
		return false, fmt.Errorf("failed to fetch tasks: %w", err)
	}
	return len(tasks) > 0, nil
}

func (r *taskRepository) Get(ctx context.Context, f TaskFilter) (model.Task, error) {
	t := model.Task{}
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return t, entity.WrapNotFound("task")
		}
		return t, fmt.Errorf("failed to get task: %w", err)
	}
	return t, nil
}

func (r *taskRepository) GetForUpdate(ctx context.Context, f TaskFilter) (model.Task, error) {
	t := model.Task{}
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return t, entity.WrapNotFound("task")
		}
		return t, fmt.Errorf("failed to get task for update: %w", err)
	}
	return t, nil
}

func (r *taskRepository) FindAll(ctx context.Context, f TaskFilter) ([]model.Task, error) {
	var tasks []model.Task
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Find(&tasks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %w", err)
	}
	return tasks, nil
}

func (r *taskRepository) Save(ctx context.Context, task *model.Task) error {
	err := r.getDb(ctx).Save(task).Error
	if err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}
	return nil
}

func (r *taskRepository) Delete(ctx context.Context, f TaskFilter) error {
	err := f.Apply(r.db.WithContext(ctx)).Delete(&model.Task{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (r *taskRepository) getDb(ctx context.Context) *gorm.DB {
	if v, ok := ctx.Value(TxKey{}).(*gorm.DB); ok {
		return v.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

type TaskFilter struct {
	ID     *int
	Name   *string
	Status *int
	Limit  *int
	Offset *int
}

func (t *TaskFilter) Apply(db *gorm.DB) *gorm.DB {
	if t.ID != nil {
		db = db.Where("id = ?", *t.ID)
	}
	if t.Name != nil {
		db = db.Where("name = ?", *t.Name)
	}
	if t.Status != nil {
		db = db.Where("status = ?", *t.Status)
	}
	if t.Limit != nil {
		db = db.Limit(*t.Limit)
		if t.Offset != nil {
			db = db.Offset(*t.Offset)
		}
	}
	return db
}
