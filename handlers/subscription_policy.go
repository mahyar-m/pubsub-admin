package handlers

import (
	"encoding/json"
	pubsubHelper "example/pubsub_manager/pubsub"
	"example/pubsub_manager/router"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
)

type retryPolicyResponse struct {
	MinimumBackoff int64
	MaximumBackoff int64
}

type SubscriptionPolicyHandler struct {
}

func (sph *SubscriptionPolicyHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		sph.postHandle(w, r)
	} else {
		sph.getHandle(w, r)
	}
}

func (sph *SubscriptionPolicyHandler) postHandle(w http.ResponseWriter, r *http.Request) {
	pubsubConfig := pubsubHelper.PubsubConfig{}
	params := router.RequestParams(r)
	subId := params["sub_id"]

	minimumBackoff, err := strconv.ParseInt(r.FormValue("MinimumBackoff"), 10, 64)
	if err != nil {
		log.Fatal(err)
		return
	}

	maximumBackoff, err := strconv.ParseInt(r.FormValue("MaximumBackoff"), 10, 64)
	if err != nil {
		log.Fatal(err)
		return
	}

	var retryPolicy pubsub.RetryPolicy
	if minimumBackoff != 0 || maximumBackoff != 0 {
		retryPolicy = pubsub.RetryPolicy{MinimumBackoff: time.Duration(minimumBackoff) * time.Second, MaximumBackoff: time.Duration(maximumBackoff) * time.Second}
	}

	err = pubsubHelper.UpdateSubPolicy(pubsubConfig, subId, &retryPolicy)
	if err != nil {
		log.Printf("Could not update policy: %v", err)
		return
	}

	jsonResponse, _ := json.Marshal("success")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (sph *SubscriptionPolicyHandler) getHandle(w http.ResponseWriter, r *http.Request) {
	pubsubConfig := pubsubHelper.PubsubConfig{}
	params := router.RequestParams(r)
	subId := params["sub_id"]

	retryPolicy, err := pubsubHelper.GetSubPolicy(pubsubConfig, subId)
	if err != nil {
		log.Printf("Could not get policy: %v", err)
		return
	}

	if retryPolicy == nil {
		return
	}

	retryPolicyResponse := retryPolicyResponse{
		MinimumBackoff: int64(retryPolicy.MinimumBackoff.(time.Duration) / time.Second),
		MaximumBackoff: int64(retryPolicy.MaximumBackoff.(time.Duration) / time.Second),
	}

	jsonResponse, _ := json.Marshal(retryPolicyResponse)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
