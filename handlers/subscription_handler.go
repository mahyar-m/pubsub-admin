package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	pubsubHelper "github.com/mahyar-m/pubsub-admin/pubsub"
)

type SubItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SubItems []SubItem

type SubscriptionHandler struct {
}

func (qh *SubscriptionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	pubsubConfig := pubsubHelper.PubsubConfig{}
	subs, err := pubsubHelper.ListSubscriptions(pubsubConfig, r.FormValue("topic_id"))
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
