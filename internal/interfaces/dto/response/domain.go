package response

import (
	"time"

	"github.com/zuxt268/sales/internal/model"
)

type Domain struct {
	ID            int          `json:"id"`
	Name          string       `json:"name"`
	Target        string       `json:"target"`
	CanView       bool         `json:"can_view"`
	IsJapan       bool         `json:"is_japan"`
	IsSend        bool         `json:"is_send"`
	Title         string       `json:"title"`
	OwnerID       string       `json:"owner_id"`
	Address       string       `json:"address"`
	Phone         string       `json:"phone"`
	MobilePhone   string       `json:"mobile_phone"`
	LandlinePhone string       `json:"landline_phone"`
	Industry      string       `json:"industry"`
	President     string       `json:"president"`
	Company       string       `json:"company"`
	Prefecture    string       `json:"prefecture"`
	IsSSL         bool         `json:"is_ssl"`
	RawPage       string       `json:"raw_page"`
	PageNum       int          `json:"page_num"`
	Status        model.Status `json:"status"`
	UpdatedAt     time.Time    `json:"updated_at"`
	CreatedAt     time.Time    `json:"created_at"`
}

type Domains struct {
	Domains []*Domain `json:"domains"`
	Paginate
}

func GetDomain(d *model.Domain) *Domain {
	return &Domain{
		ID:            d.ID,
		Name:          d.Name,
		Target:        d.Target,
		CanView:       d.CanView,
		IsJapan:       d.IsJapan,
		IsSend:        d.IsSend,
		Title:         d.Title,
		OwnerID:       d.OwnerID,
		Address:       d.Address,
		Phone:         d.Phone,
		MobilePhone:   d.MobilePhone,
		LandlinePhone: d.LandlinePhone,
		Industry:      d.Industry,
		President:     d.President,
		Company:       d.Company,
		Prefecture:    d.Prefecture,
		IsSSL:         d.IsSSL,
		RawPage:       d.RawPage,
		PageNum:       d.PageNum,
		Status:        d.Status,
		UpdatedAt:     d.UpdatedAt,
		CreatedAt:     d.CreatedAt,
	}
}

func GetDomains(domains []*model.Domain, total int64) *Domains {
	resDomains := make([]*Domain, 0, len(domains))
	for _, d := range domains {
		resDomains = append(resDomains, GetDomain(d))
	}
	return &Domains{
		Domains: resDomains,
		Paginate: Paginate{
			Total: total,
			Count: len(domains),
		},
	}
}
