package pubsub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

func ListTopics(config PubsubConfig) ([]*pubsub.Topic, error) {
	// projectID := "my-project-id"
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.GetProjectId())
	if err != nil {
		return nil, fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	var topics []*pubsub.Topic

	it := client.Topics(ctx)
	for {
		topic, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("next: %v", err)
		}
		topics = append(topics, topic)
	}

	return topics, nil
}
