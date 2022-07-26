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
	"net/http"
)

// RequestType identifies the type of request internally.
type RequestType string

const (
	Catalog         RequestType = "catalog"
	Provision       RequestType = "provision"
	Deprovision     RequestType = "deprovision"
	Bind            RequestType = "bind"
	Unbind          RequestType = "unbind"
	Update          RequestType = "update"
	FetchInProgress RequestType = "fetch_in_progress"
	FetchFinished   RequestType = "fetch_finished"
)

// getStringValue extracts a string value from a given key in a map.
func getStringValue(key string, m map[string]interface{}) (string, bool) {
	if m == nil {
		return "", false
	}

	rawKey, ok := m[key]
	if !ok {
		return "", false
	}

	value, ok := rawKey.(string)
	if !ok {
		return "", false
	}

	return value, true
}

// ServiceBinding holds data about a mock service binding.
type ServiceBinding struct {
	BindingData map[string]interface{} `json:"binding_data"`
}

// ServiceInstance holds data about a mock service instance.
type ServiceInstance struct {
	ProvisionData map[string]interface{}    `json:"provision_data"`
	FetchCount    int                       `json:"fetch_count"`
	Deleted       bool                      `json:"deleted"`
	Bindings      map[string]ServiceBinding `json:"bindings"`
}

// PlanID returns the plan for the service instance. If the plan is undefined,
// "default" is returned.
func (si *ServiceInstance) PlanID() string {
	planID, ok := getStringValue("plan_id", si.ProvisionData)
	if !ok {
		return "default"
	}
	return planID
}

// Delete marks the instance as deleted.
func (si *ServiceInstance) Delete() {
	si.FetchCount = 0
	si.Deleted = true
}

// IncrementFetchCount increments the number of times this item has been
// polled and returns the new count.
func (si *ServiceInstance) IncrementFetchCount() int {
	si.FetchCount++
	return si.FetchCount
}

// Behavior controls the response to the service broker HTTP request.
type Behavior struct {
	// NumSleepSeconds is the number of seconds to sleep before returning a response.
	SleepSeconds int `json:"sleep_seconds,omitempty"`

	// AsyncOnly determines whether the response being successful is predicated on
	// ?accepts_incomplete=true being set in the OSB request.
	AsyncOnly bool `json:"async_only,omitempty"`

	// StatusCode is the HTTP status code to return.
	StatusCode int `json:"status,omitempty"`

	// Body is a valid JSON object to respond with.
	Body *json.RawMessage `json:"body,omitempty"`

	// RawBody is sent as a response if body is unset and allows sending
	// invalid JSON as responses.
	RawBody []byte `json:"raw_body,omitempty"`
}

// FetchBehavior contains a set of behaviors to return to clients requesting
// asynchronous service instance information.
type FetchBehavior struct {
	InProgress Behavior `json:"in_progress,omitempty"`
	Finished   Behavior `json:"finished,omitempty"`
}

// Datastore contains configuration and state for mocking a service broker.
type Datastore struct {
	// Explicitly allow comments to be set.
	_ []string `json:"comments,omitempty"`

	// Behaviors contains a set of responses to requests. All behaviors except for
	// Catalog are map[string]Behavior objects indexed by the service plan GUID.
	// If no entry exists for a GUID, then the behavior with the "default" key is
	// used.
	Behaviors struct {
		// Catalog defines the response for the /v2/catalog endpoint.
		Catalog Behavior `json:"catalog,omitempty"`

		// Provision contains a set of behaviors indexed by plan GUID to replay
		// to clients issuing PUT /v2/service_instances/:id/ requests.
		Provision map[string]Behavior `json:"provision,omitempty"`

		// Deprovision contains a set of behaviors indexed by plan GUID to replay
		// to clients issuing DELETE /v2/service_instances/:id/ requests.
		Deprovision map[string]Behavior `json:"deprovision,omitempty"`

		// Bind contains a set of behaviors indexed by plan GUID to replay
		// to clients issuing
		// PUT /v2/service_instances/:instance_id/service_bindings/:id
		// requests.
		Bind map[string]Behavior `json:"bind,omitempty"`

		// Unbind contains a set of behaviors indexed by plan GUID to replay
		// to clients issuing
		// DELETE /v2/service_instances/:instance_id/service_bindings/:id
		// requests.
		Unbind map[string]Behavior `json:"unbind,omitempty"`

		// Update contains a set of behaviors indexed by plan GUID to replay
		// to clients issuing
		// UPDATE /v2/service_instances/:instance_id
		// requests.
		Update map[string]Behavior `json:"update,omitempty"`

		// Fetch contains a set of behaviors indexed by plan GUID then provision
		// state to replay to clients issuing
		// GET /v2/service_instances/:instance_id/service_bindings/:id
		// requests.
		Fetch map[string]FetchBehavior `json:"fetch,omitempty"`
	} `json:"behaviors,omitempty"`

	// MaxFetchServiceInstanceRequests is the number of async "in progress"
	// responses to send before responding with a ready fetch service response.
	MaxFetchServiceInstanceRequests int `json:"max_fetch_service_instance_requests,omitempty"`

	// ServiceInstances is a map of service instance ID to ServiceInstance for
	// each instance the broker has provisioned.
	ServiceInstances map[string]*ServiceInstance `json:"service_instances,omitempty"`
}

// CreateServiceInstance adds a ServiceInstance to the datastore.
func (ds *Datastore) CreateServiceInstance(id string, provisionData map[string]interface{}) *ServiceInstance {
	if ds.ServiceInstances == nil {
		ds.ServiceInstances = make(map[string]*ServiceInstance)
	}

	tmp := &ServiceInstance{
		ProvisionData: provisionData,
	}

	ds.ServiceInstances[id] = tmp

	return tmp
}

// CreateServiceInstance adds a ServiceBinding to a ServiceInstance in the datastore.
func (ds *Datastore) CreateServiceBinding(instanceID, bindingID string, bindingData map[string]interface{}) ServiceBinding {
	if ds.ServiceInstances == nil {
		ds.ServiceInstances = make(map[string]*ServiceInstance)
	}

	instance, ok := ds.ServiceInstances[instanceID]
	if !ok {
		instance = ds.CreateServiceInstance(instanceID, nil)
	}

	if instance.Bindings == nil {
		instance.Bindings = make(map[string]ServiceBinding)
	}

	tmp := ServiceBinding{
		BindingData: bindingData,
	}

	instance.Bindings[bindingID] = tmp

	return tmp
}

// Behavior finds the correct behavior to mock given the request type and
// plan.
func (ds *Datastore) Behavior(t RequestType, planID string) Behavior {
	invalidBehavior := Behavior{
		StatusCode: http.StatusTeapot, // Use 412 - "I'm a little teapot." to indicate a mock error.
		RawBody:    []byte(fmt.Sprintf("no behavior or default defined for request type %q for plan_id %q", t, planID)),
	}

	var behaviorMap map[string]Behavior
	switch t {
	case Catalog:
		return ds.Behaviors.Catalog

	case Bind:
		behaviorMap = ds.Behaviors.Bind

	case Unbind:
		behaviorMap = ds.Behaviors.Unbind

	case Deprovision:
		behaviorMap = ds.Behaviors.Deprovision

	case Provision:
		behaviorMap = ds.Behaviors.Provision

	case Update:
		behaviorMap = ds.Behaviors.Update

	case FetchInProgress:
		behaviorMap = make(map[string]Behavior)
		for k, v := range ds.Behaviors.Fetch {
			behaviorMap[k] = v.InProgress
		}

	case FetchFinished:
		behaviorMap = make(map[string]Behavior)
		for k, v := range ds.Behaviors.Fetch {
			behaviorMap[k] = v.Finished
		}

	default:
		return invalidBehavior
	}

	if behavior, ok := behaviorMap[planID]; ok {
		return behavior
	}

	if behavior, ok := behaviorMap["default"]; ok {
		return behavior
	}

	return invalidBehavior
}
