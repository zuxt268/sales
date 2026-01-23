package model

type Wix struct {
	Name    string `gorm:"column:name"`
	OwnerID string `gorm:"column:owner_id"`
}

func (Wix) TableName() string {
	return "wixes"
}