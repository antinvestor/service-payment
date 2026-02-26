package authz

const (
	NamespaceTenant  = "ledger_tenant"
	NamespaceProfile = "profile"
)

const (
	PermissionManageLedger       = "manage_ledger"
	PermissionViewLedger         = "view_ledger"
	PermissionManageAccount      = "manage_account"
	PermissionViewAccount        = "view_account"
	PermissionCreateTransaction  = "create_transaction"
	PermissionReverseTransaction = "reverse_transaction"
	PermissionUpdateTransaction  = "update_transaction"
	PermissionViewTransaction    = "view_transaction"
)

const (
	RoleOwner    = "owner"
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)
