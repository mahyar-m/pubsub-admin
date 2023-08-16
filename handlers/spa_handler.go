package handlers

import (
	pubsubHelper "example/pubsub_manager/pubsub"
	"html/template"
	"log"
	"net/http"

	"cloud.google.com/go/pubsub"
)

type Page struct {
	Topics []*pubsub.Topic
}

type SpaHandler struct {
	Page     Page
	Template *template.Template
}

func (spah *SpaHandler) Handle(w http.ResponseWriter, r *http.Request) {
	pubsubConfig := pubsubHelper.PubsubConfig{}
	topics, err := pubsubHelper.ListTopics(pubsubConfig)
	if err != nil {
		log.Fatalf("Could not list topics: %v", err)
		return
	}

	spah.Page = Page{Topics: topics}
	spah.Template = template.Must(template.ParseFiles("static/templates/main.html"))

	if err := spah.Template.Execute(w, spah.Page); err != nil {
		log.Fatalf("Could not execute template: %v", err)
	}
}
