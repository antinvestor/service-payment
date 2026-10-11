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

package authz_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/ledger/service/authz"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// The OPL is generated from the LedgerService manifest; these files are the
// copies Keto is deployed with.
var oplFiles = []string{
	filepath.Join("..", "..", "..", "..", "opl", "ledger", "service_ledger.opl.ts"),
	filepath.Join("..", "..", "..", "default", "service_ledger.opl.ts"),
}

func ledgerManifest(t *testing.T) *commonv1.ServicePermissions {
	t.Helper()
	svc := ledgerv1.File_v1_ledger_proto.Services().ByName("LedgerService")
	require.NotNil(t, svc)
	manifest, ok := proto.GetExtension(svc.Options(), commonv1.E_ServicePermissions).(*commonv1.ServicePermissions)
	require.True(t, ok)
	require.Equal(t, authz.NamespaceLedger, manifest.GetNamespace())
	return manifest
}

// TestRolePermissionsFollowManifest proves the role mapping grants exactly
// the manifest's role bindings.
func TestRolePermissionsFollowManifest(t *testing.T) {
	manifest := ledgerManifest(t)
	roles := authz.RolePermissions()
	require.Len(t, roles, len(manifest.GetRoleBindings()))
	for _, binding := range manifest.GetRoleBindings() {
		role := strings.ToLower(strings.TrimPrefix(binding.GetRole().String(), "ROLE_"))
		require.ElementsMatch(t, binding.GetPermissions(), roles[role], "role %s", role)
	}
}

// TestOPLDeclaresEveryManifestPermission proves the shipped namespace has a
// permit and a granted_ relation for every permission the manifest declares
// and every RPC checks, so Keto never rejects a check as unknown.
func TestOPLDeclaresEveryManifestPermission(t *testing.T) {
	manifest := ledgerManifest(t)
	declared := map[string]bool{}
	for _, p := range manifest.GetPermissions() {
		declared[p] = true
	}

	methods := ledgerv1.File_v1_ledger_proto.Services().ByName("LedgerService").Methods()
	for i := range methods.Len() {
		m := methods.Get(i)
		perms, ok := proto.GetExtension(m.Options(), commonv1.E_MethodPermissions).(*commonv1.MethodPermissions)
		require.True(t, ok)
		for _, p := range perms.GetPermissions() {
			require.True(t, declared[p], "%s checks %s, which the manifest does not declare", m.Name(), p)
		}
	}

	var first string
	for _, path := range oplFiles {
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		opl := string(raw)
		for p := range declared {
			require.Regexp(t, regexp.MustCompile(`(?m)^\s+`+authz.GrantedRelation(p)+`: `), opl,
				"%s lacks the %s relation", path, authz.GrantedRelation(p))
			require.Regexp(t, regexp.MustCompile(`(?m)^\s+`+p+`: \(ctx: Context\)`), opl,
				"%s lacks the %s permit", path, p)
		}
		if first == "" {
			first = opl
			continue
		}
		require.Equal(t, first, opl, "the ledger OPL copies have diverged")
	}
}
