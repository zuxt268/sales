package model

import (
	"strconv"
	"strings"
	"time"
)

type Homsta struct {
	ID          int        `gorm:"column:id;primaryKey;autoIncrement"`
	Domain      string     `gorm:"column:domain"`
	BlogName    string     `gorm:"column:blog_name"`
	Path        string     `gorm:"column:path"`
	SiteURL     string     `gorm:"column:site_url"`
	Description string     `gorm:"column:description"`
	DBName      string     `gorm:"column:db_name"`
	Users       string     `gorm:"column:users"`
	DBUsage     string     `gorm:"column:db_usage"`
	DiscUsage   string     `gorm:"column:disc_usage"`
	Industry    string     `gorm:"column:industry"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;autoUpdateTime"`
	CreatedAt   *time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Homsta) TableName() string {
	return "homstas"
}

func (h Homsta) GetDbUsage() int {
	rate := 1
	if strings.HasSuffix(h.DBUsage, "GB") {
		rate = 1000
	}
	numStr := strings.ReplaceAll(h.DBUsage, "GB", "")
	numStr = strings.ReplaceAll(numStr, "MB", "")
	num, _ := strconv.Atoi(numStr)
	return num * rate
}

func (h Homsta) GetDiscUsage() int {
	num, _ := strconv.Atoi(h.DiscUsage)
	return num
}
