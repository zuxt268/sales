package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/zuxt268/sales/internal/config"
)

type SlackAdapter interface {
	Send(ctx context.Context, msg string) error
}

type slackAdapter struct {
	webhookURL string
}

func NewSlackAdapter() SlackAdapter {
	return &slackAdapter{
		webhookURL: config.Env.NoticeWebAppChannelUrl,
	}
}

func (s *slackAdapter) Send(ctx context.Context, msg string) error {
	payload := map[string]string{
		"text":       msg,
		"username":   "[Sales]",
		"icon_emoji": ":panda_face:",
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}
	return nil
}
