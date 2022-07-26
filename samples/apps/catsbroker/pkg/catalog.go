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
	"net/http"
)

// NewDefaultCatalog creates a valid default configuration for the broker
// suitable for basic tests.
func NewDefaultCatalog() *Datastore {
	datastore := &Datastore{}

	datastore.Behaviors.Catalog = Behavior{
		StatusCode: http.StatusOK,
		Body: jsonPointer(`{
       "services": [
         {
           "name": "fake-service",
           "id": "f479b64b-7c25-42e6-8d8f-e6d22c456c9b",
           "description": "fake service",
           "tags": [
             "no-sql",
             "relational"
           ],
           "requires": [
             "route_forwarding"
           ],
           "max_db_per_node": 5,
           "bindable": true,
           "metadata": {
             "provider": {
               "name": "The name"
             },
             "listing": {
               "imageUrl": "http://catgifpage.com/cat.gif",
               "blurb": "fake broker that is fake",
               "longDescription": "A long time ago, in a galaxy far far away..."
             },
             "displayName": "The Fake Broker"
           },
           "dashboard_client": {
             "id": "sso-test",
             "secret": "sso-secret",
             "redirect_uri": "http://localhost:5551"
           },
           "plan_updateable": true,
           "plans": [
             {
               "name": "fake-plan",
               "id": "fake-plan-guid",
               "description": "Shared fake Server, 5tb persistent disk, 40 max concurrent connections",
               "max_storage_tb": 5,
               "metadata": {"cost": 0}
             },
             {
               "name": "fake-async-plan",
               "id": "fake-async-plan-guid",
               "description": "Shared fake Server, 5tb persistent disk, 40 max concurrent connections. 100 async",
               "max_storage_tb": 5,
               "metadata": {"cost": 0}
             },
             {
               "name": "fake-async-only-plan",
               "id": "fake-async-only-plan-guid",
               "description": "Shared fake Server, 5tb persistent disk, 40 max concurrent connections. 100 async",
               "max_storage_tb": 5,
               "metadata": {"cost": 0}
             }
           ]
         }
       ]
     }`),
	}

	instanceOperations := func() map[string]Behavior {
		return map[string]Behavior{
			"fake-async-plan-guid": {
				StatusCode: http.StatusAccepted,
				Body:       jsonPointer(`{}`),
			},
			"fake-async-only-plan-guid": {
				AsyncOnly:  true,
				StatusCode: http.StatusAccepted,
				Body:       jsonPointer(`{}`),
			},
			"default": {
				StatusCode: http.StatusOK,
				Body:       jsonPointer(`{}`),
			},
		}
	}

	datastore.Behaviors.Provision = instanceOperations()
	datastore.Behaviors.Update = instanceOperations()
	datastore.Behaviors.Deprovision = instanceOperations()

	datastore.Behaviors.Fetch = map[string]FetchBehavior{
		"default": {
			InProgress: Behavior{
				StatusCode: http.StatusOK,
				Body:       jsonPointer(`{"state": "in progress"}`),
			},
			Finished: Behavior{
				StatusCode: http.StatusOK,
				Body:       jsonPointer(`{"state": "succeeded"}`),
			},
		},
	}

	datastore.Behaviors.Bind = map[string]Behavior{
		"default": {
			StatusCode: http.StatusCreated,
			Body: jsonPointer(`{
				 "route_service_url": "https://logging-route-service.bosh-lite.com",
				 "credentials": {
					 "uri": "fake-service://fake-user:fake-password@fake-host:3306/fake-dbname",
					 "username": "fake-user",
					 "password": "fake-password",
					 "host": "fake-host",
					 "port": 3306,
					 "database": "fake-dbname"
				 }
			 }`),
		},
	}

	datastore.Behaviors.Unbind = map[string]Behavior{
		"default": {
			StatusCode: http.StatusOK,
			Body:       jsonPointer(`{}`),
		},
	}

	datastore.MaxFetchServiceInstanceRequests = 1

	return datastore
}

func jsonPointer(contents string) *json.RawMessage {
	tmp := json.RawMessage([]byte(contents))
	return &tmp
}
