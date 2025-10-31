package request

import (
	"fmt"
	"strings"

	"github.com/zuxt268/sales/internal/entity"
)

type DeployRequest struct {
	Src entity.Deploy   `json:"src"`
	Dst []entity.Deploy `json:"dst"`
}

func (r *DeployRequest) Validate() error {
	// ソース側のバリデーション
	if r.Src.Domain == "" {
		return fmt.Errorf("ソースドメインが指定されていません。")
	}
	if r.Src.ServerID == "" {
		return fmt.Errorf("ソースサーバーIDが指定されていません。")
	}
	if strings.Contains(r.Src.Domain, "*") {
		return fmt.Errorf("ソースドメインにワイルドカード（*）を含めることはできません: %s", r.Src.Domain)
	}
	if strings.Contains(r.Src.Domain, "..") {
		return fmt.Errorf("ソースドメインに連続したドット（..）を含めることはできません: %s", r.Src.Domain)
	}

	// デスティネーション側のバリデーション
	for i, d := range r.Dst {
		prefix := fmt.Sprintf("宛先[%d]", i+1)
		if d.Domain == "" {
			return fmt.Errorf("%s のドメインが指定されていません。", prefix)
		}
		if d.ServerID == "" {
			return fmt.Errorf("%s のサーバーIDが指定されていません。", prefix)
		}
		if strings.Contains(d.Domain, "*") {
			return fmt.Errorf("%s のドメインにワイルドカード（*）を含めることはできません: %s", prefix, d.Domain)
		}
		if strings.Contains(d.Domain, "..") {
			return fmt.Errorf("%s のドメインに連続したドット（..）を含めることはできません: %s", prefix, d.Domain)
		}
		if d.Domain == "base2.hp-standard.com" || d.Domain == "hps.hp-standard.xyz" {
			return fmt.Errorf("%s のディレクトリには展開できません。", d.Domain)
		}
	}
	return nil
}
