package handlers

import (
	"encoding/json"
	pubsubHelper "example/pubsub_manager/pubsub"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
)

type retryPolicyResponse struct {
	MinimumBackoff int
	MaximumBackoff int
}

type SubscriptionPolicyHandler struct {
}

func (sph *SubscriptionPolicyHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "post" {
		sph.postHandle(w, r)
	} else {
		sph.getHandle(w, r)
	}
}

func (sph *SubscriptionPolicyHandler) postHandle(w http.ResponseWriter, r *http.Request) {
	pubsubConfig := pubsubHelper.PubsubConfig{}
	subId := r.FormValue("sub_id")

	minimumBackoff, err := strconv.ParseInt(r.FormValue("min_back_off"), 10, 64)
	if err != nil {
		log.Fatal(err)
		return
	}

	maximumBackoff, err := strconv.ParseInt(r.FormValue("max_back_off"), 10, 64)
	if err != nil {
		log.Fatal(err)
		return
	}
	retryPolicy := &pubsub.RetryPolicy{MinimumBackoff: time.Duration(minimumBackoff) * time.Second, MaximumBackoff: time.Duration(maximumBackoff) * time.Second}
	err = pubsubHelper.UpdateSubPolicy(pubsubConfig, subId, retryPolicy)
	if err != nil {
		log.Printf("Could not update policy: %v", err)
		return
	}

	jsonResponse, _ := json.Marshal("")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (sph *SubscriptionPolicyHandler) getHandle(w http.ResponseWriter, r *http.Request) {
	pubsubConfig := pubsubHelper.PubsubConfig{}
	// TODO: conver any to stirng!
	// subId := r.Context().Value("sub_id")
	subId := "test-sub-with-retry"

	retryPolicy, err := pubsubHelper.GetSubPolicy(pubsubConfig, subId)
	if err != nil {
		log.Printf("Could not get policy: %v", err)
		return
	}

	if retryPolicy == nil {
		return
	}

	retryPolicyResponse := retryPolicyResponse{
		// TODO: figure out how to convert to int
		// max: (retryPolicy.MaximumBackoff / time.Second),
		// min: retryPolicy.MinimumBackoff / time.Second,
		MinimumBackoff: 5,
		MaximumBackoff: 60,
	}

	jsonResponse, _ := json.Marshal(retryPolicyResponse)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
