package handlers

import (
	"encoding/json"
	pubsubHelper "example/pubsub_manager/pubsub"
	"log"
	"net/http"
)

type PullHandler struct {
}

func (h *PullHandler) Handle(w http.ResponseWriter, r *http.Request) {
	pubsubConfig := pubsubHelper.PubsubConfig{}

	subId := r.FormValue("sub_id")
	messageRecievedCount, err := pubsubHelper.Pull(pubsubConfig, subId)
	if err != nil {
		log.Fatal(err)
	}

	jsonResponse, _ := json.Marshal(messageRecievedCount)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
