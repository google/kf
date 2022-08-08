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

package pkg

import (
	"bytes"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	respondOk = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	respondNotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	lookupOneIP = func(addr string) ([]net.IP, error) {
		return []net.IP{
			net.IPv4(8, 8, 8, 8),
		}, nil
	}

	lookupZeroIP = func(addr string) ([]net.IP, error) {
		return []net.IP{}, nil
	}

	lookupIPErr = func(addr string) ([]net.IP, error) {
		return nil, errors.New("couldn't look up address")
	}
)

func defaultRequest() *http.Request {
	return httptest.NewRequest("GET", "http://proxy", nil)
}

func defaultValues(t *testing.T, lb *LoadBalancer) {
	if lb.SessionCookie == "" {
		lb.SessionCookie = "JSESSIONID"
	}

	if lb.StickyCookie == "" {
		lb.StickyCookie = "__VCAP_ID__"
	}

	if lb.ProxyService == "" {
		lb.ProxyService = "service.namespace.svc.cluster.local"
	}

	if lb.ProxyPort == "" {
		lb.ProxyPort = "8080"
	}

	if lb.ProxyScheme == "" {
		lb.ProxyScheme = "http"
	}

	if lb.ReverseProxy == nil {
		lb.ReverseProxy = respondOk
	}

	if lb.LookupIP == nil {
		lb.LookupIP = lookupOneIP
	}

	if lb.ErrorLog == nil {
		lb.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
	}

	if lb.RandIntn == nil {
		lb.RandIntn = func(int) int {
			return 0
		}
	}
}

func AssertCookie(t *testing.T, name string, value string, recorder *httptest.ResponseRecorder) {
	t.Helper()

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == name {
			if cv := cookie.Value; cv != value {
				t.Errorf("Cookie %q has value %q, espected %q", name, cv, value)
			} else {
				return
			}
		}
	}

	t.Errorf("No cookie with name %s", name)
}

func AssertNoCookie(t *testing.T, name string, recorder *httptest.ResponseRecorder) {
	t.Helper()

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == name {
			t.Errorf("Expected no cookie with name %s", name)
			return
		}
	}
}

func AssertResponseCode(t *testing.T, code int, recorder *httptest.ResponseRecorder) {
	t.Helper()
	if actual := recorder.Result().StatusCode; actual != code {
		t.Errorf("Expected response %d got %d", code, actual)
	}
}

func TestLoadBalancer_ServeHTTP(t *testing.T) {
	for tn, tc := range map[string]struct {
		request       *http.Request
		loadBalancer  *LoadBalancer
		checkResponse func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		"error LookupIP": {
			loadBalancer: &LoadBalancer{
				LookupIP: lookupIPErr,
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Result().StatusCode != http.StatusBadGateway {
					t.Errorf("Expected bad gateway response")
				}
			},
		},
		"empty LookupIP": {
			loadBalancer: &LoadBalancer{
				LookupIP: lookupZeroIP,
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Result().StatusCode != http.StatusBadGateway {
					t.Errorf("Expected bad gateway response")
				}
			},
		},
		"no session req/response": {
			loadBalancer: &LoadBalancer{
				StickyCookie: "__VCAP_ID__",
				LookupIP:     lookupOneIP,
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				AssertResponseCode(t, http.StatusOK, recorder)
				AssertNoCookie(t, "__VCAP_ID__", recorder)
			},
		},
		"first request with session": {
			loadBalancer: &LoadBalancer{
				SessionCookie: "SESSION",
				StickyCookie:  "STICKY",
				LookupIP: func(addr string) ([]net.IP, error) {
					return []net.IP{
						net.IPv4(8, 8, 8, 8),
					}, nil
				},
				ReverseProxy: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.SetCookie(w, &http.Cookie{Name: "SESSION", Value: "OK"})
					w.WriteHeader(http.StatusOK)
				}),
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				AssertResponseCode(t, http.StatusOK, recorder)
				AssertCookie(t, "STICKY", "8.8.8.8", recorder)
				AssertCookie(t, "SESSION", "OK", recorder)
			},
		},
		"maching request session": {
			loadBalancer: &LoadBalancer{
				SessionCookie: "SESSION",
				StickyCookie:  "STICKY",
				LookupIP: func(addr string) ([]net.IP, error) {
					return []net.IP{
						net.IPv4(1, 1, 1, 1),
						net.IPv4(8, 8, 8, 8),
					}, nil
				},
				RandIntn: func(int) int {
					// By default, pick the first IP
					return 0
				},
				ReverseProxy: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.SetCookie(w, &http.Cookie{Name: "SESSION", Value: r.URL.String()})
					w.WriteHeader(http.StatusOK)
				}),
			},
			request: (func() *http.Request {
				req := defaultRequest()
				req.AddCookie(&http.Cookie{Name: "STICKY", Value: "8.8.8.8"})
				return req
			}()),
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				AssertResponseCode(t, http.StatusOK, recorder)
				AssertCookie(t, "STICKY", "8.8.8.8", recorder)
				AssertCookie(t, "SESSION", "http://8.8.8.8:8080", recorder)
			},
		},
		"invalid sticky IP defaults": {
			loadBalancer: &LoadBalancer{
				SessionCookie: "SESSION",
				StickyCookie:  "STICKY",
				LookupIP: func(addr string) ([]net.IP, error) {
					return []net.IP{
						net.IPv4(1, 1, 1, 1),
					}, nil
				},
				RandIntn: func(int) int {
					// By default, pick the first IP
					return 0
				},
				ReverseProxy: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.SetCookie(w, &http.Cookie{Name: "SESSION", Value: r.URL.String()})
					w.WriteHeader(http.StatusOK)
				}),
			},
			request: (func() *http.Request {
				req := defaultRequest()
				req.AddCookie(&http.Cookie{Name: "STICKY", Value: "4.4.4.4"})
				return req
			}()),
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				AssertResponseCode(t, http.StatusOK, recorder)
				AssertCookie(t, "SESSION", "http://1.1.1.1:8080", recorder)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			defaultValues(t, tc.loadBalancer)
			request := defaultRequest()
			if tc.request != nil {
				request = tc.request
			}
			responseRecorder := httptest.NewRecorder()

			tc.loadBalancer.ServeHTTP(responseRecorder, request)

			tc.checkResponse(t, responseRecorder)
		})
	}

}
