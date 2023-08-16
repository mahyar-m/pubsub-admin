package main

import (
	"log"
	"net/http"
	"os"

	"example/pubsub_manager/handlers"
	"example/pubsub_manager/router"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
	port := "8080"

	router := &router.Router{}
	router.Route("GET", "/", (&handlers.SpaHandler{}).Handle)
	router.Route("POST", "/init", (&handlers.InitHandler{}).Handle)
	router.Route("GET", "/sub", (&handlers.SubscriptionHandler{}).Handle)
	router.Route("POST", "/pull", (&handlers.PullHandler{}).Handle)
	router.Route("GET", "/query", (&handlers.QueryHandler{}).Handle)

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
