package pubsub

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

func Publish(config PubsubConfig, topicID, msg string) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.GetProjectId())
	if err != nil {
		return fmt.Errorf("pubsub: NewClient: %v", err)
	}
	defer client.Close()

	topic := client.Topic(topicID)
	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})

	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: result.Get: %v", err)
	}
	log.Printf("Published a message; msg ID: %v\n", id)

	return nil
}
