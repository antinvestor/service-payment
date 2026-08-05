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

import (
	"context"

	"github.com/pitabwire/frame/v2/security"
	"github.com/pitabwire/util"
)

const (
	NamespacePayment       = "service_payment"
	NamespaceTenancyAccess = "tenancy_access"
	NamespaceProfile       = "profile_user"

	RoleMember  = "member"
	RoleService = "service"
)

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

// HealServiceTenancyAccess provisions a missing Plane-1 #service tuple when an
// internal system caller is denied tenancy access. Wire with
// authorizer.WithOnTenancyAccessDenied(authz.HealServiceTenancyAccess).
func HealServiceTenancyAccess(
	ctx context.Context,
	auth security.Authorizer,
	tenancyPath, subjectID string,
) error {
	fields := map[string]any{
		"tenant_id":  tenancyPath,
		"subject_id": subjectID,
	}
	claims := security.ClaimsFromContext(ctx)
	if claims == nil || !claims.IsInternalSystem() {
		util.Log(ctx).WithFields(fields).Error("PERMISSION DENIED: tenancy access denied")
		return nil
	}
	if err := auth.WriteTuple(ctx, BuildServiceAccessTuple(tenancyPath, subjectID)); err != nil {
		util.Log(ctx).WithFields(fields).WithError(err).
			Error("PERMISSION DENIED: self-heal of tenancy service access failed")
		return err
	}
	util.Log(ctx).WithFields(fields).
		Info("self-healed missing tenancy service tuple for internal caller")
	return nil
}

// BuildServiceInheritanceTuples creates bridge tuples that allow service bots
// to inherit functional permissions via subject sets.
// Only the bridge tuple is needed — the OPL permits already check the service
// role directly, so explicit granted_* tuples per permission are redundant.
func BuildServiceInheritanceTuples(tenancyPath string) []security.RelationTuple {
	return []security.RelationTuple{{
		Object:   security.ObjectRef{Namespace: NamespacePayment, ID: tenancyPath},
		Relation: RoleService,
		Subject: security.SubjectRef{
			Namespace: NamespaceTenancyAccess,
			ID:        tenancyPath,
			Relation:  RoleService,
		},
	}}
}
