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
	authorizer security.Authorizer
}

func NewMiddleware(authorizer security.Authorizer) Middleware {
	return &middleware{authorizer: authorizer}
}

func (m *middleware) CanSendPayment(ctx context.Context) error {
	return m.check(ctx, PermissionSendPayment)
}

func (m *middleware) CanReceivePayment(ctx context.Context) error {
	return m.check(ctx, PermissionReceivePayment)
}

func (m *middleware) CanSearchPayments(ctx context.Context) error {
	return m.check(ctx, PermissionSearchPayments)
}

func (m *middleware) CanViewPaymentStatus(ctx context.Context) error {
	return m.check(ctx, PermissionViewPaymentStatus)
}

func (m *middleware) CanUpdatePaymentStatus(ctx context.Context) error {
	return m.check(ctx, PermissionUpdatePaymentStatus)
}

func (m *middleware) CanReleasePayment(ctx context.Context) error {
	return m.check(ctx, PermissionReleasePayment)
}

func (m *middleware) CanInitiatePrompt(ctx context.Context) error {
	return m.check(ctx, PermissionInitiatePrompt)
}

func (m *middleware) CanCreatePaymentLink(ctx context.Context) error {
	return m.check(ctx, PermissionCreatePaymentLink)
}

func (m *middleware) CanReconcile(ctx context.Context) error {
	return m.check(ctx, PermissionReconcile)
}

func (m *middleware) check(ctx context.Context, permission string) error {
	claims := security.ClaimsFromContext(ctx)
	if claims == nil {
		return authorizer.ErrInvalidSubject
	}

	subjectID, err := claims.GetSubject()
	if err != nil || subjectID == "" {
		return authorizer.ErrInvalidSubject
	}

	tenantID := claims.GetTenantID()
	if tenantID == "" {
		return authorizer.ErrInvalidObject
	}

	req := security.CheckRequest{
		Object:     security.ObjectRef{Namespace: NamespaceTenant, ID: tenantID},
		Permission: permission,
		Subject:    security.SubjectRef{Namespace: NamespaceProfile, ID: subjectID},
	}

	result, err := m.authorizer.Check(ctx, req)
	if err != nil {
		return err
	}
	if !result.Allowed {
		return authorizer.NewPermissionDeniedError(req.Object, permission, req.Subject, result.Reason)
	}

	return nil
}
