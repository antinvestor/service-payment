package authz

import (
	"context"

	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/frame/security/authorizer"
)

type Middleware interface {
	CanManageLedger(ctx context.Context) error
	CanViewLedger(ctx context.Context) error
	CanManageAccount(ctx context.Context) error
	CanViewAccount(ctx context.Context) error
	CanCreateTransaction(ctx context.Context) error
	CanReverseTransaction(ctx context.Context) error
	CanUpdateTransaction(ctx context.Context) error
	CanViewTransaction(ctx context.Context) error
}

type middleware struct {
	checker *authorizer.FunctionChecker
}

func NewMiddleware(service security.Authorizer) Middleware {
	return &middleware{
		checker: authorizer.NewFunctionChecker(service, NamespaceLedger),
	}
}

func (m *middleware) CanManageLedger(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionManageLedger)
}

func (m *middleware) CanViewLedger(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionViewLedger)
}

func (m *middleware) CanManageAccount(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionManageAccount)
}

func (m *middleware) CanViewAccount(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionViewAccount)
}

func (m *middleware) CanCreateTransaction(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionCreateTransaction)
}

func (m *middleware) CanReverseTransaction(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionReverseTransaction)
}

func (m *middleware) CanUpdateTransaction(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionUpdateTransaction)
}

func (m *middleware) CanViewTransaction(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionViewTransaction)
}
