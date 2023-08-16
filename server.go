package main

import (
	"log"
	"net/http"
	"os"

	"example/pubsub_manager/handlers"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
	port := "8080"

	http.HandleFunc("/", (&handlers.SpaHandler{}).Handle)
	http.HandleFunc("/init", (&handlers.InitHandler{}).Handle)
	http.HandleFunc("/sub", (&handlers.SubscriptionHandler{}).Handle)
	http.HandleFunc("/pull", (&handlers.PullHandler{}).Handle)
	http.HandleFunc("/query", (&handlers.QueryHandler{}).Handle)

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
