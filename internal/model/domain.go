package model

import (
	"time"
)

type Domain struct {
	ID            int       `gorm:"column:id;primaryKey;autoIncrement"`
	Name          string    `gorm:"column:name;unique"`
	Target        string    `gorm:"column:target"`
	CanView       bool      `gorm:"column:can_view"`
	IsJapan       bool      `gorm:"column:is_japan"`
	IsSend        bool      `gorm:"column:is_send"`
	Title         string    `gorm:"column:title"`
	OwnerID       string    `gorm:"column:owner_id"`
	Address       string    `gorm:"column:address"`
	Phone         string    `gorm:"column:phone"`
	MobilePhone   string    `gorm:"column:mobile_phone"`
	LandlinePhone string    `gorm:"column:landline_phone"`
	Industry      string    `gorm:"column:industry"`
	President     string    `gorm:"column:president"`
	Company       string    `gorm:"column:company"`
	Prefecture    string    `gorm:"column:prefecture"`
	IsSSL         bool      `gorm:"column:is_ssl"`
	RawPage       string    `gorm:"column:raw_page"`
	PageNum       int       `gorm:"column:page_num"`
	Status        Status    `gorm:"column:status"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
}

type Status string

const (
	StatusUnknown       Status = "unknown"
	StatusInitialize    Status = "initialize"
	StatusCheckView     Status = "check_view"
	StatusCheckJapan    Status = "check_japan"
	StatusCrawlCompInfo Status = "crawl_comp_info"
	StatusPendingOutput Status = "pending_output"
	StatusDone          Status = "done"
	StatusTrash         Status = "trash"
)

// ValidStatuses は有効なステータスのリスト
var ValidStatuses = []Status{
	StatusUnknown,
	StatusInitialize,
	StatusCheckView,
	StatusCheckJapan,
	StatusCrawlCompInfo,
	StatusDone,
	StatusTrash,
}

// IsValidStatus は指定されたステータスが有効かどうかを判定
func IsValidStatus(s Status) bool {
	for _, valid := range ValidStatuses {
		if s == valid {
			return true
		}
	}
	return false
}
