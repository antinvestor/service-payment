package authz

import (
	"context"

	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/frame/security/authorizer"
)

type Middleware interface {
	CanSendPayment(ctx context.Context) error
	CanReceivePayment(ctx context.Context) error
	CanSearchPayments(ctx context.Context) error
	CanViewPaymentStatus(ctx context.Context) error
	CanUpdatePaymentStatus(ctx context.Context) error
	CanReleasePayment(ctx context.Context) error
	CanInitiatePrompt(ctx context.Context) error
	CanCreatePaymentLink(ctx context.Context) error
	CanReconcile(ctx context.Context) error
}

type middleware struct {
	checker *authorizer.FunctionChecker
}

func NewMiddleware(service security.Authorizer) Middleware {
	return &middleware{
		checker: authorizer.NewFunctionChecker(service, NamespacePayment),
	}
}

func (m *middleware) CanSendPayment(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionSendPayment)
}

func (m *middleware) CanReceivePayment(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionReceivePayment)
}

func (m *middleware) CanSearchPayments(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionSearchPayments)
}

func (m *middleware) CanViewPaymentStatus(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionViewPaymentStatus)
}

func (m *middleware) CanUpdatePaymentStatus(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionUpdatePaymentStatus)
}

func (m *middleware) CanReleasePayment(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionReleasePayment)
}

func (m *middleware) CanInitiatePrompt(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionInitiatePrompt)
}

func (m *middleware) CanCreatePaymentLink(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionCreatePaymentLink)
}

func (m *middleware) CanReconcile(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionReconcile)
}
