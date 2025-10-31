package adapter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/util"
	"golang.org/x/crypto/ssh"
)

type SSHAdapter interface {
	Run(cfg config.SSHConfig, command string) error
	RunOutput(cfg config.SSHConfig, command string) (string, error)
	UploadFile(ctx context.Context, cfg config.SSHConfig, localPath, remotePath string) error
	DownloadFile(ctx context.Context, cfg config.SSHConfig, remotePath, localPath string) error
	WriteFile(cfg config.SSHConfig, content []byte, remotePath string) error
	WriteFileWithPerm(cfg config.SSHConfig, content []byte, remotePath string, perm string) error
}

type sshAdapter struct {
}

func NewSSHAdapter() SSHAdapter {
	return &sshAdapter{}
}

// getSSHClientConfig loads the private key and returns *ssh.ClientConfig
func (a *sshAdapter) getSSHClientConfig(cfg config.SSHConfig) (*ssh.ClientConfig, error) {
	// チルダを展開
	keyPath := util.ExpandTilde(cfg.KeyPath)

	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}

	clientConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         cfg.Timeout,
	}

	return clientConfig, nil
}

// Run executes a shell command on a remote host via SSH
func (a *sshAdapter) Run(cfg config.SSHConfig, command string) error {
	clientConfig, err := a.getSSHClientConfig(cfg)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	fmt.Printf("📡 Running on %s: %s\n", cfg.Host, command)
	err = session.Run(command)

	// 結果を整形して返す
	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	if stdout != "" {
		fmt.Printf("---- STDOUT (%s) ----\n%s\n", cfg.Host, stdout)
	}
	if stderr != "" {
		fmt.Printf("---- STDERR (%s) ----\n%s\n", cfg.Host, stderr)
	}

	if err != nil {
		var exitErr *ssh.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("run failed on %s: exit=%d, stderr=%s",
				cfg.Host, exitErr.ExitStatus(), stderr)
		}
		return fmt.Errorf("run failed on %s: %w", cfg.Host, err)
	}

	return nil
}

func (a *sshAdapter) RunOutput(cfg config.SSHConfig, command string) (string, error) {
	clientConfig, err := a.getSSHClientConfig(cfg)
	if err != nil {
		return "", err
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return "", fmt.Errorf("dial: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	out, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("remote run: %w, output: %s", err, string(out))
	}
	return string(out), nil
}

// UploadFile copies a local file to the remote host via SCP
func (a *sshAdapter) UploadFile(ctx context.Context, cfg config.SSHConfig, localPath, remotePath string) error {
	// チルダを展開
	keyPath := util.ExpandTilde(cfg.KeyPath)

	clientConfig, err := auth.PrivateKey(cfg.User, keyPath, ssh.InsecureIgnoreHostKey())
	if err != nil {
		return fmt.Errorf("scp auth: %w", err)
	}

	client := scp.NewClient(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), &clientConfig)
	err = client.Connect()
	if err != nil {
		return fmt.Errorf("scp connect: %w", err)
	}
	defer client.Close()

	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open local: %w", err)
	}
	defer file.Close()

	fmt.Printf("⬆️ Uploading %s → %s:%s\n", localPath, cfg.Host, remotePath)
	err = client.CopyFile(ctx, file, remotePath, "0644")
	if err != nil {
		return fmt.Errorf("scp upload: %w", err)
	}

	return nil
}

// DownloadFile copies a remote file to the local host via SCP
func (a *sshAdapter) DownloadFile(ctx context.Context, cfg config.SSHConfig, remotePath, localPath string) error {
	// チルダを展開
	keyPath := util.ExpandTilde(cfg.KeyPath)

	clientConfig, err := auth.PrivateKey(cfg.User, keyPath, ssh.InsecureIgnoreHostKey())
	if err != nil {
		return fmt.Errorf("scp auth: %w", err)
	}

	client := scp.NewClient(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), &clientConfig)
	err = client.Connect()
	if err != nil {
		return fmt.Errorf("scp connect: %w", err)
	}
	defer client.Close()

	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("create local: %w", err)
	}
	defer outFile.Close()

	fmt.Printf("⬇️ Downloading %s:%s → %s\n", cfg.Host, remotePath, localPath)
	err = client.CopyFromRemote(ctx, outFile, remotePath)
	if err != nil && err != io.EOF {
		return fmt.Errorf("scp download: %w", err)
	}

	return nil
}

// WriteFile writes content directly to a remote file via SSH stdin
func (a *sshAdapter) WriteFile(cfg config.SSHConfig, content []byte, remotePath string) error {
	return a.WriteFileWithPerm(cfg, content, remotePath, "0644")
}

// WriteFileWithPerm writes content directly to a remote file via SSH stdin with specified permissions
func (a *sshAdapter) WriteFileWithPerm(cfg config.SSHConfig, content []byte, remotePath string, perm string) error {
	clientConfig, err := a.getSSHClientConfig(cfg)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	// stdinにコンテンツを流し込む
	pipe, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	cmd := fmt.Sprintf("cat > %s && chmod %s %s", remotePath, perm, remotePath)
	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	if _, err := pipe.Write(content); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	pipe.Close()

	if err := session.Wait(); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	fmt.Printf("📝 Wrote %s:%s (%d bytes, perm: %s)\n", cfg.Host, remotePath, len(content), perm)
	return nil
}
