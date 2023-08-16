package pubsub

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mahyar-m/pubsub-admin/db"
	"github.com/mahyar-m/pubsub-admin/models"

	"cloud.google.com/go/pubsub"
)

type MessageCount struct {
	TotalRecived int32
	Inserted     int32
	Duplicate    int32
}

var (
	mutex sync.Mutex
)

func Publish(config PubsubConfig, topicID string, msg_data []byte) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.GetProjectId())
	if err != nil {
		return fmt.Errorf("pubsub: NewClient: %v", err)
	}
	defer client.Close()

	topic := client.Topic(topicID)
	result := topic.Publish(ctx, &pubsub.Message{
		Data: msg_data,
	})

	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: result.Get: %v", err)
	}
	log.Printf("Published a message; msg ID: %v\n", id)

	return nil
}

func Pull(config PubsubConfig, subID string, timeout int, limit int, isAck bool) (MessageCount, error) {
	var messageCount MessageCount

	client, sub, err := GetSub(config, subID)
	if err != nil {
		return messageCount, err
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

	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		mutex.Lock()
		atomic.AddInt32(&messageCount.TotalRecived, 1)
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

			atomic.AddInt32(&messageCount.Inserted, 1)
		} else {
			log.Printf("Duplicate message: %q\n", string(msg.ID))
			atomic.AddInt32(&messageCount.Duplicate, 1)
		}

		if isAck {
			msg.Ack()
		} else {
			msg.Nack()
		}

		if limit != 0 && messageCount.Inserted == int32(limit) {
			cancel()
		}

		mutex.Unlock()
	})
	if err != nil {
		return messageCount, fmt.Errorf("sub.Receive: %v", err)
	}
	log.Printf("Received total %d messages. Inserted: %d and Duplicated: %d\n", messageCount.TotalRecived, messageCount.Inserted, messageCount.Duplicate)

	return messageCount, nil
}
