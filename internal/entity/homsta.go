package entity

type HomstaDetails struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SiteURL     string `json:"siteUrl"`
	Description string `json:"description"`
	DBName      string `json:"dbName"`
	Users       string `json:"users"`
	DBUsage     string `json:"dbUsage"`
	DiscUsage   string `json:"discUsage"`
	Industry    string `json:"industry"`
}