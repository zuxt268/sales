package response

type Paginate struct {
	Total int64 `json:"total"`
	Count int   `json:"count"`
}
