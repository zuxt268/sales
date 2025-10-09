package domain

import "time"

type Domain struct {
	ID       int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name     string    `gorm:"column:name;unique" json:"name"`
	CanView  bool      `gorm:"column:can_view" json:"can_view"`
	IsJapan  bool      `gorm:"column:is_japan" json:"is_japan"`
	IsSend   bool      `gorm:"column:is_send" json:"is_send"`
	Title    string    `gorm:"column:title" json:"title"`
	OwnerID  string    `gorm:"column:owner_id" json:"owner_id"`
	Address  string    `gorm:"column:address" json:"address"`
	Phone    string    `gorm:"column:phone" json:"phone"`
	Industry string    `gorm:"column:industry" json:"industry"` // 業種
	IsSSL    bool      `gorm:"column:is_ssl" json:"is_ssl"`
	RawPage  string    `gorm:"column:raw_page" json:"raw_page"`
	PageNum  int       `gorm:"column:page_num" json:"page_num"`
	Status   Status    `gorm:"column:status" json:"status"`
	UpdateAt time.Time `gorm:"column:update_at;autoUpdateTime" json:"update_at"`
	CreateAt time.Time `gorm:"column:create_at;autoCreateTime" json:"create_at"`
}

type Status string

const (
	StatusUnknown       Status = "unknown"
	StatusInitialize    Status = "initialize"
	StatusCheckView     Status = "check_view"
	StatusCheckJapan    Status = "check_japan"
	StatusCrawlCompInfo Status = "crawl_comp_info"
	StatusDone          Status = "done"
)

// ValidStatuses は有効なステータスのリスト
var ValidStatuses = []Status{
	StatusUnknown,
	StatusInitialize,
	StatusCheckView,
	StatusCheckJapan,
	StatusCrawlCompInfo,
	StatusDone,
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

// ConvertToStatus は文字列をStatus型に変換し、バリデーション
func ConvertToStatus(s string) (Status, error) {
	status := Status(s)
	if !IsValidStatus(status) {
		return "", WrapValidation("invalid status value", nil)
	}
	return status, nil
}
