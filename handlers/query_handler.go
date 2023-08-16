package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mahyar-m/pubsub-admin/db"
	"github.com/mahyar-m/pubsub-admin/models"
)

type QueryHandler struct {
}

func (qh *QueryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	config := db.MysqlConfig{}
	query := "SELECT * FROM `message` WHERE " + r.FormValue("query")

	messageRows, err := (&models.MessageModel{}).Select(config, query)
	if err != nil {
		log.Fatalf("Could get messages from DB: %v", err)
		return
	}

	jsonResponse, _ := json.Marshal(messageRows)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
