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
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// SafeRemoteTransport returns a remote.Option that installs an HTTP transport
// which refuses to dial private, loopback, link-local, multicast, unspecified,
// or RFC6598 (CGNAT) addresses. It is intended to be passed to every
// remote.Image / remote.Get / remote.Index call that operates on a
// tenant-controlled image reference.
//
// The transport blocks SSRF via attacker-controlled WWW-Authenticate Bearer
// realm URLs (the same class of bug fixed upstream in
// google/go-containerregistry#2243) and via DNS rebinding, because the address
// check runs in the dialer Control hook after DNS resolution and before the
// kernel connects.
//
// The kf pin of go-containerregistry predates the upstream fix by ~4 years and
// a clean dep bump pulls invasive OpenTelemetry / k8s.io / knative / tekton
// updates. This file provides defense-in-depth without that upgrade.
func SafeRemoteTransport() remote.Option {
	return remote.WithTransport(SafeTransport())
}

// SafeTransport returns an *http.Transport identical to http.DefaultTransport
// except that its Dialer rejects connections to non-public IP addresses.
func SafeTransport() *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   rejectInternalAddress,
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// rejectInternalAddress is a Dialer.Control hook. The kernel has already
// performed DNS resolution by this point, so the address argument contains an
// IP literal (not a hostname). This makes the check DNS-rebinding safe.
func rejectInternalAddress(network, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("kf: cannot parse dial address %q: %w", address, err)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("kf: dial address %q is not an IP literal", host)
	}
	if isBlockedIP(ip) {
		return fmt.Errorf("kf: refused to dial blocked address %s (%s/%s)", ip, network, address)
	}
	return nil
}

// isBlockedIP reports whether ip is in a range that should never be the target
// of a registry HTTP request driven by a tenant-supplied image reference.
func isBlockedIP(ip net.IP) bool {
	if ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() ||
		ip.IsPrivate() {
		return true
	}
	// RFC6598 CGNAT 100.64.0.0/10. Not covered by IsPrivate().
	if v4 := ip.To4(); v4 != nil {
		if v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
			return true
		}
	}
	return false
}
