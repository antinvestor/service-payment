package authz

const (
	NamespaceLedger        = "service_ledger"
	NamespaceTenancyAccess = "tenancy_access"
	NamespaceProfile       = "profile/user"
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
	RoleMember   = "member"
	RoleService  = "service"
)

// RolePermissions returns the permissions granted by each role.
func RolePermissions() map[string][]string {
	return map[string][]string{
		RoleOwner: {
			PermissionManageLedger, PermissionViewLedger,
			PermissionManageAccount, PermissionViewAccount,
			PermissionCreateTransaction, PermissionReverseTransaction,
			PermissionUpdateTransaction, PermissionViewTransaction,
		},
		RoleAdmin: {
			PermissionManageLedger, PermissionViewLedger,
			PermissionManageAccount, PermissionViewAccount,
			PermissionCreateTransaction, PermissionReverseTransaction,
			PermissionUpdateTransaction, PermissionViewTransaction,
		},
		RoleOperator: {
			PermissionViewLedger,
			PermissionManageAccount, PermissionViewAccount,
			PermissionCreateTransaction, PermissionViewTransaction,
		},
		RoleViewer: {
			PermissionViewLedger, PermissionViewAccount, PermissionViewTransaction,
		},
		RoleMember: {
			PermissionViewLedger, PermissionViewAccount, PermissionViewTransaction,
		},
		RoleService: {
			PermissionManageLedger, PermissionViewLedger,
			PermissionManageAccount, PermissionViewAccount,
			PermissionCreateTransaction, PermissionReverseTransaction,
			PermissionUpdateTransaction, PermissionViewTransaction,
		},
	}
}
