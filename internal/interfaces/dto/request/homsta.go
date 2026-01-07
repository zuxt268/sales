package request

type Homsta struct {
	Path        string `json:"path"`
	Description string `json:"description"`
	SiteUrl     string `json:"siteUrl"`
	BlogName    string `json:"blogName"`
	Users       string `json:"users"`
	DbUsage     string `json:"dbUsage"`
	DiscUsage   string `json:"discUsage"`
}
