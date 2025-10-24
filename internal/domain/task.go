package domain

import "time"

type Task struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	Status      int       `gorm:"column:status" json:"status"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

type TaskStatus int

const (
	TaskStatusUnknown  = -1
	TaskStatusDisabled = 0
	TaskStatusPending  = 1
	TaskStatusRunning  = 2
)

func (t *TaskStatus) ToString() string {
	if t == nil {
		return ""
	}
	switch *t {
	case TaskStatusPending:
		return "pending"
	case TaskStatusRunning:
		return "running"
	case TaskStatusDisabled:
		return "disabled"
	default:
		return ""
	}
}

func StatusStrToTaskStatus(status string) TaskStatus {
	switch status {
	case "pending":
		return TaskStatusPending
	case "running":
		return TaskStatusRunning
	case "disabled":
		return TaskStatusDisabled
	default:
		return TaskStatusUnknown
	}
}
