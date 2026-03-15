package authz_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/antinvestor/service-payments/apps/ledger/service/authz"
	"github.com/antinvestor/service-payments/apps/ledger/tests/testketo"
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
		Object:   security.ObjectRef{Namespace: authz.NamespaceLedger, ID: tenancyPath},
		Relation: role,
		Subject:  security.SubjectRef{Namespace: authz.NamespaceProfile, ID: profileID},
	})

	for _, perm := range permissions {
		tuples = append(tuples, security.RelationTuple{
			Object:   security.ObjectRef{Namespace: authz.NamespaceLedger, ID: tenancyPath},
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

	s.NoError(mw.CanLedgerManage(ctx))
	s.NoError(mw.CanLedgerView(ctx))
	s.NoError(mw.CanAccountManage(ctx))
	s.NoError(mw.CanAccountView(ctx))
	s.NoError(mw.CanTransactionCreate(ctx))
	s.NoError(mw.CanTransactionReverse(ctx))
	s.NoError(mw.CanTransactionUpdate(ctx))
	s.NoError(mw.CanTransactionView(ctx))
}

func (s *MiddlewareTestSuite) TestOperatorPermissions() {
	auth := s.newAuthorizer()
	s.seedRole(auth, testTenancyPath(), "user2", authz.RoleOperator)

	mw := authz.NewMiddleware(auth)
	ctx := s.ctxWithClaims("user2")

	s.NoError(mw.CanLedgerView(ctx))
	s.NoError(mw.CanAccountManage(ctx))
	s.NoError(mw.CanAccountView(ctx))
	s.NoError(mw.CanTransactionCreate(ctx))
	s.NoError(mw.CanTransactionView(ctx))

	// Operator cannot manage ledger, reverse/update transactions
	s.Require().Error(mw.CanLedgerManage(ctx))
	s.Require().Error(mw.CanTransactionReverse(ctx))
	s.Require().Error(mw.CanTransactionUpdate(ctx))
}

func (s *MiddlewareTestSuite) TestViewerPermissions() {
	auth := s.newAuthorizer()
	s.seedRole(auth, testTenancyPath(), "user3", authz.RoleViewer)

	mw := authz.NewMiddleware(auth)
	ctx := s.ctxWithClaims("user3")

	s.NoError(mw.CanLedgerView(ctx))
	s.NoError(mw.CanAccountView(ctx))
	s.NoError(mw.CanTransactionView(ctx))

	s.Require().Error(mw.CanLedgerManage(ctx))
	s.Require().Error(mw.CanAccountManage(ctx))
	s.Require().Error(mw.CanTransactionCreate(ctx))
	s.Require().Error(mw.CanTransactionReverse(ctx))
	s.Require().Error(mw.CanTransactionUpdate(ctx))
}

func (s *MiddlewareTestSuite) TestNoClaims() {
	auth := s.newAuthorizer()
	mw := authz.NewMiddleware(auth)

	err := mw.CanLedgerView(context.Background())
	s.ErrorIs(err, authorizer.ErrInvalidSubject)
}

func (s *MiddlewareTestSuite) TestNoTenant() {
	auth := s.newAuthorizer()
	mw := authz.NewMiddleware(auth)

	claims := &security.AuthenticationClaims{}
	claims.Subject = "user1"
	ctx := claims.ClaimsToContext(context.Background())
	err := mw.CanLedgerView(ctx)
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
	s.NoError(mw.CanLedgerManage(botCtx))
	s.NoError(mw.CanTransactionCreate(botCtx))
	s.NoError(mw.CanTransactionView(botCtx))
}
