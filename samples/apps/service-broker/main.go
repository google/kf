// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

const (
	catalog = `{
		"services":[
			{
				"name": "fake-service",
				"id": "fake-service-id",
				"description": "a fake service",
				"bindable": true,
				"tags": ["fake-tag"],
				"plans": [
					{
						"id": "fake-plan-id",
						"name": "fake-plan",
						"description": "a fake service plan"
					}
				]
			}
		]
	}`
	creds = `{
		"credentials": {
			"username": "fake-user",
			"password": "fake-pw"
		}
	}`
)

func fakeJsonResponse(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}
}

func methodMapper(mapping map[string]http.HandlerFunc) http.HandlerFunc {
	badMethod := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Unknown method on endpoint: "+r.Method)
	}

	return func(w http.ResponseWriter, r *http.Request) {

		handler, ok := mapping[r.Method]
		if ok {
			handler(w, r)
		} else {
			badMethod(w, r)
		}
	}
}

func logRequest(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Method, r.URL)
		handler.ServeHTTP(w, r)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/v2/catalog", fakeJsonResponse(http.StatusOK, catalog))

	r.HandleFunc("/v2/service_instances/{instance_id}", methodMapper(map[string]http.HandlerFunc{
		http.MethodPut:    fakeJsonResponse(http.StatusCreated, `{}`),
		http.MethodDelete: fakeJsonResponse(http.StatusOK, `{}`),
	}))

	r.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", methodMapper(map[string]http.HandlerFunc{
		http.MethodPut:    fakeJsonResponse(http.StatusCreated, creds),
		http.MethodDelete: fakeJsonResponse(http.StatusOK, `{}`),
	}))

	log.Fatal(http.ListenAndServe(hostPort(), logRequest(r)))
}

func hostPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return fmt.Sprintf(":%s", port)
}
