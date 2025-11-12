package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zuxt268/sales/internal/interfaces/dto/external"
)

type ViewDNSAdapter interface {
	GetReverseIP(ctx context.Context, req *external.ReverseIpRequest) (*external.ReverseIpResponse, error)
}

type viewDNSAdapter struct {
	baseURL string
}

func NewViewDNSAdapter(baseURL string) ViewDNSAdapter {
	return &viewDNSAdapter{
		baseURL: baseURL,
	}
}

func (r *viewDNSAdapter) GetReverseIP(ctx context.Context, params *external.ReverseIpRequest) (*external.ReverseIpResponse, error) {
	url := fmt.Sprintf("%s/reverseip/?host=%s&apikey=%s", r.baseURL, params.Host, params.ApiKey)

	if params.Page != 0 {
		url = fmt.Sprintf("%s&page=%d", url, params.Page)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch reverse ip address: %s", resp.Status)
	}

	var response external.ReverseIpResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode reverse ip response: %w", err)
	}

	return &response, nil
}
