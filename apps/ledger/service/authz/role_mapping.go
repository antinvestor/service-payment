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

package authz

import "github.com/pitabwire/frame/v2/security"

const (
	NamespaceLedger        = "service_ledger"
	NamespaceTenancyAccess = "tenancy_access"
	NamespaceProfile       = "profile_user"
)

const (
	PermissionLedgerManage       = "ledger_manage"
	PermissionLedgerView         = "ledger_view"
	PermissionAccountManage      = "account_manage"
	PermissionAccountView        = "account_view"
	PermissionTransactionCreate  = "transaction_create"
	PermissionTransactionReverse = "transaction_reverse"
	PermissionTransactionUpdate  = "transaction_update"
	PermissionTransactionView    = "transaction_view"
)

const (
	RoleOwner    = "owner"
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
	RoleMember   = "member"
	RoleService  = "service"
)

// GrantedRelation returns the OPL relation name used for direct permission grants.
func GrantedRelation(permission string) string {
	return "granted_" + permission
}

// RolePermissions returns the permissions granted by each role.
func RolePermissions() map[string][]string {
	return map[string][]string{
		RoleOwner: {
			PermissionLedgerManage, PermissionLedgerView,
			PermissionAccountManage, PermissionAccountView,
			PermissionTransactionCreate, PermissionTransactionReverse,
			PermissionTransactionUpdate, PermissionTransactionView,
		},
		RoleAdmin: {
			PermissionLedgerManage, PermissionLedgerView,
			PermissionAccountManage, PermissionAccountView,
			PermissionTransactionCreate, PermissionTransactionReverse,
			PermissionTransactionUpdate, PermissionTransactionView,
		},
		RoleOperator: {
			PermissionLedgerView,
			PermissionAccountManage, PermissionAccountView,
			PermissionTransactionCreate, PermissionTransactionView,
		},
		RoleViewer: {
			PermissionLedgerView, PermissionAccountView, PermissionTransactionView,
		},
		RoleMember: {
			PermissionLedgerView, PermissionAccountView, PermissionTransactionView,
		},
		RoleService: {
			PermissionLedgerManage, PermissionLedgerView,
			PermissionAccountManage, PermissionAccountView,
			PermissionTransactionCreate, PermissionTransactionReverse,
			PermissionTransactionUpdate, PermissionTransactionView,
		},
	}
}

// BuildAccessTuple creates a tenancy_access#member tuple for a user.
func BuildAccessTuple(tenancyPath, profileID string) security.RelationTuple {
	return security.RelationTuple{
		Object:   security.ObjectRef{Namespace: NamespaceTenancyAccess, ID: tenancyPath},
		Relation: RoleMember,
		Subject:  security.SubjectRef{Namespace: NamespaceProfile, ID: profileID},
	}
}

// BuildServiceAccessTuple creates a tenancy_access#service tuple for a service bot.
func BuildServiceAccessTuple(tenancyPath, profileID string) security.RelationTuple {
	return security.RelationTuple{
		Object:   security.ObjectRef{Namespace: NamespaceTenancyAccess, ID: tenancyPath},
		Relation: RoleService,
		Subject:  security.SubjectRef{Namespace: NamespaceProfile, ID: profileID},
	}
}

// BuildServiceInheritanceTuples creates bridge tuples that allow service bots
// to inherit functional permissions via subject sets.
// Only the bridge tuple is needed — the OPL permits already check the service
// role directly, so explicit granted_* tuples per permission are redundant.
func BuildServiceInheritanceTuples(tenancyPath string) []security.RelationTuple {
	return []security.RelationTuple{{
		Object:   security.ObjectRef{Namespace: NamespaceLedger, ID: tenancyPath},
		Relation: RoleService,
		Subject: security.SubjectRef{
			Namespace: NamespaceTenancyAccess,
			ID:        tenancyPath,
			Relation:  RoleService,
		},
	}}
}
