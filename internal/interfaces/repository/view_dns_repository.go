package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/external"
	"net/http"
)

type ViewDNSRepository interface {
	GetReverseIP(ctx context.Context, req *external.ReverseIpRequest) (*external.ReverseIpResponse, error)
}

type viewDNSRepository struct {
	baseURL string
}

func NewViewDNSRepository(baseURL string) ViewDNSRepository {
	return &viewDNSRepository{
		baseURL: baseURL,
	}
}

func (r *viewDNSRepository) GetReverseIP(ctx context.Context, params *external.ReverseIpRequest) (*external.ReverseIpResponse, error) {
	url := fmt.Sprintf("%s/reverseip/?host=%s&apikey=%s", r.baseURL, params.Host, params.ApiKey)

	if params.Page != 0 {
		url = fmt.Sprintf("%s&page=%d", url, params.Page)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, domain.WrapExternalAPI("ViewDNS", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, domain.WrapExternalAPI("ViewDNS", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.WrapExternalAPI("ViewDNS", fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	var response external.ReverseIpResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, domain.WrapExternalAPI("ViewDNS", err)
	}

	return &response, nil
}
