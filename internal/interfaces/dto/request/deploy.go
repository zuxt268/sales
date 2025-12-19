package request

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/zuxt268/sales/internal/entity"
)

type DeployRequest struct {
	Src entity.Deploy   `json:"src"`
	Dst []entity.Deploy `json:"dst"`
}

type DeployOneRequest struct {
	Src entity.Deploy `json:"src"`
	Dst entity.Deploy `json:"dst"`
}

var (
	reDomain = regexp.MustCompile(`\A[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*\z`)
	reServer = regexp.MustCompile(`\A[a-zA-Z0-9._-]+\z`)

	forbiddenDomains = map[string]struct{}{
		"base2.hp-standard.com": {},
		"hps.hp-standard.xyz":   {},
	}
)

func normalizeDomain(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimSuffix(s, ".")
	return s
}

func validateDomain(domain, label string) (string, error) {
	d := normalizeDomain(domain)
	if d == "" {
		return "", fmt.Errorf("%sドメインが指定されていません。", label)
	}
	// まず危険文字を早期排除（コマンド/パス混入対策）
	if strings.ContainsAny(d, " \t\r\n/\\;|&$`'\"<>") {
		return "", fmt.Errorf("%sドメインに使用できない文字が含まれています: %q", label, domain)
	}
	if strings.Contains(d, "*") || strings.Contains(d, "..") {
		return "", fmt.Errorf("%sドメインが不正です: %q", label, domain)
	}
	if len(d) > 253 {
		return "", fmt.Errorf("%sドメインが長すぎます: %q", label, domain)
	}
	if !reDomain.MatchString(d) {
		return "", fmt.Errorf("%sドメイン形式が不正です: %q", label, domain)
	}

	return d, nil
}

func validateServerID(id, label string) (string, error) {
	s := strings.TrimSpace(id)
	if s == "" {
		return "", fmt.Errorf("%sサーバーIDが指定されていません。", label)
	}
	if strings.ContainsAny(s, " \t\r\n/\\") || !reServer.MatchString(s) {
		return "", fmt.Errorf("%sサーバーID形式が不正です: %q", label, id)
	}
	return s, nil
}

func (r *DeployRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("request is nil")
	}
	if len(r.Dst) == 0 {
		return fmt.Errorf("宛先が1件も指定されていません。")
	}

	srcDomain, err := validateDomain(r.Src.Domain, "ソース")
	if err != nil {
		return err
	}
	srcServer, err := validateServerID(r.Src.ServerID, "ソース")
	if err != nil {
		return err
	}

	seen := make(map[string]int, len(r.Dst))

	for i, d := range r.Dst {
		prefix := fmt.Sprintf("宛先[%d]", i+1)

		if _, ng := forbiddenDomains[d.Domain]; ng {
			return fmt.Errorf("%s のディレクトリには展開できません。", d)
		}

		dstDomain, err := validateDomain(d.Domain, prefix)
		if err != nil {
			return err
		}
		dstServer, err := validateServerID(d.ServerID, prefix)
		if err != nil {
			return err
		}

		// src=dst 誤爆防止
		if dstDomain == srcDomain && dstServer == srcServer {
			return fmt.Errorf("%s がソースと同一です（domain/server）。", prefix)
		}

		// 重複防止（同じ宛先を複数回）
		key := dstServer + ":" + dstDomain
		if prev, ok := seen[key]; ok {
			return fmt.Errorf("%s が重複しています（宛先[%d] と同一）。", prefix, prev)
		}
		seen[key] = i + 1
	}

	return nil
}

func (r *DeployOneRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("request is nil")
	}

	srcDomain, err := validateDomain(r.Src.Domain, "ソース")
	if err != nil {
		return err
	}

	srcServer, err := validateServerID(r.Src.ServerID, "ソース")
	if err != nil {
		return err
	}

	dstDomain, err := validateDomain(r.Dst.Domain, "デスティネーション")
	if err != nil {
		return err
	}
	dstServer, err := validateServerID(r.Dst.ServerID, "デスティネーション")
	if err != nil {
		return err
	}

	// src=dst 誤爆防止
	if dstDomain == srcDomain && dstServer == srcServer {
		return fmt.Errorf("デスティネーションとソースと同一です（domain/server）。")
	}

	return nil
}
