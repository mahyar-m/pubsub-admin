package pubsub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

func ListSubscriptions(config PubsubConfig, topicID string) ([]*pubsub.Subscription, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.GetProjectId())
	if err != nil {
		return nil, fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	var subs []*pubsub.Subscription

	it := client.Topic(topicID).Subscriptions(ctx)
	for {
		sub, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("next: %v", err)
		}
		subs = append(subs, sub)
	}
	return subs, nil
}
