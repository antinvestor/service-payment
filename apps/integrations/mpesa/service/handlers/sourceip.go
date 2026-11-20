// Copyright 2023-2026 Ant Investor Ltd
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

package handlers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/pitabwire/util"
)

// SourceGuard admits callbacks only from an allowlist of source CIDRs. With
// an empty list every source is admitted (and nothing counts as verified).
type SourceGuard struct {
	nets []*net.IPNet
	hops int
}

// NewSourceGuard parses a comma-separated CIDR list. hops is the number of
// X-Forwarded-For entries appended by our own proxies; 0 uses RemoteAddr.
func NewSourceGuard(cidrs string, hops int) (*SourceGuard, error) {
	g := &SourceGuard{hops: hops}
	for _, raw := range strings.Split(cidrs, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if !strings.Contains(raw, "/") {
			raw += "/32"
			if strings.Contains(raw, ":") {
				raw = strings.TrimSuffix(raw, "/32") + "/128"
			}
		}
		_, n, err := net.ParseCIDR(raw)
		if err != nil {
			return nil, fmt.Errorf("parse callback allowlist %q: %w", raw, err)
		}
		g.nets = append(g.nets, n)
	}
	return g, nil
}

// Configured reports whether an allowlist is in force.
func (g *SourceGuard) Configured() bool { return g != nil && len(g.nets) > 0 }

// ClientIP is the caller's address: RemoteAddr, or with trusted proxy hops
// the X-Forwarded-For entry our first proxy saw. Fewer entries than hops
// means the chain was not ours, so RemoteAddr is used (fails closed).
func (g *SourceGuard) ClientIP(r *http.Request) net.IP {
	remote := r.RemoteAddr
	if host, _, err := net.SplitHostPort(remote); err == nil {
		remote = host
	}
	if g == nil || g.hops <= 0 {
		return net.ParseIP(remote)
	}
	var entries []string
	for _, v := range r.Header.Values("X-Forwarded-For") {
		for _, e := range strings.Split(v, ",") {
			if e = strings.TrimSpace(e); e != "" {
				entries = append(entries, e)
			}
		}
	}
	if len(entries) < g.hops {
		return net.ParseIP(remote)
	}
	return net.ParseIP(entries[len(entries)-g.hops])
}

// Allowed reports whether the caller is inside the allowlist. An empty
// allowlist admits everyone.
func (g *SourceGuard) Allowed(r *http.Request) bool {
	if !g.Configured() {
		return true
	}
	ip := g.ClientIP(r)
	if ip == nil {
		return false
	}
	for _, n := range g.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

type sourceVerifiedKey struct{}

// Wrap refuses callers outside the allowlist with 403 and marks admitted
// requests as source-verified when an allowlist is configured.
func (g *SourceGuard) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !g.Allowed(r) {
			util.Log(r.Context()).WithField("remote", r.RemoteAddr).
				WithField("client_ip", g.ClientIP(r)).Warn("M-Pesa callback refused: source not in allowlist")
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		ctx := r.Context()
		if g.Configured() {
			ctx = context.WithValue(ctx, sourceVerifiedKey{}, true)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// sourceVerified reports whether the request passed a configured allowlist.
func sourceVerified(ctx context.Context) bool {
	v, _ := ctx.Value(sourceVerifiedKey{}).(bool)
	return v
}
