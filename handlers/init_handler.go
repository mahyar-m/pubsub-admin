package handlers

import (
	"encoding/json"
	pubsubHelper "example/pubsub_manager/pubsub"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/pubsub"
)

type InitHandler struct {
}

func (h *InitHandler) Handle(w http.ResponseWriter, r *http.Request) {
	pubsubConfig := pubsubHelper.PubsubConfig{}

	err := pubsubHelper.CreateTopic(pubsubConfig, "test-topic")
	if err != nil {
		fmt.Println("Error:", err)
	}

	err = pubsubHelper.CreateSub(pubsubConfig, "test-sub", "test-topic", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	err = pubsubHelper.CreateSub(pubsubConfig, "test-sub1", "test-topic", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}

	retryPolicy := &pubsub.RetryPolicy{MinimumBackoff: 5 * time.Second, MaximumBackoff: 60 * time.Second}
	// pubsubHelper.DeleteSub(pubsubConfig, "test-sub-with-retry")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	err = pubsubHelper.CreateSub(pubsubConfig, "test-sub-with-retry", "test-topic", retryPolicy)
	if err != nil {
		fmt.Println("Error:", err)
	}

	err = pubsubHelper.Publish(pubsubConfig, "test-topic", "test message")
	if err != nil {
		fmt.Println("Publish Error:", err)
	}
	err = pubsubHelper.CreateTopic(pubsubConfig, "test-topic-1")
	if err != nil {
		fmt.Println("Error:", err)
	}

	jsonResponse, _ := json.Marshal("success")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
