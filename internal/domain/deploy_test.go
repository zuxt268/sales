package domain

import (
	"fmt"
	"testing"
)

func TestDeploy(t *testing.T) {
	req := DeployRequest{
		Src: Deploy{
			Domain:   "base2.hp-standard.com",
			ServerID: "xb932770",
		},
		Dst: Deploy{
			Domain:   "b028efa8.hp-standard.net",
			ServerID: "xb157298",
		},
	}
	fmt.Println(req.Src.WordpressRootDirectory())
	fmt.Println(req.Src.SecretConfigPath())
	fmt.Println(req.Src.MuPluginDirectory())

	fmt.Println(req.Dst.WordpressRootDirectory())
	fmt.Println(req.Dst.SecretConfigPath())
	fmt.Println(req.Dst.MuPluginDirectory())
}
