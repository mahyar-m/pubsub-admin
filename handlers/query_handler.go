package handlers

import (
	"encoding/json"
	"example/pubsub_manager/db"
	"example/pubsub_manager/models"
	"log"
	"net/http"
)

type QueryHandler struct {
}

func (qh *QueryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	config := db.MysqlConfig{}
	query := r.FormValue("query")

	messageRows, err := (&models.MessageModel{}).Select(config, query)
	if err != nil {
		log.Fatalf("Could get messages from DB: %v", err)
		return
	}

	jsonResponse, _ := json.Marshal(messageRows)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
