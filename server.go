package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mahyar-m/pubsub-admin/handlers"
	"github.com/mahyar-m/pubsub-admin/router"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
	port := "8080"

	router := &router.Router{}
	router.Route("GET", "/", (&handlers.SpaHandler{}).Handle, false)
	router.Route("POST", "/init", (&handlers.InitHandler{}).Handle, false)
	router.Route("GET", `/sub/(?P<sub_id>[A-Za-z0-9_-]+)/sub_policy`, (&handlers.SubscriptionPolicyHandler{}).Handle, true)
	router.Route("POST", `/sub/(?P<sub_id>[A-Za-z0-9_-]+)/sub_policy`, (&handlers.SubscriptionPolicyHandler{}).Handle, true)
	router.Route("GET", "/sub", (&handlers.SubscriptionHandler{}).Handle, false)
	router.Route("POST", "/pull", (&handlers.PullHandler{}).Handle, false)
	router.Route("GET", "/query", (&handlers.QueryHandler{}).Handle, false)

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
