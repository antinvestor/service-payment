package authz

import (
	"context"

	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/frame/security/authorizer"
)

type Middleware interface {
	CanPaymentSend(ctx context.Context) error
	CanPaymentReceive(ctx context.Context) error
	CanPaymentsSearch(ctx context.Context) error
	CanPaymentStatusView(ctx context.Context) error
	CanPaymentStatusUpdate(ctx context.Context) error
	CanPaymentRelease(ctx context.Context) error
	CanPromptInitiate(ctx context.Context) error
	CanPaymentLinkCreate(ctx context.Context) error
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

func (m *middleware) CanPaymentSend(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionPaymentSend)
}

func (m *middleware) CanPaymentReceive(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionPaymentReceive)
}

func (m *middleware) CanPaymentsSearch(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionPaymentsSearch)
}

func (m *middleware) CanPaymentStatusView(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionPaymentStatusView)
}

func (m *middleware) CanPaymentStatusUpdate(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionPaymentStatusUpdate)
}

func (m *middleware) CanPaymentRelease(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionPaymentRelease)
}

func (m *middleware) CanPromptInitiate(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionPromptInitiate)
}

func (m *middleware) CanPaymentLinkCreate(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionPaymentLinkCreate)
}

func (m *middleware) CanReconcile(ctx context.Context) error {
	return m.checker.Check(ctx, PermissionReconcile)
}
