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
	authorizer security.Authorizer
}

func NewMiddleware(authorizer security.Authorizer) Middleware {
	return &middleware{authorizer: authorizer}
}

func (m *middleware) CanManageLedger(ctx context.Context) error {
	return m.check(ctx, PermissionManageLedger)
}

func (m *middleware) CanViewLedger(ctx context.Context) error {
	return m.check(ctx, PermissionViewLedger)
}

func (m *middleware) CanManageAccount(ctx context.Context) error {
	return m.check(ctx, PermissionManageAccount)
}

func (m *middleware) CanViewAccount(ctx context.Context) error {
	return m.check(ctx, PermissionViewAccount)
}

func (m *middleware) CanCreateTransaction(ctx context.Context) error {
	return m.check(ctx, PermissionCreateTransaction)
}

func (m *middleware) CanReverseTransaction(ctx context.Context) error {
	return m.check(ctx, PermissionReverseTransaction)
}

func (m *middleware) CanUpdateTransaction(ctx context.Context) error {
	return m.check(ctx, PermissionUpdateTransaction)
}

func (m *middleware) CanViewTransaction(ctx context.Context) error {
	return m.check(ctx, PermissionViewTransaction)
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
