package domain

import "time"

type Target struct {
	ID       int          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	IP       string       `gorm:"column:ip;unique" json:"ip"`
	Name     string       `gorm:"column:name" json:"name"`
	Status   TargetStatus `gorm:"column:status" json:"status"`
	UpdateAt time.Time    `gorm:"column:update_at;autoUpdateTime" json:"update_at"`
	CreateAt time.Time    `gorm:"column:create_at;autoCreateTime" json:"create_at"`
}

type TargetStatus string

const (
	TargetStatusInit TargetStatus = "init"
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
