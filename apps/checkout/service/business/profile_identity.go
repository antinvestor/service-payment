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

	"google.golang.org/protobuf/types/known/structpb"
)

// Profile property keys for identity. service_authentication writes the display
// name as au_name on create/update (see KeyProfileName in service-authentication).
// Checkout must prefer that key so pay pages always reflect the authenticated user.
const (
	// ProfilePropName is the canonical display name set by service_authentication.
	ProfilePropName = "au_name"
	// ProfilePropAvatarURL is set by auth avatar sync (optional UI).
	ProfilePropAvatarURL = "au_avatar_url"
	// ProfilePropAvatarFileID is the files-service id for the avatar.
	ProfilePropAvatarFileID = "au_avatar_file_id"
)

// legacyProfileNameKeys are older / alternate property keys sometimes written
// by products or older clients. Checked only after au_name.
var legacyProfileNameKeys = []string{
	"name",
	"display_name",
	"displayName",
	"full_name",
	"fullName",
}

// ProfileDisplayName extracts the user's display name from profile properties.
//
// Precedence (first non-empty wins):
//  1. au_name — service_authentication SoT
//  2. legacy single-field name keys (name, display_name, …)
//  3. given_name + family_name (or givenName + familyName)
//
// Values are trimmed; whitespace-only is treated as empty.
func ProfileDisplayName(props *structpb.Struct) string {
	if props == nil {
		return ""
	}
	fields := props.GetFields()
	if fields == nil {
		return ""
	}

	// 1) Canonical auth identity.
	if n := stringField(fields, ProfilePropName); n != "" {
		return n
	}

	// 2) Legacy single-string name fields.
	for _, k := range legacyProfileNameKeys {
		if n := stringField(fields, k); n != "" {
			return n
		}
	}

	// 3) Given + family (OIDC-style splits).
	given := firstNonEmpty(
		stringField(fields, "given_name"),
		stringField(fields, "givenName"),
	)
	family := firstNonEmpty(
		stringField(fields, "family_name"),
		stringField(fields, "familyName"),
	)
	return strings.TrimSpace(strings.TrimSpace(given) + " " + strings.TrimSpace(family))
}

// ResolveDisplayName chooses the prefill display name for a checkout session.
//
// When a profile is loaded successfully, au_name (and other profile identity
// fields) win over a merchant-supplied DisplayName so the pay page always
// shows the authenticated user from service_authentication — not a stale or
// product-local label. Caller name is only used when the profile has no name.
func ResolveDisplayName(callerName string, props *structpb.Struct) string {
	if n := ProfileDisplayName(props); n != "" {
		return n
	}
	return strings.TrimSpace(callerName)
}

func stringField(fields map[string]*structpb.Value, key string) string {
	v, ok := fields[key]
	if !ok || v == nil {
		return ""
	}
	// Prefer string; tolerate number-like only if someone stored name wrong.
	switch k := v.GetKind().(type) {
	case *structpb.Value_StringValue:
		return strings.TrimSpace(k.StringValue)
	default:
		return ""
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
