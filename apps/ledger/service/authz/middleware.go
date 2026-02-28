package authz

import (
	"context"

	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/frame/security/authorizer"
)

type Middleware interface {
	CanLedgerManage(ctx context.Context) error
	CanLedgerView(ctx context.Context) error
	CanAccountManage(ctx context.Context) error
	CanAccountView(ctx context.Context) error
	CanTransactionCreate(ctx context.Context) error
	CanTransactionReverse(ctx context.Context) error
	CanTransactionUpdate(ctx context.Context) error
	CanTransactionView(ctx context.Context) error
}

type middleware struct {
	checker *authorizer.FunctionChecker
}

func NewMiddleware(service security.Authorizer) Middleware {
	return &middleware{
		checker: authorizer.NewFunctionChecker(service, NamespaceLedger),
	}
}

func (m *middleware) CanLedgerManage(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionLedgerManage)
}

func (m *middleware) CanLedgerView(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionLedgerView)
}

func (m *middleware) CanAccountManage(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionAccountManage)
}

func (m *middleware) CanAccountView(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionAccountView)
}

func (m *middleware) CanTransactionCreate(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionTransactionCreate)
}

func (m *middleware) CanTransactionReverse(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionTransactionReverse)
}

func (m *middleware) CanTransactionUpdate(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionTransactionUpdate)
}

func (m *middleware) CanTransactionView(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionTransactionView)
}
