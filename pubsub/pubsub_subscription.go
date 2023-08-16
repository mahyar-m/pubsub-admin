package pubsub

import (
	"context"
	"fmt"
	"os"

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

func CreateSub(config PubsubConfig, subID string, topicID string, retryPolicy *pubsub.RetryPolicy) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.GetProjectId())
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	topic := client.Topic(topicID)

	sub, err := client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
		Topic:       topic,
		RetryPolicy: nil,
	})
	if err != nil {
		return fmt.Errorf("CreateSubscription: %v", err)
	}
	fmt.Fprintf(os.Stdout, "Created subscription: %v\n", sub)
	return nil
}

func GetSub(config PubsubConfig, subID string) (*pubsub.Client, *pubsub.Subscription, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.GetProjectId())
	if err != nil {
		return nil, nil, fmt.Errorf("pubsub.NewClient: %v", err)
	}

	return client, client.Subscription(subID), nil
}

func DeleteSub(config PubsubConfig, subID string) error {
	client, sub, err := GetSub(config, subID)
	if err != nil {
		return err
	}
	defer client.Close()

	ctx := context.Background()
	if err := sub.Delete(ctx); err != nil {
		return err
	}

	return nil
}
