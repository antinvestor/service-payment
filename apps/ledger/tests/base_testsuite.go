// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tests

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	aconfig "github.com/antinvestor/service-payments/apps/ledger/config"
	"github.com/antinvestor/service-payments/apps/ledger/service/authz"
	"github.com/antinvestor/service-payments/apps/ledger/service/business"
	"github.com/antinvestor/service-payments/apps/ledger/service/repository"
	"github.com/antinvestor/service-payments/apps/ledger/tests/testketo"

	// Register PostgreSQL driver for database connections.
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame"
	"github.com/pitabwire/frame/config"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/frametests"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/pitabwire/frame/frametests/deps/testpostgres"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/require"
)

const (
	DefaultRandomStringLength = 8
)

type ServiceResources struct {
	LedgerRepository      repository.LedgerRepository
	AccountRepository     repository.AccountRepository
	TransactionRepository repository.TransactionRepository
	LedgerBusiness        business.LedgerBusiness
	AccountBusiness       business.AccountBusiness
	TransactionBusiness   business.TransactionBusiness
}

type BaseTestSuite struct {
	frametests.FrameBaseTestSuite
	ctx          context.Context
	resources    *ServiceResources
	ketoReadURI  string
	ketoWriteURI string
}

// ServiceResources returns the shared service dependencies for the test suite.
func (bs *BaseTestSuite) ServiceResources() *ServiceResources {
	// Create resources once and cache them to avoid unnecessary reinstantiation
	if bs.resources == nil {
		ctx, _, resources := bs.CreateService(bs.T(), nil)
		bs.ctx = ctx
		bs.resources = resources
	}
	return bs.resources
}

func initResources(_ context.Context) []definition.TestResource {
	pg := testpostgres.NewWithOpts("service_ledger", definition.WithUserName("ant"))
	keto := testketo.NewWithOpts(
		definition.WithDependancies(pg),
		definition.WithEnableLogging(true),
	)
	return []definition.TestResource{pg, keto}
}

func (bs *BaseTestSuite) SetupSuite() {
	bs.InitResourceFunc = initResources
	bs.FrameBaseTestSuite.SetupSuite()

	ctx := bs.T().Context()

	var ketoDep definition.DependancyConn
	for _, res := range bs.Resources() {
		if res.Name() == testketo.ImageName {
			ketoDep = res
			break
		}
	}
	bs.Require().NotNil(ketoDep, "keto dependency should be available")

	writeURL, err := url.Parse(string(ketoDep.GetDS(ctx)))
	bs.Require().NoError(err)
	bs.ketoWriteURI = writeURL.Host

	readPort, err := ketoDep.PortMapping(ctx, "4466/tcp")
	bs.Require().NoError(err)
	bs.ketoReadURI = fmt.Sprintf("%s:%s", writeURL.Hostname(), readPort)
}

func (bs *BaseTestSuite) CreateService(
	t *testing.T,
	depOpts *definition.DependencyOption,
	frameOpts ...frame.Option,
) (context.Context, *frame.Service, *ServiceResources) {
	ctx := t.Context()
	cfg, err := config.FromEnv[aconfig.LedgerConfig]()
	require.NoError(t, err)

	cfg.LogLevel = "debug"
	cfg.RunServiceSecurely = false
	cfg.ServerPort = ""
	cfg.DatabaseMigrate = true
	cfg.DatabaseTraceQueries = true

	res := depOpts.ByIsDatabase(ctx)
	testDS, cleanup, err0 := res.GetRandomisedDS(t.Context(), depOpts.Prefix())
	require.NoError(t, err0)

	t.Cleanup(func() {
		cleanup(t.Context())
	})

	cfg.DatabasePrimaryURL = []string{testDS.String()}
	cfg.DatabaseReplicaURL = []string{testDS.String()}

	cfg.AuthorizationServiceReadURI = bs.ketoReadURI
	cfg.AuthorizationServiceWriteURI = bs.ketoWriteURI

	frameOpts = append(
		[]frame.Option{
			frame.WithName("ledger tests"), frame.WithConfig(&cfg),
			frame.WithDatastore(), frametests.WithNoopDriver()}, frameOpts...)

	ctx, svc := frame.NewServiceWithContext(ctx, frameOpts...)

	dbManager := svc.DatastoreManager()
	dbPool := dbManager.GetPool(ctx, datastore.DefaultPoolName)
	workMan := svc.WorkManager()

	ledgerRepo := repository.NewLedgerRepository(ctx, dbPool, workMan)
	accountRepo := repository.NewAccountRepository(ctx, dbPool, workMan)
	transactionRepo := repository.NewTransactionRepository(ctx, dbPool, workMan, accountRepo)
	ledgerBusiness := business.NewLedgerBusiness(workMan, ledgerRepo, accountRepo)
	accountBusiness := business.NewAccountBusiness(workMan, ledgerRepo, accountRepo)
	transactionBusiness := business.NewTransactionBusiness(workMan, accountRepo, transactionRepo)

	resources := &ServiceResources{
		LedgerRepository:      ledgerRepo,
		AccountRepository:     accountRepo,
		TransactionRepository: transactionRepo,
		LedgerBusiness:        ledgerBusiness,
		AccountBusiness:       accountBusiness,
		TransactionBusiness:   transactionBusiness,
	}

	err = repository.Migrate(ctx, dbManager, "../../migrations/0001")
	require.NoError(t, err)

	err = svc.Run(ctx, "")
	require.NoError(t, err)

	return ctx, svc, resources
}

func (bs *BaseTestSuite) TearDownSuite() {
	bs.FrameBaseTestSuite.TearDownSuite()
}

// WithTestDependencies Creates subtests with each known DependancyOption.
func (bs *BaseTestSuite) WithTestDependencies(
	t *testing.T,
	testFn func(t *testing.T, dep *definition.DependencyOption),
) {
	options := []*definition.DependencyOption{
		definition.NewDependancyOption(
			"default",
			util.RandomAlphaNumericString(DefaultRandomStringLength),
			bs.Resources(),
		),
	}

	frametests.WithTestDependencies(t, options, testFn)
}

// WithAuthClaims creates a context with authentication claims for the given tenant, partition, and profile.
func (bs *BaseTestSuite) WithAuthClaims(ctx context.Context, tenantID, partitionID, profileID string) context.Context {
	claims := &security.AuthenticationClaims{
		TenantID:    tenantID,
		PartitionID: partitionID,
		AccessID:    util.IDString(),
		ContactID:   profileID,
		SessionID:   util.IDString(),
		DeviceID:    "test-device",
	}
	claims.Subject = profileID
	return claims.ClaimsToContext(ctx)
}

// SeedTenantAccess writes a tenancy_access member tuple so the profile can pass
// the TenancyAccessChecker (data access layer).
func (bs *BaseTestSuite) SeedTenantAccess(
	ctx context.Context,
	svc *frame.Service,
	tenantID, partitionID, profileID string,
) {
	auth := svc.SecurityManager().GetAuthorizer(ctx)
	tenancyPath := fmt.Sprintf("%s/%s", tenantID, partitionID)
	err := auth.WriteTuple(ctx, authz.BuildAccessTuple(tenancyPath, profileID))
	bs.Require().NoError(err, "failed to seed tenant access")
}

// SeedTenantRole writes functional permission tuples in the service_ledger
// namespace for the given role. Uses tenancyPath (tenantID/partitionID) as object ID.
func (bs *BaseTestSuite) SeedTenantRole(
	ctx context.Context,
	svc *frame.Service,
	tenantID, partitionID, profileID, role string,
) {
	auth := svc.SecurityManager().GetAuthorizer(ctx)
	tenancyPath := fmt.Sprintf("%s/%s", tenantID, partitionID)

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

	err := auth.WriteTuples(ctx, tuples)
	bs.Require().NoError(err, "failed to seed tenant role")
}
