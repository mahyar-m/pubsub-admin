package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	pubsubHelper "github.com/mahyar-m/pubsub-admin/pubsub"
)

type PullHandler struct {
}

func (h *PullHandler) Handle(w http.ResponseWriter, r *http.Request) {
	pubsubConfig := pubsubHelper.PubsubConfig{}

	subId := r.FormValue("sub_id")
	timeout, err := strconv.ParseInt(r.FormValue("timeout"), 10, 64)
	if err != nil {
		log.Fatal(err)
		return
	}

	limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 64)
	if err != nil {
		log.Fatal(err)
		return
	}

	isAck, err := strconv.ParseBool(r.FormValue("is_ack"))
	if err != nil {
		log.Fatal(err)
		return
	}

	messageRecievedCount, err := pubsubHelper.Pull(pubsubConfig, subId, int(timeout), int(limit), isAck)
	if err != nil {
		log.Fatal(err)
	}

	jsonResponse, _ := json.Marshal(messageRecievedCount)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
