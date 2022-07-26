// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package catsbroker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// respondNotFound sends a 404 response, nothing
// should be added to the ResponseWriter after this function is called.
func respondNotFound(w http.ResponseWriter, resource string) {
	w.WriteHeader(http.StatusNotFound)

	w.Write([]byte(fmt.Sprintf("%q not found", resource)))
}

// respondErr sends a 500 response for a runtime error, nothing
// should be added to the ResponseWriter after this function is called.
func respondErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)

	w.Write([]byte(fmt.Sprintf("error: %v", err)))
}

// respondWithBehavior actuates a behavior and writes the response, nothing
// should be added to the ResponseWriter after this function is called.
func respondWithBehavior(w http.ResponseWriter, behavior Behavior, acceptsIncomplete bool) {
	time.Sleep(time.Duration(behavior.SleepSeconds) * time.Second)

	if behavior.AsyncOnly && !acceptsIncomplete {
		// respond async required
		msg := json.RawMessage([]byte(`{"error":"AsyncRequired", "description":"This plan requires asynchronous operations."}`))
		behavior = Behavior{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       &msg,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(behavior.StatusCode)

	if behavior.Body != nil {
		w.Write(*behavior.Body)
	} else {
		w.Write(behavior.RawBody)
	}
}

// acceptsIncomplete checks for the accepts_incomplete=true query parameter used
// by OSB clients to indicate they support polling for asynchronous operations.
func acceptsIncomplete(r *http.Request) bool {
	if r == nil {
		return false
	}

	return r.URL.Query().Get("accepts_incomplete") == "true"
}

// unmarshalBodyOrFail deserializes the body of the request or fails. If an error
// is returned, nothing should be added to the ResponseWriter because the error
// has already been sent to the client.
func unmarshalBodyOrFail(w http.ResponseWriter, r *http.Request, dest interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, err)
		return err
	}
	if err := json.Unmarshal(body, &dest); err != nil {
		respondErr(w, err)
		return err
	}

	return nil
}

// NewBroker creates a mock broker.
func NewBroker() http.Handler {
	data := NewDefaultCatalog()

	r := mux.NewRouter()

	// Health check.
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	// Get catalog.
	r.Methods("GET").Path("/v2/catalog").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondWithBehavior(w, data.Behaviors.Catalog, acceptsIncomplete(r))
	})

	// Create service.
	r.Methods("PUT").Path("/v2/service_instances/{instance_id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		instanceID := mux.Vars(r)["instance_id"]

		var requestData map[string]interface{}
		if err := unmarshalBodyOrFail(w, r, &requestData); err != nil {
			return // function handles writing a response
		}

		instance := data.CreateServiceInstance(instanceID, requestData)
		behavior := data.Behavior(Provision, instance.PlanID())
		respondWithBehavior(w, behavior, acceptsIncomplete(r))
	})

	// Delete service.
	r.Methods("DELETE").Path("/v2/service_instances/{instance_id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		instanceID := mux.Vars(r)["instance_id"]

		instance, ok := data.ServiceInstances[instanceID]
		if ok {
			instance.Delete()
			respondWithBehavior(w, data.Behavior(Deprovision, instance.PlanID()), acceptsIncomplete(r))
			return
		}

		respondNotFound(w, "service instance")
	})

	// Poll last operation.
	r.Methods("GET").Path("/v2/service_instances/{instance_id}/last_operation").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		instanceID := mux.Vars(r)["instance_id"]

		instance, ok := data.ServiceInstances[instanceID]
		if ok {
			fetchCount := instance.IncrementFetchCount()
			if maxCt := data.MaxFetchServiceInstanceRequests; maxCt == 0 || fetchCount > maxCt {
				respondWithBehavior(w, data.Behavior(FetchFinished, instance.PlanID()), acceptsIncomplete(r))
			} else {
				respondWithBehavior(w, data.Behavior(FetchInProgress, instance.PlanID()), acceptsIncomplete(r))
			}
			return
		}

		respondNotFound(w, "service instance")
	})

	// Update service.
	r.Methods("PATCH").Path("/v2/service_instances/{instance_id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		instanceID := mux.Vars(r)["instance_id"]

		instance, ok := data.ServiceInstances[instanceID]
		if !ok {
			respondNotFound(w, "service instance")
			return
		}

		behavior := data.Behavior(Update, instance.PlanID())
		if behavior.StatusCode == http.StatusOK || behavior.StatusCode == http.StatusAccepted {
			// Marshal directly into ProvisionData to merge the properties.
			if err := unmarshalBodyOrFail(w, r, &instance.ProvisionData); err != nil {
				return // function handles writing a response
			}
			instance.FetchCount = 0 // Updating starts a new operation.
		}

		respondWithBehavior(w, behavior, acceptsIncomplete(r))
	})

	// Create binding.
	r.Methods("PUT").Path("/v2/service_instances/{instance_id}/service_bindings/{binding_id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		instanceID := mux.Vars(r)["instance_id"]
		bindingID := mux.Vars(r)["binding_id"]

		var requestData map[string]interface{}
		if err := unmarshalBodyOrFail(w, r, &requestData); err != nil {
			return // function handles writing a response
		}

		instance, ok := data.ServiceInstances[instanceID]
		if !ok {
			respondNotFound(w, "service instance")
			return
		}

		data.CreateServiceBinding(instanceID, bindingID, requestData)
		behavior := data.Behavior(Bind, instance.PlanID())
		respondWithBehavior(w, behavior, acceptsIncomplete(r))
	})

	// Delete binding.
	r.Methods("DELETE").Path("/v2/service_instances/{instance_id}/service_bindings/{binding_id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		instanceID := mux.Vars(r)["instance_id"]
		bindingID := mux.Vars(r)["binding_id"]

		instance, ok := data.ServiceInstances[instanceID]
		if !ok {
			respondNotFound(w, "service instance")
			return
		}

		if _, ok := instance.Bindings[bindingID]; !ok {
			respondNotFound(w, "service binding")
			return
		}
		delete(instance.Bindings, bindingID)

		respondWithBehavior(w, data.Behavior(Unbind, instance.PlanID()), acceptsIncomplete(r))
	})

	r.Methods("GET").Path("/config").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			respondErr(w, err)
			return
		}

		respondWithBehavior(w, Behavior{
			StatusCode: http.StatusOK,
			RawBody:    dataBytes,
		}, false)
	})

	r.Methods("POST").Path("/config").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tmp Datastore
		if err := unmarshalBodyOrFail(w, r, &tmp); err != nil {
			return // function handles writing a response
		}

		data = &tmp

		w.WriteHeader(http.StatusOK)
	})

	// Reset the configuration to the default state.
	r.Methods("POST").Path("/config/reset").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data = NewDefaultCatalog()
		w.WriteHeader(http.StatusOK)
	})

	// Perform a partial update of the configuration.
	r.Methods("PATCH").Path("/config").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := unmarshalBodyOrFail(w, r, &data); err != nil {
			return // function handles writing a response
		}
		w.WriteHeader(http.StatusOK)
	})

	return r
}
