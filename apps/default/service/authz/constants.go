package authz

const (
	NamespacePayment       = "service_payment"
	NamespaceTenancyAccess = "tenancy_access"
	NamespaceProfile       = "profile_user"
)

const (
	PermissionPaymentSend         = "payment_send"
	PermissionPaymentReceive      = "payment_receive"
	PermissionPaymentsSearch      = "payments_search"
	PermissionPaymentStatusView   = "payment_status_view"
	PermissionPaymentStatusUpdate = "payment_status_update"
	PermissionPaymentRelease      = "payment_release"
	PermissionPromptInitiate      = "prompt_initiate"
	PermissionPaymentLinkCreate   = "payment_link_create"
	PermissionReconcile           = "reconcile"
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
// Direct grant relations are prefixed with "granted_" to avoid name conflicts
// with permit functions (Keto skips permit evaluation when a relation shares
// the same name as a permit function).
func GrantedRelation(permission string) string {
	return "granted_" + permission
}

// RolePermissions returns the permissions granted by each role.
func RolePermissions() map[string][]string {
	return map[string][]string{
		RoleOwner: {
			PermissionPaymentSend, PermissionPaymentReceive, PermissionPaymentsSearch,
			PermissionPaymentStatusView, PermissionPaymentStatusUpdate, PermissionPaymentRelease,
			PermissionPromptInitiate, PermissionPaymentLinkCreate, PermissionReconcile,
		},
		RoleAdmin: {
			PermissionPaymentSend, PermissionPaymentReceive, PermissionPaymentsSearch,
			PermissionPaymentStatusView, PermissionPaymentStatusUpdate, PermissionPaymentRelease,
			PermissionPromptInitiate, PermissionPaymentLinkCreate, PermissionReconcile,
		},
		RoleOperator: {
			PermissionPaymentSend, PermissionPaymentReceive, PermissionPaymentsSearch,
			PermissionPaymentStatusView, PermissionPaymentRelease,
			PermissionPromptInitiate, PermissionPaymentLinkCreate,
		},
		RoleViewer: {
			PermissionPaymentsSearch, PermissionPaymentStatusView,
		},
		RoleMember: {
			PermissionPaymentsSearch, PermissionPaymentStatusView,
		},
		RoleService: {
			PermissionPaymentSend, PermissionPaymentReceive, PermissionPaymentsSearch,
			PermissionPaymentStatusView, PermissionPaymentStatusUpdate, PermissionPaymentRelease,
			PermissionPromptInitiate, PermissionPaymentLinkCreate, PermissionReconcile,
		},
	}
}
