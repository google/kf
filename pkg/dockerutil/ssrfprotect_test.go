// Copyright 2026 Google LLC
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

package dockerutil

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsBlockedIP(t *testing.T) {
	cases := []struct {
		ip      string
		blocked bool
	}{
		{"127.0.0.1", true},
		{"127.5.5.5", true},
		{"169.254.169.254", true}, // GCE / EC2 / Azure metadata
		{"169.254.170.2", true},   // ECS metadata
		{"fe80::1", true},          // IPv6 link-local
		{"::1", true},              // IPv6 loopback
		{"::", true},               // unspecified
		{"0.0.0.0", true},          // unspecified
		{"10.0.0.1", true},         // RFC1918
		{"10.255.255.254", true},
		{"172.16.0.1", true},       // RFC1918
		{"172.31.255.254", true},
		{"192.168.0.1", true},      // RFC1918
		{"fc00::1", true},          // ULA
		{"100.64.0.1", true},       // CGNAT (RFC6598)
		{"100.127.255.254", true},
		{"224.0.0.1", true},        // multicast
		{"ff02::1", true},          // IPv6 multicast
		{"1.1.1.1", false},
		{"8.8.8.8", false},
		{"172.15.0.1", false},      // just outside RFC1918 lower bound
		{"172.32.0.1", false},      // just outside RFC1918 upper bound
		{"100.63.255.254", false},  // just outside CGNAT lower bound
		{"100.128.0.1", false},     // just outside CGNAT upper bound
		{"2001:4860:4860::8888", false}, // Google Public DNS v6
	}
	for _, tc := range cases {
		ip := net.ParseIP(tc.ip)
		if ip == nil {
			t.Errorf("net.ParseIP(%q) returned nil", tc.ip)
			continue
		}
		if got := isBlockedIP(ip); got != tc.blocked {
			t.Errorf("isBlockedIP(%s) = %v, want %v", tc.ip, got, tc.blocked)
		}
	}
}

func TestSafeTransport_BlocksLoopback(t *testing.T) {
	// Spin up a localhost server. SafeTransport should refuse to connect to it.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("SafeTransport must not reach loopback handler, got request: %s %s", r.Method, r.URL)
	}))
	defer srv.Close()

	client := &http.Client{Transport: SafeTransport()}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
		t.Fatalf("expected SafeTransport to refuse loopback dial, got status %s", resp.Status)
	}
	if !strings.Contains(err.Error(), "refused to dial blocked address") {
		t.Errorf("expected blocked-address error, got: %v", err)
	}
}

func TestSafeTransport_AllowsPublicHost(t *testing.T) {
	// Synthesize a request whose host resolves to a public IP. We cannot
	// actually call out from a hermetic test environment, so we verify only
	// that the dialer Control hook accepts the post-resolution address.
	cases := []string{"1.1.1.1:443", "8.8.8.8:53", "[2001:4860:4860::8888]:53"}
	for _, addr := range cases {
		if err := rejectInternalAddress("tcp", addr, nil); err != nil {
			t.Errorf("rejectInternalAddress(tcp, %q) = %v, want nil", addr, err)
		}
	}
}

func TestRejectInternalAddress_MalformedAddress(t *testing.T) {
	if err := rejectInternalAddress("tcp", "not-an-addr", nil); err == nil {
		t.Errorf("expected error for malformed address, got nil")
	}
}

func TestRejectInternalAddress_HostnameInsteadOfIP(t *testing.T) {
	// The kernel resolves DNS before calling the Control hook, so the address
	// should always be an IP literal. If for some reason it is not, reject it
	// rather than allow an unverified connection.
	if err := rejectInternalAddress("tcp", "evil.example.com:443", nil); err == nil {
		t.Errorf("expected error for hostname-in-address, got nil")
	}
}
