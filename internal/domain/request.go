package domain

import "strings"

type GetDomainsRequest struct {
	Limit    *int    `query:"limit"`
	Offset   *int    `query:"offset"`
	Target   *string `query:"target"`
	CanView  *bool   `query:"can_view"`
	IsJapan  *bool   `query:"is_japan"`
	IsSend   *bool   `query:"is_send"`
	OwnerID  *string `query:"owner_id"`
	Industry *string `query:"industry"`
	IsSSL    *bool   `query:"is_ssl"`
	Status   *Status `query:"status"`
	Name     *string `query:"name"`
}

type UpdateDomainRequest struct {
	Name          string  `json:"name" binding:"required"`
	Status        string  `json:"status" binding:"required"`
	Target        *string `json:"target"`
	IsSend        *bool   `json:"is_send"`
	CanView       *bool   `json:"can_view"`
	IsJapan       *bool   `json:"is_japan"`
	Title         *string `json:"title"`
	OwnerID       *string `json:"owner_id"`
	Address       *string `json:"address"`
	Phone         *string `json:"phone"`
	MobilePhone   *string `json:"mobile_phone"`
	LandlinePhone *string `json:"landline_phone"`
	Industry      *string `json:"industry"`
	IsSSL         *bool   `json:"is_ssl"`
	RawPage       *string `json:"raw_page"`
	PageNum       *int    `json:"page_num"`
}

func (r *UpdateDomainRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return WrapValidation("name is required and cannot be empty or whitespace", nil)
	}

	if strings.TrimSpace(r.Status) == "" {
		return WrapValidation("status is required and cannot be empty or whitespace", nil)
	}

	status := Status(r.Status)
	if !IsValidStatus(status) {
		return WrapValidation("invalid status value: must be one of [unknown, initialize, check_view, crawl_comp_info, phone, done]", nil)
	}

	return nil
}

type PostFetchRequest struct {
	Target string `json:"target" binding:"required"`
}

func (r *PostFetchRequest) Validate() error {
	if strings.TrimSpace(r.Target) == "" {
		return WrapValidation("target is required and cannot be empty or whitespace", nil)
	}
	return nil
}

type User struct {
	Email string `json:"email" binding:"required"`
}

type GetTargetsRequest struct {
	Limit  *int `query:"limit"`
	Offset *int `query:"offset"`
}

type UpdateTargetRequest struct {
	IP   *string `json:"ip"`
	Name *string `json:"name"`
}

type CreateTargetRequest struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
}

type GetLogsRequest struct {
	Name     *string `query:"name"`
	Category *string `query:"category"`
	Limit    *int    `query:"limit"`
	Offset   *int    `query:"offset"`
}

type CreateLogRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Message  string `json:"message"`
}

type CreateTaskRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

type UpdateTaskRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *int    `json:"status"`
}
