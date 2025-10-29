package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kevinburke/ssh_config"
	"github.com/zuxt268/sales/internal/domain"
)

var cfg *ssh_config.Config

func init() {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".ssh", "config")

	file, err := os.Open(configPath)
	if err != nil {
		fmt.Printf("Failed to open config file: %s\n", err)
		return
	}
	defer file.Close()

	cfg, err = ssh_config.Decode(file)
	if err != nil {
		panic(err)
	}
}

func GetSSHConfig(serverID string) (domain.SSHConfig, error) {
	if cfg == nil {
		return domain.SSHConfig{}, fmt.Errorf("SSH config not loaded (config file not found)")
	}
	hostname, err := cfg.Get(serverID, "Hostname")
	if err != nil {
		return domain.SSHConfig{}, err
	}
	user, err := cfg.Get(serverID, "User")
	if err != nil {
		return domain.SSHConfig{}, err
	}
	port := 22
	if p, err := cfg.Get(serverID, "Port"); err == nil && p != "" {
		_, _ = fmt.Sscanf(p, "%d", &port)
	}
	identity, err := cfg.Get(serverID, "IdentityFile")
	if err != nil {
		return domain.SSHConfig{}, err
	}
	return domain.SSHConfig{
		Host:    hostname,
		User:    user,
		Port:    port,
		KeyPath: identity,
		Timeout: 60 * time.Second,
	}, nil
}
