package pubsub

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"example/pubsub_manager/db"
	"example/pubsub_manager/models"

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

func Pull(config PubsubConfig, subID string) (int, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.GetProjectId())
	if err != nil {
		return 0, fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(subID)

	// Receive messages for 5 seconds, which simplifies testing.
	// Comment this out in production, since `Receive` should
	// be used as a long running operation.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	dbConfig := db.MysqlConfig{}
	messageModel := models.MessageModel{}

	var received int32
	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))

		err := messageModel.Insert(dbConfig, subID, msg)
		if err != nil {
			panic(err.Error())
		}

		atomic.AddInt32(&received, 1)
		msg.Ack()
	})
	if err != nil {
		return 0, fmt.Errorf("sub.Receive: %v", err)
	}
	log.Printf("Received %d messages\n", received)

	return int(received), nil
}
