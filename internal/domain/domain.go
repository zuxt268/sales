package domain

import (
	"strings"
	"time"
)

type Domain struct {
	ID            int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name          string    `gorm:"column:name;unique" json:"name"`
	Target        string    `gorm:"column:target" json:"target"`
	CanView       bool      `gorm:"column:can_view" json:"can_view"`
	IsJapan       bool      `gorm:"column:is_japan" json:"is_japan"`
	IsSend        bool      `gorm:"column:is_send" json:"is_send"`
	Title         string    `gorm:"column:title" json:"title"`
	OwnerID       string    `gorm:"column:owner_id" json:"owner_id"`
	Address       string    `gorm:"column:address" json:"address"`
	Phone         string    `gorm:"column:phone" json:"phone"`
	MobilePhone   string    `gorm:"column:mobile_phone" json:"mobile_phone"`
	LandlinePhone string    `gorm:"column:landline_phone" json:"landline_phone"`
	Industry      string    `gorm:"column:industry" json:"industry"`
	President     string    `gorm:"column:president" json:"president"`
	Company       string    `gorm:"column:company" json:"company"`
	Prefecture    string    `gorm:"column:prefecture" json:"prefecture"`
	IsSSL         bool      `gorm:"column:is_ssl" json:"is_ssl"`
	RawPage       string    `gorm:"column:raw_page" json:"raw_page"`
	PageNum       int       `gorm:"column:page_num" json:"page_num"`
	Status        Status    `gorm:"column:status" json:"status"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (u *Domain) SetPhone() {
	mobile := make([]string, 0)
	landline := make([]string, 0)
	for _, phone := range strings.Split(u.Phone, ",") {
		phone = strings.TrimSpace(phone)
		if phone == "" {
			continue
		}
		if strings.HasPrefix(phone, "080") || strings.HasPrefix(phone, "090") || strings.HasPrefix(phone, "070") {
			mobile = append(mobile, phone)
		} else {
			landline = append(landline, phone)
		}
	}
	u.MobilePhone = strings.Join(mobile, ",")
	u.LandlinePhone = strings.Join(landline, ",")
}

type Status string

const (
	StatusUnknown       Status = "unknown"
	StatusInitialize    Status = "initialize"
	StatusCheckView     Status = "check_view"
	StatusCheckJapan    Status = "check_japan"
	StatusCrawlCompInfo Status = "crawl_comp_info"
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

// ConvertToStatus は文字列をStatus型に変換し、バリデーション
func ConvertToStatus(s string) (Status, error) {
	status := Status(s)
	if !IsValidStatus(status) {
		return "", WrapValidation("invalid status value", nil)
	}
	return status, nil
}
