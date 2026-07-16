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

package business

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Method describes one renderable payment method and its prompt route hint.
//
// Availability (Link-style) is driven by:
//   - Currencies: ISO 4217 codes this method accepts (empty = all)
//   - Prefixes: E.164 country/network prefixes for MSISDN locality (e.g. "254")
//   - Countries: ISO 3166-1 alpha-2 country codes (e.g. "KE") for geo headers
//   - Redirect: provider-hosted page (legacy); prefer Embed for cards
//   - Embed: collect instrument on our page (Stripe Link style for cards)
type Method struct {
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	Route      string   `json:"route"`
	Prefixes   []string `json:"prefixes"`
	Currencies []string `json:"currencies"`
	Countries  []string `json:"countries"`
	Redirect   bool     `json:"redirect"`
	Embed      bool     `json:"embed"` // card form on pay.stawi.org
}

// IsEmbedded reports whether this method collects payment details on our domain.
func (m Method) IsEmbedded() bool {
	return m.Embed || (!m.Redirect && strings.EqualFold(m.Key, "card"))
}

// MethodRegistry is the configured catalog of payment methods for this service.
type MethodRegistry struct {
	Methods []Method
}

// MethodFilter is the Link-style context used to decide which methods appear
// and which one is preselected.
//
// Preselect priority (highest first):
//  1. Cached / previously used method (profile clue or guest cookie)
//  2. Phone MSISDN locality (prefix match)
//  3. Country / geo locality
//  4. First non-redirect method, else first available
//
// Availability is the intersection of:
//   - global registry
//   - partition allowlist (if configured)
//   - session restriction (merchant per-session methods[])
//   - currency
//   - locality when signals exist (universal methods without locality
//     constraints always remain, e.g. card)
type MethodFilter struct {
	Currency           string
	Phone              string   // E.164 or national digits
	Country            string   // ISO 3166-1 alpha-2
	SessionRestriction []string // CreateCheckoutSession methods[]
	PartitionAllowlist []string // partition-configured method keys
	ClueMethod         string   // profile lastProvider / lastMethod
	GuestMethod        string   // device cookie last method
}

// MethodResolution is the filtered, ordered method list plus the preselected key.
type MethodResolution struct {
	Available []Method
	Selected  Method
	// Reason documents why Selected was chosen (tests / observability).
	Reason string
}

// ParseMethodRegistry parses the CHECKOUT_METHODS config JSON.
func ParseMethodRegistry(raw string) (*MethodRegistry, error) {
	var methods []Method
	if err := json.Unmarshal([]byte(raw), &methods); err != nil {
		return nil, fmt.Errorf("parse method registry: %w", err)
	}
	if len(methods) == 0 {
		return nil, errors.New("method registry is empty")
	}
	return &MethodRegistry{Methods: methods}, nil
}

// PartitionAllowlists maps partition ID → allowed method keys.
// Special key "*" is the default allowlist when a partition has no entry.
type PartitionAllowlists map[string][]string

// ParsePartitionAllowlists parses CHECKOUT_PARTITION_METHODS JSON.
// Format: {"partition-id": ["mpesa","card"], "*": ["card"]}
// Empty string returns an empty map (no partition filtering).
func ParsePartitionAllowlists(raw string) (PartitionAllowlists, error) {
	if strings.TrimSpace(raw) == "" {
		return PartitionAllowlists{}, nil
	}
	var m map[string][]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, fmt.Errorf("parse partition method allowlists: %w", err)
	}
	out := make(PartitionAllowlists, len(m))
	for k, v := range m {
		out[strings.TrimSpace(k)] = v
	}
	return out, nil
}

// ForPartition returns the allowlist for a partition, falling back to "*".
// A nil return means "no partition restriction".
func (p PartitionAllowlists) ForPartition(partitionID string) []string {
	if len(p) == 0 {
		return nil
	}
	if partitionID != "" {
		if keys, ok := p[partitionID]; ok {
			return keys
		}
	}
	if keys, ok := p["*"]; ok {
		return keys
	}
	return nil
}

// Get returns the method for a key, or false.
func (r *MethodRegistry) Get(key string) (Method, bool) {
	for _, m := range r.Methods {
		if m.Key == key {
			return m, true
		}
	}
	return Method{}, false
}

// Available is a thin wrapper kept for backward-compatible call sites/tests.
// Prefer Resolve for full Link-style selection.
func (r *MethodRegistry) Available(restriction []string, currency ...string) []Method {
	cur := ""
	if len(currency) > 0 {
		cur = currency[0]
	}
	res := r.Resolve(MethodFilter{
		Currency:           cur,
		SessionRestriction: restriction,
	})
	return res.Available
}

// Resolve filters and ranks methods for a checkout session (Stripe Link style).
func (r *MethodRegistry) Resolve(f MethodFilter) MethodResolution {
	phone := normalizePhone(f.Phone)
	country := strings.ToUpper(strings.TrimSpace(f.Country))
	if country == "" {
		country = InferCountryFromPhone(phone)
	}
	currency := strings.ToUpper(strings.TrimSpace(f.Currency))

	allow := intersectAllowlists(f.SessionRestriction, f.PartitionAllowlist)
	hasLocality := phone != "" || country != ""

	var available []Method
	for _, m := range r.Methods {
		if len(allow) > 0 && !containsFold(allow, m.Key) {
			continue
		}
		if currency != "" && len(m.Currencies) > 0 && !methodSupportsCurrency(m, currency) {
			continue
		}
		// When locality is known, drop methods that declare a locality and do
		// not match — keep universal methods (no prefixes and no countries).
		if hasLocality && !methodMatchesLocality(m, phone, country) && !methodIsUniversal(m) {
			continue
		}
		available = append(available, m)
	}

	// Traveler fallback: if locality wiped the list, re-open currency/allowlist set.
	if len(available) == 0 && hasLocality {
		for _, m := range r.Methods {
			if len(allow) > 0 && !containsFold(allow, m.Key) {
				continue
			}
			if currency != "" && len(m.Currencies) > 0 && !methodSupportsCurrency(m, currency) {
				continue
			}
			available = append(available, m)
		}
	}

	if len(available) == 0 {
		return MethodResolution{}
	}

	available = rankMethods(available, f.ClueMethod, f.GuestMethod, phone, country)
	selected, reason := preselect(available, f.ClueMethod, f.GuestMethod, phone, country)
	return MethodResolution{
		Available: available,
		Selected:  selected,
		Reason:    reason,
	}
}

// Preselect is kept for tests; uses the same ranking rules as Resolve.
func Preselect(methods []Method, clueKey, phoneNumber string) Method {
	if len(methods) == 0 {
		return Method{}
	}
	phone := normalizePhone(phoneNumber)
	selected, _ := preselect(methods, clueKey, "", phone, InferCountryFromPhone(phone))
	return selected
}

func preselect(methods []Method, clueKey, guestKey, phone, country string) (Method, string) {
	if len(methods) == 0 {
		return Method{}, "empty"
	}
	// 1) Cached profile clue (Stripe Link "last used")
	if clueKey != "" {
		if m, ok := findMethod(methods, clueKey); ok {
			return m, "cached_profile"
		}
	}
	// 2) Guest device cookie
	if guestKey != "" {
		if m, ok := findMethod(methods, guestKey); ok {
			return m, "cached_device"
		}
	}
	// 3) Phone locality (MSISDN prefix)
	if phone != "" {
		for _, m := range methods {
			if m.Redirect {
				continue
			}
			if methodMatchesPhone(m, phone) {
				return m, "location_phone"
			}
		}
	}
	// 4) Country geo
	if country != "" {
		for _, m := range methods {
			if methodMatchesCountry(m, country) {
				return m, "location_country"
			}
		}
	}
	// 5) Prefer embedded card (universal), then non-redirect local rails, else first
	for _, m := range methods {
		if m.IsEmbedded() {
			return m, "default_embed_card"
		}
	}
	for _, m := range methods {
		if !m.Redirect {
			return m, "default_local"
		}
	}
	return methods[0], "default_first"
}

func rankMethods(methods []Method, clue, guest, phone, country string) []Method {
	score := func(m Method) int {
		s := 0
		if clue != "" && strings.EqualFold(m.Key, clue) {
			s += 100
		}
		if guest != "" && strings.EqualFold(m.Key, guest) {
			s += 80
		}
		if phone != "" && methodMatchesPhone(m, phone) {
			s += 40
		}
		if country != "" && methodMatchesCountry(m, country) {
			s += 30
		}
		// Prefer embedded card over redirect/hosted when tied.
		if m.IsEmbedded() {
			s += 15
		}
		if m.Redirect {
			s--
		}
		return s
	}
	out := make([]Method, len(methods))
	copy(out, methods)
	for i := 1; i < len(out); i++ {
		j := i
		for j > 0 && score(out[j]) > score(out[j-1]) {
			out[j], out[j-1] = out[j-1], out[j]
			j--
		}
	}
	return out
}

func findMethod(methods []Method, key string) (Method, bool) {
	for _, m := range methods {
		if strings.EqualFold(m.Key, key) {
			return m, true
		}
	}
	return Method{}, false
}

func methodSupportsCurrency(m Method, currency string) bool {
	for _, c := range m.Currencies {
		if strings.EqualFold(c, currency) {
			return true
		}
	}
	return false
}

func methodIsUniversal(m Method) bool {
	return len(m.Prefixes) == 0 && len(m.Countries) == 0
}

func methodMatchesLocality(m Method, phone, country string) bool {
	if phone != "" && methodMatchesPhone(m, phone) {
		return true
	}
	if country != "" && methodMatchesCountry(m, country) {
		return true
	}
	return false
}

func methodMatchesPhone(m Method, phone string) bool {
	if phone == "" || len(m.Prefixes) == 0 {
		return false
	}
	for _, p := range m.Prefixes {
		p = strings.TrimPrefix(strings.TrimSpace(p), "+")
		if p != "" && strings.HasPrefix(phone, p) {
			return true
		}
	}
	return false
}

func methodMatchesCountry(m Method, country string) bool {
	if country == "" || len(m.Countries) == 0 {
		return false
	}
	for _, c := range m.Countries {
		if strings.EqualFold(c, country) {
			return true
		}
	}
	return false
}

func normalizePhone(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "+")
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// InferCountryFromPhone maps common E.164 prefixes to ISO alpha-2.
func InferCountryFromPhone(phone string) string {
	phone = normalizePhone(phone)
	if phone == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(phone, "254"):
		return "KE"
	case strings.HasPrefix(phone, "255"):
		return "TZ"
	case strings.HasPrefix(phone, "256"):
		return "UG"
	case strings.HasPrefix(phone, "250"):
		return "RW"
	case strings.HasPrefix(phone, "260"):
		return "ZM"
	case strings.HasPrefix(phone, "233"):
		return "GH"
	case strings.HasPrefix(phone, "234"):
		return "NG"
	case strings.HasPrefix(phone, "27"):
		return "ZA"
	case strings.HasPrefix(phone, "1"):
		return "US"
	case strings.HasPrefix(phone, "44"):
		return "GB"
	default:
		return ""
	}
}

// DetectCountryFromHeaders reads common edge geo headers.
func DetectCountryFromHeaders(getHeader func(string) string) string {
	if getHeader == nil {
		return ""
	}
	for _, h := range []string{
		"CF-IPCountry",
		"CloudFront-Viewer-Country",
		"X-Country-Code",
		"X-Geo-Country",
		"X-Appengine-Country",
	} {
		v := strings.ToUpper(strings.TrimSpace(getHeader(h)))
		if v == "" || v == "XX" || v == "T1" {
			continue
		}
		if len(v) == 2 {
			return v
		}
	}
	return ""
}

func intersectAllowlists(a, b []string) []string {
	a = cleanKeys(a)
	b = cleanKeys(b)
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	set := make(map[string]bool, len(b))
	for _, k := range b {
		set[strings.ToLower(k)] = true
	}
	var out []string
	for _, k := range a {
		if set[strings.ToLower(k)] {
			out = append(out, k)
		}
	}
	return out
}

func cleanKeys(keys []string) []string {
	var out []string
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k != "" {
			out = append(out, k)
		}
	}
	return out
}

func containsFold(keys []string, key string) bool {
	for _, k := range keys {
		if strings.EqualFold(k, key) {
			return true
		}
	}
	return false
}
