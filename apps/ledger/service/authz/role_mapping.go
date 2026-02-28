package authz

import "github.com/pitabwire/frame/security"

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
func BuildServiceInheritanceTuples(tenancyPath string) []security.RelationTuple {
	serviceBridge := security.RelationTuple{
		Object:   security.ObjectRef{Namespace: NamespaceLedger, ID: tenancyPath},
		Relation: RoleService,
		Subject: security.SubjectRef{
			Namespace: NamespaceTenancyAccess,
			ID:        tenancyPath,
			Relation:  RoleService,
		},
	}

	permissions := RolePermissions()[RoleService]
	tuples := make([]security.RelationTuple, 0, 1+len(permissions))
	tuples = append(tuples, serviceBridge)

	for _, perm := range permissions {
		tuples = append(tuples, security.RelationTuple{
			Object:   security.ObjectRef{Namespace: NamespaceLedger, ID: tenancyPath},
			Relation: GrantedRelation(perm),
			Subject: security.SubjectRef{
				Namespace: NamespaceLedger,
				ID:        tenancyPath,
				Relation:  RoleService,
			},
		})
	}

	return tuples
}
