package entity

type DomainDetails struct {
	Path        string `json:"path"`
	Server      string `json:"server"`
	Description string `json:"description"`
	SiteUrl     string `json:"siteUrl"`
	BlogName    string `json:"blogName"`
	Users       string `json:"users"`
	DBUsage     string `json:"dbUsage"`
	DiscUsage   string `json:"discUsage"`
	MailUsage   string `json:"mailUsage"`
}
