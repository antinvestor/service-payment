package authz

const (
	NamespacePayment       = "service_payment"
	NamespaceTenancyAccess = "tenancy_access"
	NamespaceProfile       = "profile/user"
)

const (
	PermissionSendPayment         = "send_payment"
	PermissionReceivePayment      = "receive_payment"
	PermissionSearchPayments      = "search_payments"
	PermissionViewPaymentStatus   = "view_payment_status"
	PermissionUpdatePaymentStatus = "update_payment_status"
	PermissionReleasePayment      = "release_payment"
	PermissionInitiatePrompt      = "initiate_prompt"
	PermissionCreatePaymentLink   = "create_payment_link"
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

// RolePermissions returns the permissions granted by each role.
func RolePermissions() map[string][]string {
	return map[string][]string{
		RoleOwner: {
			PermissionSendPayment, PermissionReceivePayment, PermissionSearchPayments,
			PermissionViewPaymentStatus, PermissionUpdatePaymentStatus, PermissionReleasePayment,
			PermissionInitiatePrompt, PermissionCreatePaymentLink, PermissionReconcile,
		},
		RoleAdmin: {
			PermissionSendPayment, PermissionReceivePayment, PermissionSearchPayments,
			PermissionViewPaymentStatus, PermissionUpdatePaymentStatus, PermissionReleasePayment,
			PermissionInitiatePrompt, PermissionCreatePaymentLink, PermissionReconcile,
		},
		RoleOperator: {
			PermissionSendPayment, PermissionReceivePayment, PermissionSearchPayments,
			PermissionViewPaymentStatus, PermissionReleasePayment,
			PermissionInitiatePrompt, PermissionCreatePaymentLink,
		},
		RoleViewer: {
			PermissionSearchPayments, PermissionViewPaymentStatus,
		},
		RoleMember: {
			PermissionSearchPayments, PermissionViewPaymentStatus,
		},
		RoleService: {
			PermissionSendPayment, PermissionReceivePayment, PermissionSearchPayments,
			PermissionViewPaymentStatus, PermissionUpdatePaymentStatus, PermissionReleasePayment,
			PermissionInitiatePrompt, PermissionCreatePaymentLink, PermissionReconcile,
		},
	}
}
