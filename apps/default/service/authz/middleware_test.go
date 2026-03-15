package authz_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/antinvestor/service-payments/apps/default/service/authz"
	"github.com/antinvestor/service-payments/apps/default/tests/testketo"
	"github.com/pitabwire/frame/config"
	"github.com/pitabwire/frame/frametests"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/pitabwire/frame/frametests/deps/testpostgres"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/frame/security/authorizer"
	"github.com/stretchr/testify/suite"
)

const (
	testTenantID    = "tenant1"
	testPartitionID = "partition1"
)

func testTenancyPath() string { return testTenantID + "/" + testPartitionID }

type MiddlewareTestSuite struct {
	frametests.FrameBaseTestSuite
	ketoReadURI  string
	ketoWriteURI string
}

func initMiddlewareResources(_ context.Context) []definition.TestResource {
	pg := testpostgres.NewWithOpts("authz_middleware_test",
		definition.WithUserName("ant"),
		definition.WithCredential("s3cr3t"),
		definition.WithEnableLogging(false),
		definition.WithUseHostMode(false),
	)
	keto := testketo.NewWithOpts(
		definition.WithDependancies(pg),
		definition.WithEnableLogging(false),
	)
	return []definition.TestResource{pg, keto}
}

func (s *MiddlewareTestSuite) SetupSuite() {
	s.InitResourceFunc = initMiddlewareResources
	s.FrameBaseTestSuite.SetupSuite()

	ctx := s.T().Context()
	var ketoDep definition.DependancyConn
	for _, res := range s.Resources() {
		if res.Name() == testketo.ImageName {
			ketoDep = res
			break
		}
	}
	s.Require().NotNil(ketoDep, "keto dependency should be available")

	writeURL, err := url.Parse(string(ketoDep.GetDS(ctx)))
	s.Require().NoError(err)
	s.ketoWriteURI = writeURL.Host

	readPort, err := ketoDep.PortMapping(ctx, "4466/tcp")
	s.Require().NoError(err)
	s.ketoReadURI = fmt.Sprintf("%s:%s", writeURL.Hostname(), readPort)
}

func (s *MiddlewareTestSuite) newAuthorizer() security.Authorizer {
	cfg := &config.ConfigurationDefault{
		AuthorizationServiceReadURI:  s.ketoReadURI,
		AuthorizationServiceWriteURI: s.ketoWriteURI,
	}
	return authorizer.NewKetoAdapter(cfg, nil)
}

func (s *MiddlewareTestSuite) ctxWithClaims(subjectID string) context.Context {
	claims := &security.AuthenticationClaims{
		TenantID:    testTenantID,
		PartitionID: testPartitionID,
	}
	claims.Subject = subjectID
	return claims.ClaimsToContext(context.Background())
}

func (s *MiddlewareTestSuite) ctxWithSystemInternalClaims(subjectID string) context.Context {
	claims := &security.AuthenticationClaims{
		TenantID:    testTenantID,
		PartitionID: testPartitionID,
		Roles:       []string{"internal"},
	}
	claims.Subject = subjectID
	return claims.ClaimsToContext(context.Background())
}

func (s *MiddlewareTestSuite) seedRole(auth security.Authorizer, tenancyPath, profileID, role string) {
	permissions := authz.RolePermissions()[role]
	tuples := make([]security.RelationTuple, 0, 1+len(permissions))

	tuples = append(tuples, security.RelationTuple{
		Object:   security.ObjectRef{Namespace: authz.NamespacePayment, ID: tenancyPath},
		Relation: role,
		Subject:  security.SubjectRef{Namespace: authz.NamespaceProfile, ID: profileID},
	})

	for _, perm := range permissions {
		tuples = append(tuples, security.RelationTuple{
			Object:   security.ObjectRef{Namespace: authz.NamespacePayment, ID: tenancyPath},
			Relation: authz.GrantedRelation(perm),
			Subject:  security.SubjectRef{Namespace: authz.NamespaceProfile, ID: profileID},
		})
	}

	err := auth.WriteTuples(s.T().Context(), tuples)
	s.Require().NoError(err)
}

func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

func (s *MiddlewareTestSuite) TestOwnerHasAllPermissions() {
	auth := s.newAuthorizer()
	s.seedRole(auth, testTenancyPath(), "user1", authz.RoleOwner)

	mw := authz.NewMiddleware(auth)
	ctx := s.ctxWithClaims("user1")

	s.NoError(mw.CanPaymentSend(ctx))
	s.NoError(mw.CanPaymentReceive(ctx))
	s.NoError(mw.CanPaymentsSearch(ctx))
	s.NoError(mw.CanPaymentStatusView(ctx))
	s.NoError(mw.CanPaymentStatusUpdate(ctx))
	s.NoError(mw.CanPaymentRelease(ctx))
	s.NoError(mw.CanPromptInitiate(ctx))
	s.NoError(mw.CanPaymentLinkCreate(ctx))
	s.NoError(mw.CanReconcile(ctx))
}

func (s *MiddlewareTestSuite) TestOperatorPermissions() {
	auth := s.newAuthorizer()
	s.seedRole(auth, testTenancyPath(), "user2", authz.RoleOperator)

	mw := authz.NewMiddleware(auth)
	ctx := s.ctxWithClaims("user2")

	s.NoError(mw.CanPaymentSend(ctx))
	s.NoError(mw.CanPaymentReceive(ctx))
	s.NoError(mw.CanPaymentsSearch(ctx))
	s.NoError(mw.CanPaymentStatusView(ctx))
	s.NoError(mw.CanPaymentRelease(ctx))
	s.NoError(mw.CanPromptInitiate(ctx))
	s.NoError(mw.CanPaymentLinkCreate(ctx))

	// Operator cannot update status or reconcile
	s.Require().Error(mw.CanPaymentStatusUpdate(ctx))
	s.Require().Error(mw.CanReconcile(ctx))
}

func (s *MiddlewareTestSuite) TestViewerPermissions() {
	auth := s.newAuthorizer()
	s.seedRole(auth, testTenancyPath(), "user3", authz.RoleViewer)

	mw := authz.NewMiddleware(auth)
	ctx := s.ctxWithClaims("user3")

	s.NoError(mw.CanPaymentsSearch(ctx))
	s.NoError(mw.CanPaymentStatusView(ctx))

	s.Require().Error(mw.CanPaymentSend(ctx))
	s.Require().Error(mw.CanPaymentReceive(ctx))
	s.Require().Error(mw.CanPaymentStatusUpdate(ctx))
	s.Require().Error(mw.CanPaymentRelease(ctx))
	s.Require().Error(mw.CanReconcile(ctx))
}

func (s *MiddlewareTestSuite) TestNoClaims() {
	auth := s.newAuthorizer()
	mw := authz.NewMiddleware(auth)

	err := mw.CanPaymentsSearch(context.Background())
	s.ErrorIs(err, authorizer.ErrInvalidSubject)
}

func (s *MiddlewareTestSuite) TestNoTenant() {
	auth := s.newAuthorizer()
	mw := authz.NewMiddleware(auth)

	claims := &security.AuthenticationClaims{}
	claims.Subject = "user1"
	ctx := claims.ClaimsToContext(context.Background())
	err := mw.CanPaymentsSearch(ctx)
	s.ErrorIs(err, authorizer.ErrInvalidObject)
}

func (s *MiddlewareTestSuite) TestAccessChecker_MemberAllowed() {
	auth := s.newAuthorizer()
	checker := authorizer.NewTenancyAccessChecker(auth, authz.NamespaceTenancyAccess)

	err := auth.WriteTuple(s.T().Context(), authz.BuildAccessTuple(testTenancyPath(), "member-user"))
	s.Require().NoError(err)

	ctx := s.ctxWithClaims("member-user")
	s.NoError(checker.CheckAccess(ctx))
}

func (s *MiddlewareTestSuite) TestAccessChecker_ServiceBotAllowed() {
	auth := s.newAuthorizer()
	checker := authorizer.NewTenancyAccessChecker(auth, authz.NamespaceTenancyAccess)

	err := auth.WriteTuple(s.T().Context(), authz.BuildServiceAccessTuple(testTenancyPath(), "bot-user"))
	s.Require().NoError(err)

	ctx := s.ctxWithSystemInternalClaims("bot-user")
	s.NoError(checker.CheckAccess(ctx))
}

func (s *MiddlewareTestSuite) TestAccessChecker_NoTupleDenied() {
	auth := s.newAuthorizer()
	checker := authorizer.NewTenancyAccessChecker(auth, authz.NamespaceTenancyAccess)

	ctx := s.ctxWithClaims("unknown-user")
	s.Require().Error(checker.CheckAccess(ctx))
}

func (s *MiddlewareTestSuite) seedServiceBridgeTuples(auth security.Authorizer, tenancyPath string) {
	tuples := authz.BuildServiceInheritanceTuples(tenancyPath)
	err := auth.WriteTuples(s.T().Context(), tuples)
	s.Require().NoError(err)
}

func (s *MiddlewareTestSuite) TestServiceBotViaSubjectSets() {
	auth := s.newAuthorizer()
	mw := authz.NewMiddleware(auth)
	accessChecker := authorizer.NewTenancyAccessChecker(auth, authz.NamespaceTenancyAccess)

	s.seedServiceBridgeTuples(auth, testTenancyPath())

	err := auth.WriteTuple(s.T().Context(), authz.BuildServiceAccessTuple(testTenancyPath(), "service-bot"))
	s.Require().NoError(err)

	botCtx := s.ctxWithSystemInternalClaims("service-bot")

	s.NoError(accessChecker.CheckAccess(botCtx))
	s.NoError(mw.CanPaymentSend(botCtx))
	s.NoError(mw.CanPaymentsSearch(botCtx))
	s.NoError(mw.CanReconcile(botCtx))
}
