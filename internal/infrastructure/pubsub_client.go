package infrastructure

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

type PubSubClient interface {
	Publish(ctx context.Context, topicID string, data []byte) (string, error)
	Subscribe(ctx context.Context, subscriptionID string, handler func(context.Context, *pubsub.Message)) error
	Close() error
}

type pubSubClient struct {
	client *pubsub.Client
	ctx    context.Context
}

// NewPubSubClient Google Cloud Pub/Subクライアントを初期化
func NewPubSubClient(projectID string, credPath string) PubSubClient {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID, option.WithCredentialsFile(credPath))
	if err != nil {
		panic(err)
	}
	return &pubSubClient{
		client: client,
		ctx:    ctx,
	}
}

// Publish トピックにメッセージをパブリッシュ
func (c *pubSubClient) Publish(ctx context.Context, topicID string, data []byte) (string, error) {
	topic := c.client.Topic(topicID)
	defer topic.Stop()

	result := topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})

	// メッセージIDを取得（ブロッキング）
	messageID, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to publish message: %w", err)
	}

	slog.Info("Published message",
		"topic_id", topicID,
		"message_id", messageID,
		"data_size", len(data),
	)

	return messageID, nil
}

// Subscribe サブスクリプションからメッセージを受信
// handlerは各メッセージに対して呼ばれるコールバック関数
// このメソッドはブロッキングし、contextがキャンセルされるまで実行し続けます
func (c *pubSubClient) Subscribe(ctx context.Context, subscriptionID string, handler func(context.Context, *pubsub.Message)) error {
	sub := c.client.Subscription(subscriptionID)

	slog.Info("Starting to receive messages",
		"subscription_id", subscriptionID,
	)

	// Receiveはcontextがキャンセルされるまでブロックする
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		slog.Info("Received message",
			"subscription_id", subscriptionID,
			"message_id", msg.ID,
			"publish_time", msg.PublishTime,
		)

		// ユーザー定義のハンドラーを実行
		handler(ctx, msg)
	})

	if err != nil {
		return fmt.Errorf("failed to receive messages: %w", err)
	}

	return nil
}

// Close クライアントをクローズ
func (c *pubSubClient) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("failed to close pubsub client: %w", err)
	}

	slog.Info("Pub/Sub client closed")
	return nil
}
