package external

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

func GetRows(domains []*model.Domain) [][]interface{} {
	rows := make([][]interface{}, len(domains))
	for i, d := range domains {
		rows[i] = []interface{}{
			d.Name,
			d.Title,
			d.OwnerID,
			d.MobilePhone,
			d.LandlinePhone,
			d.Industry,
			d.President,
			d.Company,
			d.Prefecture,
			d.PageNum,
		}
	}
	return rows
}
