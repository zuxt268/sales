package external

type ReverseIpRequest struct {
	Host   string `form:"host"`
	ApiKey string `form:"apikey"`
	Page   int    `form:"page"`
}

type ReverseIpResponse struct {
	Query struct {
		Tool string `json:"tool"`
		Host string `json:"host"`
	} `json:"query"`
	Response struct {
		DomainCount string `json:"domain_count"`
		Domains     []struct {
			Name         string `json:"name"`
			LastResolved string `json:"last_resolved"`
		} `json:"domains"`
	} `json:"response"`
}
