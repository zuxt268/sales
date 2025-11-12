package adapter

import (
	"context"
	"encoding/json"

	"github.com/zuxt268/sales/internal/infrastructure"
	"github.com/zuxt268/sales/internal/interfaces/dto/external"
)

type PubSubAdapter interface {
	PushDomain(ctx context.Context, domain *external.DomainMessage) error
}

type pubSubAdapter struct {
	pubSubClient infrastructure.PubSubClient
}

func NewPubSubAdapter(
	pubSubClient infrastructure.PubSubClient,
) PubSubAdapter {
	return &pubSubAdapter{
		pubSubClient: pubSubClient,
	}
}

func (a *pubSubAdapter) PushDomain(ctx context.Context, domain *external.DomainMessage) error {
	jsonData, err := json.Marshal(domain)
	if err != nil {
		return err
	}
	_, err = a.pubSubClient.Publish(ctx, "domain-pipeline", jsonData)
	if err != nil {
		return err
	}
	return err
}
