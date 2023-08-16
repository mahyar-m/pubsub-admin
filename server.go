package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"cloud.google.com/go/pubsub"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/api/iterator"
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

	http.HandleFunc("/", mainPageHandler)
	http.HandleFunc("/sub", listSubHandler)
	http.HandleFunc("/pull", pullHandler)
	http.HandleFunc("/query", queryHandler)

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	topics, err := listTopics()
	if err != nil {
		log.Fatalf("Could not list topics: %v", err)
		return
	}

	page := Page{Topics: topics}
	tmpl := template.Must(template.ParseFiles("static/templates/main.html"))

	if err := tmpl.Execute(w, page); err != nil {
		log.Fatalf("Could not execute template: %v", err)
	}
}

type SubItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SubItems []SubItem

func listSubHandler(w http.ResponseWriter, r *http.Request) {
	subs, err := listSubscriptions(r.FormValue("topic_id"))
	if err != nil {
		log.Printf("Could list topics: %v", err)
		return
	}
	subItems := SubItems{}
	for _, sub := range subs {
		log.Printf("Sub: %v", sub.String())
		subItem := SubItem{
			ID:   sub.ID(),
			Name: sub.String(),
		}
		subItems = append(subItems, subItem)
	}
	jsonResponse, _ := json.Marshal(subItems)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func listTopics() ([]*pubsub.Topic, error) {
	// projectID := "my-project-id"
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
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

func listSubscriptions(topicID string) ([]*pubsub.Subscription, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
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

func queryHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")

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
	rows, err := db.Query(query)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer rows.Close() // Close the statement when we leave main() / the program terminates

	var messageRows []MessageRow

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var message MessageRow
		if err := rows.Scan(&message.Id, &message.Sub, &message.Data, &message.Attribute, &message.PublishTime, &message.DeliveryAttempt, &message.OrderingKey); err != nil {
			panic(err.Error())
		}
		messageRows = append(messageRows, message)
	}

	jsonResponse, _ := json.Marshal(messageRows)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

const (
	projectID string = "test"
)

// The Page struct holds page data for rendering
type Page struct {
	Topics []*pubsub.Topic
}

type MessageRow struct {
	Id              string
	Sub             string
	Data            sql.NullString
	Attribute       sql.NullString
	PublishTime     sql.NullTime
	DeliveryAttempt sql.NullInt32
	OrderingKey     sql.NullString
}
