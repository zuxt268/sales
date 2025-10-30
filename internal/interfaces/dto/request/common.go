package request

type Pagination struct {
	Limit  *int `query:"limit"`
	Offset *int `query:"offset"`
}
