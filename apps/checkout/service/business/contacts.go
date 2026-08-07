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
	"strings"
	"unicode"

	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
)

// contactKind classifies a profile contact as email, phone, or unknown.
type contactKind int

const (
	contactKindUnknown contactKind = iota
	contactKindEmail
	contactKindPhone
)

// classifyContactDetail inspects a raw contact value (email or phone).
// Used when type is missing, wrong, or is the protobuf zero-value EMAIL (0).
func classifyContactDetail(detail string) contactKind {
	s := strings.TrimSpace(detail)
	if s == "" {
		return contactKindUnknown
	}
	// Email: must contain @ with a local and domain part (basic, not full RFC).
	if at := strings.IndexByte(s, '@'); at > 0 && at < len(s)-1 {
		domain := s[at+1:]
		if strings.Contains(domain, ".") && !strings.ContainsAny(s, " \t\n") {
			return contactKindEmail
		}
	}
	// Phone / MSISDN: mostly digits, optional leading +, spaces, dashes.
	digits := 0
	for _, r := range s {
		switch {
		case unicode.IsDigit(r):
			digits++
		case r == '+' || r == ' ' || r == '-' || r == '(' || r == ')':
			// allowed decoration
		default:
			// letters / other → not a pure phone
			return contactKindUnknown
		}
	}
	// E.164-ish: at least 7 digits (national short codes rarely used for pay).
	if digits >= 7 {
		return contactKindPhone
	}
	return contactKindUnknown
}

// classifyProfileContact prefers the declared ContactType, but corrects the
// common case where Type is unset (EMAIL=0 default) and detail is a phone
// number — or Type is MSISDN but detail is clearly an email.
func classifyProfileContact(c *profilev1.ContactObject) contactKind {
	if c == nil {
		return contactKindUnknown
	}
	detail := strings.TrimSpace(c.GetDetail())
	if detail == "" {
		return contactKindUnknown
	}
	shape := classifyContactDetail(detail)
	switch c.GetType() {
	case profilev1.ContactType_MSISDN:
		// Declared phone; only reclassify if detail is unmistakably email.
		if shape == contactKindEmail {
			return contactKindEmail
		}
		return contactKindPhone
	case profilev1.ContactType_EMAIL:
		// EMAIL is protobuf zero (0). Many stores omit type → defaults to EMAIL.
		// Trust shape when type and shape disagree.
		if shape == contactKindPhone {
			return contactKindPhone
		}
		if shape == contactKindEmail {
			return contactKindEmail
		}
		// Type says email but shape unclear — still treat as email if has @.
		if strings.Contains(detail, "@") {
			return contactKindEmail
		}
		return contactKindUnknown
	default:
		return shape
	}
}

// contactPick is a resolved email or phone contact with id for preference write-back.
type contactPick struct {
	ContactID string
	Detail    string
}

// pickEmailFromContacts chooses the preferred email contact when preferredID
// matches an email contact; otherwise the first email-shaped contact.
func pickEmailFromContacts(contacts []*profilev1.ContactObject, preferredID string) contactPick {
	var first contactPick
	for _, c := range contacts {
		if classifyProfileContact(c) != contactKindEmail {
			continue
		}
		detail := strings.TrimSpace(c.GetDetail())
		if detail == "" {
			continue
		}
		id := c.GetId()
		if preferredID != "" && id == preferredID {
			return contactPick{ContactID: id, Detail: detail}
		}
		if first.Detail == "" {
			first = contactPick{ContactID: id, Detail: detail}
		}
	}
	return first
}

// pickPhoneFromContacts chooses the preferred phone contact when preferredID
// matches a phone contact; otherwise the first phone-shaped contact.
func pickPhoneFromContacts(contacts []*profilev1.ContactObject, preferredID string) contactPick {
	var first contactPick
	for _, c := range contacts {
		if classifyProfileContact(c) != contactKindPhone {
			continue
		}
		detail := strings.TrimSpace(c.GetDetail())
		if detail == "" {
			continue
		}
		id := c.GetId()
		if preferredID != "" && id == preferredID {
			return contactPick{ContactID: id, Detail: detail}
		}
		if first.Detail == "" {
			first = contactPick{ContactID: id, Detail: detail}
		}
	}
	return first
}

// firstEmailFromContacts returns the first contact that is an email address.
func firstEmailFromContacts(contacts []*profilev1.ContactObject) string {
	return pickEmailFromContacts(contacts, "").Detail
}

// phoneContactsFromProfile returns phone-shaped contacts for the pay UI chips.
// preferredID is marked preferred and ordered first when present.
func phoneContactsFromProfile(contacts []*profilev1.ContactObject, preferredID string) []map[string]any {
	var preferred, rest []map[string]any
	for _, c := range contacts {
		if c == nil {
			continue
		}
		if classifyProfileContact(c) != contactKindPhone {
			continue
		}
		msisdn := strings.TrimSpace(c.GetDetail())
		if msisdn == "" {
			continue
		}
		id := c.GetId()
		m := map[string]any{
			"contactId": id,
			"msisdn":    msisdn,
		}
		if preferredID != "" && id == preferredID {
			m["preferred"] = true
			preferred = append(preferred, m)
			continue
		}
		rest = append(rest, m)
	}
	return append(preferred, rest...)
}

// normalizeCallerPhoneContacts keeps only phone-shaped caller contacts.
// Product may pass a raw identifier that is email or phone in Msisdn field.
func normalizeCallerPhoneContacts(in []PayerContactInput, preferredID string) []map[string]any {
	var preferred, rest []map[string]any
	for _, c := range in {
		raw := strings.TrimSpace(c.Msisdn)
		if raw == "" {
			continue
		}
		if classifyContactDetail(raw) != contactKindPhone {
			continue
		}
		m := map[string]any{
			"contactId": c.ContactID,
			"msisdn":    raw,
		}
		if preferredID != "" && c.ContactID == preferredID {
			m["preferred"] = true
			preferred = append(preferred, m)
			continue
		}
		rest = append(rest, m)
	}
	return append(preferred, rest...)
}

// firstEmailFromCallerContacts extracts an email if product put one in Msisdn/detail.
func firstEmailFromCallerContacts(in []PayerContactInput) string {
	for _, c := range in {
		raw := strings.TrimSpace(c.Msisdn)
		if classifyContactDetail(raw) == contactKindEmail {
			return raw
		}
	}
	return ""
}

// pickEmailFromCallerContacts returns email + id from caller contacts when possible.
func pickEmailFromCallerContacts(in []PayerContactInput, preferredID string) contactPick {
	var first contactPick
	for _, c := range in {
		raw := strings.TrimSpace(c.Msisdn)
		if classifyContactDetail(raw) != contactKindEmail {
			continue
		}
		if preferredID != "" && c.ContactID == preferredID {
			return contactPick{ContactID: c.ContactID, Detail: raw}
		}
		if first.Detail == "" {
			first = contactPick{ContactID: c.ContactID, Detail: raw}
		}
	}
	return first
}

// contactKindString is the stable wire value for prefill/UI (email|phone).
func contactKindString(k contactKind) string {
	switch k {
	case contactKindEmail:
		return "email"
	case contactKindPhone:
		return "phone"
	default:
		return ""
	}
}

// parseContactKind maps wire values back to contactKind.
func parseContactKind(s string) contactKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "email":
		return contactKindEmail
	case "phone", "msisdn":
		return contactKindPhone
	default:
		return contactKindUnknown
	}
}

// ContactKindFromString exposes kind parsing for handlers / method filtering.
func ContactKindFromString(s string) contactKind {
	return parseContactKind(s)
}

// unifiedProfileContacts builds the pay UI contact list from profile contacts
// only (email + phone). preferredIDs order preferred chips first.
// Products must not inject free-text payers when a profile is attached —
// only these contacts may charge.
func unifiedProfileContacts(
	contacts []*profilev1.ContactObject,
	preferredEmailID, preferredPhoneID, preferredAnyID string,
) []map[string]any {
	var preferred, rest []map[string]any
	for _, c := range contacts {
		if c == nil {
			continue
		}
		kind := classifyProfileContact(c)
		if kind != contactKindEmail && kind != contactKindPhone {
			continue
		}
		detail := strings.TrimSpace(c.GetDetail())
		if detail == "" {
			continue
		}
		id := c.GetId()
		m := map[string]any{
			"contactId": id,
			"detail":    detail,
			"kind":      contactKindString(kind),
		}
		if kind == contactKindPhone {
			m["msisdn"] = detail
		}
		isPref := false
		if kind == contactKindEmail && preferredEmailID != "" && id == preferredEmailID {
			isPref = true
		}
		if kind == contactKindPhone && preferredPhoneID != "" && id == preferredPhoneID {
			isPref = true
		}
		if preferredAnyID != "" && id == preferredAnyID {
			isPref = true
		}
		if isPref {
			m["preferred"] = true
			preferred = append(preferred, m)
			continue
		}
		rest = append(rest, m)
	}
	return append(preferred, rest...)
}

// prefillContact is a resolved entry from session prefill contacts[].
type prefillContact struct {
	ContactID string
	Detail    string
	Kind      contactKind
}

// findPrefillContact looks up contactID in session prefill contacts.
func findPrefillContact(prefill map[string]any, contactID string) (prefillContact, bool) {
	if prefill == nil || strings.TrimSpace(contactID) == "" {
		return prefillContact{}, false
	}
	contactsRaw, ok := prefill["contacts"]
	if !ok {
		return prefillContact{}, false
	}
	list, ok := contactsRaw.([]any)
	if !ok {
		return prefillContact{}, false
	}
	for _, raw := range list {
		m, isMap := raw.(map[string]any)
		if !isMap {
			continue
		}
		cid, _ := m["contactId"].(string)
		if cid == "" || cid != contactID {
			continue
		}
		detail, _ := m["detail"].(string)
		if detail == "" {
			detail, _ = m["msisdn"].(string)
		}
		detail = strings.TrimSpace(detail)
		if detail == "" {
			continue
		}
		kind := parseContactKind(fmtString(m["kind"]))
		if kind == contactKindUnknown {
			kind = classifyContactDetail(detail)
		}
		return prefillContact{ContactID: cid, Detail: detail, Kind: kind}, true
	}
	return prefillContact{}, false
}

func fmtString(v any) string {
	s, _ := v.(string)
	return s
}

// firstPrefillContactOfKind returns the first prefill contact matching kind.
func firstPrefillContactOfKind(prefill map[string]any, kind contactKind) (prefillContact, bool) {
	if prefill == nil {
		return prefillContact{}, false
	}
	contactsRaw, ok := prefill["contacts"]
	if !ok {
		return prefillContact{}, false
	}
	list, ok := contactsRaw.([]any)
	if !ok {
		return prefillContact{}, false
	}
	for _, raw := range list {
		m, isMap := raw.(map[string]any)
		if !isMap {
			continue
		}
		detail, _ := m["detail"].(string)
		if detail == "" {
			detail, _ = m["msisdn"].(string)
		}
		detail = strings.TrimSpace(detail)
		if detail == "" {
			continue
		}
		k := parseContactKind(fmtString(m["kind"]))
		if k == contactKindUnknown {
			k = classifyContactDetail(detail)
		}
		if k != kind {
			continue
		}
		cid, _ := m["contactId"].(string)
		return prefillContact{ContactID: cid, Detail: detail, Kind: k}, true
	}
	return prefillContact{}, false
}
