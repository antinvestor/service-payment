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

package business_test

import (
	"testing"

	"github.com/antinvestor/service-payments/apps/checkout/service/business"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestProfileDisplayName_AuNameCanonical(t *testing.T) {
	t.Parallel()
	props, err := structpb.NewStruct(map[string]any{
		business.ProfilePropName: "Auth Name",
		"name":                   "Legacy",
		"given_name":             "Given",
		"family_name":            "Family",
	})
	require.NoError(t, err)
	assert.Equal(t, "Auth Name", business.ProfileDisplayName(props))
}

func TestProfileDisplayName_LegacyAndGivenFamily(t *testing.T) {
	t.Parallel()
	legacy, err := structpb.NewStruct(map[string]any{"name": "  Legacy Bob  "})
	require.NoError(t, err)
	assert.Equal(t, "Legacy Bob", business.ProfileDisplayName(legacy))

	split, err := structpb.NewStruct(map[string]any{
		"given_name":  "Ada",
		"family_name": "Lovelace",
	})
	require.NoError(t, err)
	assert.Equal(t, "Ada Lovelace", business.ProfileDisplayName(split))

	assert.Equal(t, "", business.ProfileDisplayName(nil))
	empty, _ := structpb.NewStruct(map[string]any{})
	assert.Equal(t, "", business.ProfileDisplayName(empty))
}

func TestResolveDisplayName_ProfileBeatsCaller(t *testing.T) {
	t.Parallel()
	props, err := structpb.NewStruct(map[string]any{business.ProfilePropName: "Real User"})
	require.NoError(t, err)
	assert.Equal(t, "Real User", business.ResolveDisplayName("Merchant Label", props))
	assert.Equal(t, "Merchant Label", business.ResolveDisplayName("Merchant Label", nil))
	assert.Equal(t, "Merchant Label", business.ResolveDisplayName("  Merchant Label  ", emptyProps(t)))
}

func emptyProps(t *testing.T) *structpb.Struct {
	t.Helper()
	p, err := structpb.NewStruct(map[string]any{"other": "x"})
	require.NoError(t, err)
	return p
}
