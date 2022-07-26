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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ResponseTest func(t *testing.T, r *httptest.ResponseRecorder)

func All(tests ...ResponseTest) ResponseTest {
	return func(t *testing.T, r *httptest.ResponseRecorder) {
		t.Helper()

		for _, test := range tests {
			test(t, r)
		}
	}
}

func Status(statusCode int) ResponseTest {
	return func(t *testing.T, r *httptest.ResponseRecorder) {
		t.Helper()

		if r.Code != statusCode {
			t.Error("wanted status code:", statusCode, "got:", r.Code)
		}
	}
}

var StatusOK = Status(http.StatusOK)
var StatusNotFound = Status(http.StatusNotFound)

func JSONBody() ResponseTest {
	return func(t *testing.T, r *httptest.ResponseRecorder) {
		t.Helper()

		if !json.Valid(r.Body.Bytes()) {
			t.Error("body wasn't valid JSON")
		}
	}
}

func BodyContains(needle string) ResponseTest {
	return func(t *testing.T, r *httptest.ResponseRecorder) {
		t.Helper()

		haystack := r.Body.Bytes()
		if !bytes.Contains(haystack, []byte(needle)) {
			t.Errorf("body didn't contain %q", needle)
		}
	}
}

type HTTPTest struct {
	// Input
	Method string
	Target string
	Body   []byte

	Want ResponseTest
}

type HTTPTestSuite []HTTPTest

func (ht HTTPTestSuite) Test(t *testing.T) {
	t.Helper()
	// Don't introduce Parallel() at this level because cases should be run in
	// order.

	broker := NewBroker()

	for i, tc := range ht {
		name := fmt.Sprintf("%d %s %s", i, tc.Method, tc.Target)
		t.Run(name, func(t *testing.T) {
			var bodyReader io.Reader
			if tc.Body != nil {
				bodyReader = bytes.NewBuffer(tc.Body)
			}

			req := httptest.NewRequest(tc.Method, tc.Target, bodyReader)
			w := httptest.NewRecorder()
			broker.ServeHTTP(w, req)

			t.Logf("Response code: %d", w.Code)
			t.Logf("Body: %q", w.Body.String())

			tc.Want(t, w)
		})
	}
}

func Test_health(t *testing.T) {
	t.Parallel()

	cases := map[string]HTTPTestSuite{
		"get health": {
			{
				Method: http.MethodGet,
				Target: "/",
				Want:   All(StatusOK),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, tc.Test)
	}
}

func Test_catalog(t *testing.T) {
	t.Parallel()

	cases := map[string]HTTPTestSuite{
		"get catalog": {
			{
				Method: http.MethodGet,
				Target: "/v2/catalog",
				Want:   All(JSONBody(), StatusOK),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, tc.Test)
	}
}

func Test_serviceInstances(t *testing.T) {
	t.Parallel()

	cases := map[string]HTTPTestSuite{
		"creating normal plan": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id",
				Body:   []byte(`{}`),
				Want:   All(JSONBody(), StatusOK),
			},
		},
		"creating normal plan with accepts_incomplete": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id?accepts_incomplete=true",
				Body:   []byte(`{}`),
				Want:   All(JSONBody(), StatusOK),
			},
		},
		"creating async only plan without accepts_incomplete": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id",
				Body:   []byte(`{"plan_id":"fake-async-only-plan-guid"}`),
				Want:   All(Status(http.StatusUnprocessableEntity)),
			},
		},
		"creating async only plan with accepts_incomplete": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id?accepts_incomplete=true",
				Body:   []byte(`{"plan_id":"fake-async-only-plan-guid"}`),
				Want:   All(Status(http.StatusAccepted)),
			},
		},
		"updating async only plan starts operation": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id?accepts_incomplete=true",
				Body:   []byte(`{"plan_id":"fake-async-only-plan-guid"}`),
				Want:   All(),
			},
			{
				Method: http.MethodPatch,
				Target: "/v2/service_instances/test-id?accepts_incomplete=true",
				Body:   []byte(`{"plan_id":"fake-async-only-plan-guid"}`),
				Want:   All(Status(http.StatusAccepted)),
			},
		},
		"updating async only plan fails without accepts_incomplete": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id?accepts_incomplete=true",
				Body:   []byte(`{"plan_id":"fake-async-only-plan-guid"}`),
				Want:   All(),
			},
			{
				Method: http.MethodPatch,
				Target: "/v2/service_instances/test-id",
				Body:   []byte(`{"plan_id":"fake-async-only-plan-guid"}`),
				Want:   All(Status(http.StatusUnprocessableEntity)),
			},
		},
		"updating async to synchronous plan returns operation": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id?accepts_incomplete=true",
				Body:   []byte(`{"plan_id":"fake-async-only-plan-guid"}`),
				Want:   All(),
			},
			{
				Method: http.MethodPatch,
				Target: "/v2/service_instances/test-id?accepts_incomplete=true",
				Body:   []byte(`{"plan_id":"fake-plan-guid"}`),
				Want:   Status(http.StatusAccepted),
			},
		},
		"updating missing instance 404s": {
			{
				Method: http.MethodPatch,
				Target: "/v2/service_instances/missing?accepts_incomplete=true",
				Body:   []byte(`{"plan_id":"fake-plan-guid"}`),
				Want:   All(StatusNotFound),
			},
		},
		"deleting async plan starts operation": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id",
				Body:   []byte(`{"plan_id":"fake-async-plan-guid"}`),
				Want:   All(),
			},
			{
				Method: http.MethodDelete,
				Target: "/v2/service_instances/test-id",
				Want:   All(Status(http.StatusAccepted)),
			},
		},
		"deleting async only plan requires accepts_incomplete": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id?accepts_incomplete=true",
				Body:   []byte(`{"plan_id":"fake-async-only-plan-guid"}`),
				Want:   All(),
			},
			{
				Method: http.MethodDelete,
				Target: "/v2/service_instances/test-id",
				Want:   All(Status(http.StatusUnprocessableEntity)),
			},
			{
				Method: http.MethodDelete,
				Target: "/v2/service_instances/test-id?accepts_incomplete=true",
				Want:   All(Status(http.StatusAccepted)),
			},
		},
		"deleting missing instance 404s": {
			{
				Method: http.MethodDelete,
				Target: "/v2/service_instances/missing?accepts_incomplete=true",
				Want:   All(StatusNotFound),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, tc.Test)
	}
}

func Test_lastOperation(t *testing.T) {
	t.Parallel()

	cases := map[string]HTTPTestSuite{
		"404s for missing instance": {
			{
				Method: http.MethodGet,
				Target: "/v2/service_instances/test-id/last_operation",
				Want:   All(StatusNotFound),
			},
		},
		"last operation keeps track of calls": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id?accepts_incomplete=true",
				Body:   []byte(`{"plan_id":"fake-async-only-plan-guid"}`),
				Want:   All(Status(http.StatusAccepted)),
			},
			{
				Method: http.MethodGet,
				Target: "/v2/service_instances/test-id/last_operation",
				Want:   All(Status(http.StatusOK), BodyContains("in progress")),
			},
			{
				Method: http.MethodGet,
				Target: "/v2/service_instances/test-id/last_operation",
				Want:   All(Status(http.StatusOK), BodyContains("succeeded")),
			},
			{ // repeated calls should succeed
				Method: http.MethodGet,
				Target: "/v2/service_instances/test-id/last_operation",
				Want:   All(Status(http.StatusOK), BodyContains("succeeded")),
			},
		},
		"number of calls can be customized via config": {
			{
				Method: http.MethodPatch,
				Target: "/config",
				Body:   []byte(`{"max_fetch_service_instance_requests": 2}`),
				Want:   All(Status(http.StatusOK)),
			},
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id?accepts_incomplete=true",
				Body:   []byte(`{"plan_id":"fake-async-only-plan-guid"}`),
				Want:   All(Status(http.StatusAccepted)),
			},
			{
				Method: http.MethodGet,
				Target: "/v2/service_instances/test-id/last_operation",
				Want:   All(Status(http.StatusOK), BodyContains("in progress")),
			},
			{
				Method: http.MethodGet,
				Target: "/v2/service_instances/test-id/last_operation",
				Want:   All(Status(http.StatusOK), BodyContains("in progress")),
			},
			{
				Method: http.MethodGet,
				Target: "/v2/service_instances/test-id/last_operation",
				Want:   All(Status(http.StatusOK), BodyContains("succeeded")),
			},
			{ // repeated calls should succeed
				Method: http.MethodGet,
				Target: "/v2/service_instances/test-id/last_operation",
				Want:   All(Status(http.StatusOK), BodyContains("succeeded")),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, tc.Test)
	}
}

func Test_config(t *testing.T) {
	t.Parallel()

	cases := map[string]HTTPTestSuite{
		"serves current configuration": {
			{
				Method: http.MethodGet,
				Target: "/config",
				Want:   All(StatusOK, JSONBody()),
			},
		},
		"allows updating configuration": {
			{
				Method: http.MethodPost,
				Target: "/config",
				Body:   []byte("{}"),
				Want:   StatusOK,
			},
			{
				Method: http.MethodGet,
				Target: "/config",
				Want:   All(BodyContains(`{"behaviors":{"catalog":{}}}`), StatusOK),
			},
		},
		"allows resetting configuration": {
			{
				Method: http.MethodPost,
				Target: "/config",
				Body:   []byte("{}"),
				Want:   StatusOK,
			},
			{
				Method: http.MethodGet,
				Target: "/v2/catalog",
				Want:   Status(0), // Invalid due to invalid config.
			},
			{
				Method: http.MethodPost,
				Target: "/config/reset",
				Want:   StatusOK,
			},
			{
				Method: http.MethodGet,
				Target: "/v2/catalog",
				Want:   StatusOK,
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, tc.Test)
	}
}

func Test_bind(t *testing.T) {
	t.Parallel()

	cases := map[string]HTTPTestSuite{
		"404s for missing instance": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id/service_bindings/some-binding",
				Body:   []byte(`{}`),
				Want:   StatusNotFound,
			},
		},
		"valid binding": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id",
				Body:   []byte(`{}`),
				Want:   All(JSONBody(), StatusOK),
			},
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id/service_bindings/some-binding",
				Body:   []byte(`{}`),
				Want:   All(JSONBody(), Status(http.StatusCreated)),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, tc.Test)
	}
}

func Test_unbind(t *testing.T) {
	t.Parallel()

	cases := map[string]HTTPTestSuite{
		"404s for missing instance": {
			{
				Method: http.MethodDelete,
				Target: "/v2/service_instances/test-id/service_bindings/some-binding",
				Body:   []byte(`{}`),
				Want:   StatusNotFound,
			},
		},
		"valid binding": {
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id",
				Body:   []byte(`{}`),
				Want:   All(JSONBody(), StatusOK),
			},
			{
				Method: http.MethodPut,
				Target: "/v2/service_instances/test-id/service_bindings/some-binding",
				Body:   []byte(`{}`),
				Want:   All(JSONBody(), Status(http.StatusCreated)),
			},
			{
				Method: http.MethodDelete,
				Target: "/v2/service_instances/test-id/service_bindings/some-binding",
				Want:   All(JSONBody(), Status(http.StatusOK)),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, tc.Test)
	}
}
