package authz

const (
	NamespaceTenant  = "payment_tenant"
	NamespaceProfile = "profile"
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
)
