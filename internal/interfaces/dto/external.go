package dto

import (
	"github.com/zuxt268/sales/internal/model"
)

var Header = []interface{}{
	"ドメイン",
	"タイトル",
	"owner_id",
	"携帯電話",
	"固定電話",
	"業種",
	"代表者",
	"企業名",
	"都道府県",
	"ページ数",
}

type Row struct {
	Columns []interface{}
}

func GetRows(domains []*model.Domain) []Row {
	rows := make([]Row, len(domains))
	for i, d := range domains {
		rows[i] = Row{
			Columns: []interface{}{
				d.Target,
				d.Title,
				d.OwnerID,
				d.MobilePhone,
				d.LandlinePhone,
				d.Industry,
				d.Company,
				d.Prefecture,
				d.PageNum,
			},
		}
	}
	return rows
}
