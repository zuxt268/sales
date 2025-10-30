package domain

import (
	"strings"
	"time"
)

type Domain struct {
	ID            int
	Name          string
	Target        string
	CanView       bool
	IsJapan       bool
	IsSend        bool
	Title         string
	OwnerID       string
	Address       string
	Phone         string
	MobilePhone   string
	LandlinePhone string
	Industry      string
	President     string
	Company       string
	Prefecture    string
	IsSSL         bool
	RawPage       string
	PageNum       int
	Status        Status
	UpdatedAt     time.Time
	CreatedAt     time.Time
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
