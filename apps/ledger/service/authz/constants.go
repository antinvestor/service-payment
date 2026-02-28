package authz

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
