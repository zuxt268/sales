package request

type GetLogs struct {
	Pagination
	Name     *string `json:"name"`
	Category *string `json:"category"`
}

type CreateLog struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Message  string `json:"message"`
}
