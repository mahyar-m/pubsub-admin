package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"example/pubsub_manager/handlers"

	"cloud.google.com/go/pubsub"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
	port := "8080"

	err := createTopic("test-topic")
	if err != nil {
		fmt.Println("Error:", err)
	}

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		fmt.Println("Error:", err)
	}
	topic := client.Topic("test-topic")
	err = createSub("test-sub", topic)
	if err != nil {
		fmt.Println("Error:", err)
	}
	err = createSub("test-sub1", topic)
	if err != nil {
		fmt.Println("Error:", err)
	}

	err = publish("test-topic", "test message")
	if err != nil {
		fmt.Println("Publish Error:", err)
	}
	err = createTopic("test-topic-1")
	if err != nil {
		fmt.Println("Error:", err)
	}

	http.HandleFunc("/", (&handlers.SpaHandler{}).Handle)
	http.HandleFunc("/sub", (&handlers.SubscriptionHandler{}).Handle)
	http.HandleFunc("/pull", pullHandler)
	http.HandleFunc("/query", (&handlers.QueryHandler{}).Handle)

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func createTopic(topicID string) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	t, err := client.CreateTopic(ctx, topicID)
	if err != nil {
		return fmt.Errorf("CreateTopic: %v", err)
	}
	fmt.Fprintf(os.Stdout, "Topic created: %v\n", t)
	return nil
}

func createSub(subID string, topic *pubsub.Topic) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	sub, err := client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
		Topic: topic,
	})
	if err != nil {
		return fmt.Errorf("CreateSubscription: %v", err)
	}
	fmt.Fprintf(os.Stdout, "Created subscription: %v\n", sub)
	return nil
}

func publish(topicID, msg string) error {
	// projectID := "my-project-id"
	// topicID := "my-topic"
	// msg := "Hello World"
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub: NewClient: %v", err)
	}
	defer client.Close()

	t := client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: result.Get: %v", err)
	}
	log.Printf("Published a message; msg ID: %v\n", id)

	return nil
}

func pullMsgs(subID string) (int, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(subID)

	// Receive messages for 10 seconds, which simplifies testing.
	// Comment this out in production, since `Receive` should
	// be used as a long running operation.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var received int32
	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))

		db, err := sql.Open("mysql", "root:root@tcp(localhost)/golang-docker?parseTime=true")
		if err != nil {
			panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		// Prepare statement for inserting data
		stmtIns, err := db.Prepare("INSERT INTO message (id, sub, data, attribute, publish_time, delivery_attempt, ordering_key) VALUES( ?, ?, ?, ?, ?, ?, ? )") // ? = placeholder
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

		var jsonResponse []byte = nil
		if msg.Attributes != nil {
			jsonResponse, _ = json.Marshal(msg.Attributes)
		}
		_, err = stmtIns.Exec(msg.ID, subID, msg.Data, jsonResponse, msg.PublishTime, msg.DeliveryAttempt, msg.OrderingKey) // Insert tuples (i, i^2)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
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

func pullHandler(w http.ResponseWriter, r *http.Request) {
	subId := r.FormValue("sub_id")
	messageRecievedCount, err := pullMsgs(subId)
	if err != nil {
		log.Fatal(err)
	}

	jsonResponse, _ := json.Marshal(messageRecievedCount)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

const (
	projectID string = "test"
)
