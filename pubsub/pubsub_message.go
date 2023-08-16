package pubsub

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"example/pubsub_manager/db"
	"example/pubsub_manager/models"

	"cloud.google.com/go/pubsub"
)

var (
	mutex sync.Mutex
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

func Pull(config PubsubConfig, subID string, timeout int, limit int, isAck bool) (int, error) {
	client, sub, err := GetSub(config, subID)
	if err != nil {
		return 0, err
	}
	defer client.Close()

	if limit != 0 {
		sub.ReceiveSettings.Synchronous = true
		sub.ReceiveSettings.MaxOutstandingMessages = limit
	}

	var cancel context.CancelFunc
	ctx, cancel := context.WithCancel(context.Background())

	if timeout != 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	}
	defer cancel()

	dbConfig := db.MysqlConfig{}
	messageModel := models.MessageModel{}

	var received int32
	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		mutex.Lock()
		log.Printf("Got message with ID: %q, and Data:%q \n", msg.ID, string(msg.Data))
		isDuplicate, err := messageModel.IsDuplicate(dbConfig, subID, msg)
		if err != nil {
			panic(err.Error())
		}

		if !isDuplicate {
			err := messageModel.Insert(dbConfig, subID, msg)
			if err != nil {
				panic(err.Error())
			}

			atomic.AddInt32(&received, 1)
		} else {
			log.Printf("Duplicate message: %q\n", string(msg.ID))
		}

		if isAck {
			msg.Ack()
		} else {
			msg.Nack()
		}

		if limit != 0 && received == int32(limit) {
			cancel()
		}

		mutex.Unlock()
	})
	if err != nil {
		return 0, fmt.Errorf("sub.Receive: %v", err)
	}
	log.Printf("Received %d messages\n", received)

	return int(received), nil
}
