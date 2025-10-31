package entity

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/zuxt268/sales/internal/config"
)

type Deploy struct {
	Domain   string `json:"domain"`
	ServerID string `json:"server_id"`
}

func (d *Deploy) WordpressRootDirectory() string {
	term := strings.Split(d.Domain, ".")
	if len(term) > 2 {
		term = term[1:]
		return fmt.Sprintf("/home/%s/%s/public_html/%s",
			d.ServerID,
			strings.Join(term, "."),
			d.Domain,
		)
	}
	return fmt.Sprintf("/home/%s/%s/public_html", d.ServerID, d.Domain)
}

func (d *Deploy) SecretConfigPath() string {
	term := strings.Split(d.Domain, ".")
	if len(term) > 2 {
		term = term[1:]
		return fmt.Sprintf("/home/%s/%s/public_html/%s/wp-content/secret-config.php",
			d.ServerID,
			strings.Join(term, "."),
			d.Domain,
		)
	}
	return fmt.Sprintf("/home/%s/%s/public_html/wp-content/secret-config.php", d.ServerID, d.Domain)
}

func (d *Deploy) MuPluginDirectory() string {
	term := strings.Split(d.Domain, ".")
	if len(term) > 2 {
		term = term[1:]
		return fmt.Sprintf("/home/%s/%s/public_html/%s/wp-content/mu-plugins",
			d.ServerID,
			strings.Join(term, "."),
			d.Domain,
		)
	}
	return fmt.Sprintf("/home/%s/%s/public_html/wp-content/mu-plugins", d.ServerID, d.Domain)
}

func (d *Deploy) GetDbName() string {
	name := strings.Split(d.Domain, ".")[0]
	if len(name) > 16 {
		name = name[:16]
	}

	// 英数字以外を削除
	result := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result += string(c)
		}
	}

	if d.IsSubDomain() {
		return fmt.Sprintf("%s_%spre", d.ServerID, result)
	}

	return fmt.Sprintf("%s_%s", d.ServerID, result)
}

func (d *Deploy) GetDbUser() string {
	return fmt.Sprintf("%s_homsta", d.ServerID)
}

func (d *Deploy) GetDbPassword() string {
	if d.ServerID == "xb932770" {
		return config.Env.DatabasePassword1
	}
	return config.Env.DatabasePassword2
}

func (d *Deploy) GetDbHost() string {
	if d.ServerID == "xb932770" {
		return config.Env.DatabaseHost1
	}
	return config.Env.DatabaseHost2
}

func (d *Deploy) IsSubDomain() bool {
	term := strings.Split(d.Domain, ".")
	return len(term) > 2
}

func (d *Deploy) GetHashData() string {
	modStr := fmt.Sprintf("%s%s", d.Domain, config.Env.HashPhrase)

	// 1回目のハッシュ化
	firstHash := sha256.Sum256([]byte(modStr))
	firstHashStr := fmt.Sprintf("%x", firstHash)
	fmt.Printf("First hash: %s\n", firstHashStr)

	// 2回目のハッシュ化
	finalHash := sha256.Sum256([]byte(firstHashStr))
	finalHashStr := fmt.Sprintf("%x", finalHash)
	fmt.Printf("Final hash: %s\n", finalHashStr)
	return finalHashStr
}

type DeployResult struct {
	Success []string `json:"success"`
	Failed  []string `json:"failed"`
}
