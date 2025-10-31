package model

import "strings"

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
