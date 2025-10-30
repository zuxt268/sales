package request

import (
	"fmt"
	"strings"

	"github.com/zuxt268/sales/internal/domain"
)

type GetDomainsRequest struct {
	Pagination
	Name          *string        `query:"name"`
	Target        *string        `query:"target"`
	CanView       *bool          `query:"can_view"`
	IsJapan       *bool          `query:"is_japan"`
	IsSend        *bool          `query:"is_send"`
	Title         *string        `query:"title"`
	OwnerID       *string        `query:"owner_id"`
	MobilePhone   *string        `query:"mobile_phone"`
	LandlinePhone *string        `query:"landline_phone"`
	Industry      *string        `query:"industry"`
	Prefecture    *string        `query:"prefecture"`
	IsSSL         *bool          `query:"is_ssl"`
	Status        *domain.Status `query:"status"`
}

type UpdateDomainRequest struct {
	Name          *string        `json:"name"`
	Target        *string        `json:"target"`
	CanView       *bool          `json:"can_view"`
	IsJapan       *bool          `json:"is_japan"`
	IsSend        *bool          `json:"is_send"`
	Title         *string        `json:"title"`
	OwnerID       *string        `json:"owner_id"`
	Address       *string        `json:"address"`
	Phone         *string        `json:"phone"`
	MobilePhone   *string        `json:"mobile_phone"`
	LandlinePhone *string        `json:"landline_phone"`
	Industry      *string        `json:"industry"`
	President     *string        `json:"president"`
	Company       *string        `json:"company"`
	Prefecture    *string        `json:"prefecture"`
	IsSSL         *bool          `json:"is_ssl"`
	RawPage       *string        `json:"raw_page"`
	PageNum       *int           `json:"page_num"`
	Status        *domain.Status `json:"status"`
}

func (r *UpdateDomainRequest) Validate() error {
	if r.Name != nil || strings.TrimSpace(*r.Name) == "" {
		return fmt.Errorf("name is required and cannot be empty or whitespace")
	}
	if r.Status != nil || strings.TrimSpace(string(*r.Status)) == "" {
		return fmt.Errorf("status is required and cannot be empty or whitespace")
	}
	return nil
}

type CreateDomainRequest struct {
	Name          string        `json:"name"`
	Target        string        `json:"target"`
	CanView       bool          `json:"can_view"`
	IsJapan       bool          `json:"is_japan"`
	IsSend        bool          `json:"is_send"`
	Title         string        `json:"title"`
	OwnerID       string        `json:"owner_id"`
	Address       string        `json:"address"`
	Phone         string        `json:"phone"`
	MobilePhone   string        `json:"mobile_phone"`
	LandlinePhone string        `json:"landline_phone"`
	Industry      string        `json:"industry"`
	President     string        `json:"president"`
	Company       string        `json:"company"`
	Prefecture    string        `json:"prefecture"`
	IsSSL         bool          `json:"is_ssl"`
	RawPage       string        `json:"raw_page"`
	PageNum       int           `json:"page_num"`
	Status        domain.Status `json:"status"`
}
