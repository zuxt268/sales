package domain

import "time"

type Target struct {
	ID        int          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	IP        string       `gorm:"column:ip;unique" json:"ip"`
	Name      string       `gorm:"column:name" json:"name"`
	Status    TargetStatus `gorm:"column:status" json:"status"`
	UpdatedAt time.Time    `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedAt time.Time    `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

type TargetStatus string

const (
	TargetStatusInit    TargetStatus = "init"
	TargetStatusFetched TargetStatus = "fetched"
)

// ValidTargetStatuses は有効なステータスのリスト
var ValidTargetStatuses = []TargetStatus{
	TargetStatusInit,
}

// IsValidTargetStatus は指定されたステータスが有効かどうかを判定
func IsValidTargetStatus(s TargetStatus) bool {
	for _, valid := range ValidTargetStatuses {
		if s == valid {
			return true
		}
	}
	return false
}
