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
	"testing"

	aconfig "github.com/antinvestor/service-payments/apps/billing/config"
	"github.com/antinvestor/service-payments/apps/billing/service/business"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"

	// Ledger integration.
	ledgerBusiness "github.com/antinvestor/service-payments/apps/ledger/service/business"
	ledgerRepo "github.com/antinvestor/service-payments/apps/ledger/service/repository"
	"github.com/antinvestor/service-payments/internal/rlsadmin"

	// Register PostgreSQL driver for database connections.
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame/v2"
	"github.com/pitabwire/frame/v2/config"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/frametests"
	"github.com/pitabwire/frame/v2/frametests/definition"
	"github.com/pitabwire/frame/v2/frametests/deps/testpostgres"
	"github.com/pitabwire/frame/v2/frametests/rlstest"
	"github.com/pitabwire/frame/v2/security"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/require"
)

const (
	DefaultRandomStringLength = 8
)

type ServiceResources struct {
	CatalogBusiness      business.CatalogBusiness
	SubscriptionBusiness business.SubscriptionBusiness
	UsageIngestion       business.UsageIngestionBusiness
	MeteringEngine       business.MeteringEngine
	PricingEngine        *business.PricingEngine
	DiscountEngine       business.DiscountEngine
	CreditEngine         business.CreditEngine
	InvoiceEngine        business.InvoiceEngine
	BillingWorkflow      business.BillingWorkflow
	LedgerIntegration    business.LedgerIntegration

	// Repositories for direct access in tests
	ComponentRepo  repository.ComponentRepository
	BillingRunRepo repository.BillingRunRepository
	InvoiceRepo    repository.InvoiceRepository
}

type BaseTestSuite struct {
	frametests.FrameBaseTestSuite
	ctx       context.Context
	resources *ServiceResources
}

func (bs *BaseTestSuite) ServiceResources() *ServiceResources {
	if bs.resources == nil {
		ctx, _, resources := bs.CreateService(bs.T(), nil)
		bs.ctx = ctx
		bs.resources = resources
	}
	return bs.resources
}

func initResources(_ context.Context) []definition.TestResource {
	pg := testpostgres.NewWithOpts("service_billing", definition.WithUserName("ant"))
	return []definition.TestResource{pg}
}

func (bs *BaseTestSuite) SetupSuite() {
	bs.InitResourceFunc = initResources
	bs.FrameBaseTestSuite.SetupSuite()
}

//nolint:funlen // DI wiring function for test setup that is inherently long
func (bs *BaseTestSuite) CreateService(
	t *testing.T,
	depOpts *definition.DependencyOption,
	frameOpts ...frame.Option,
) (context.Context, *frame.Service, *ServiceResources) {
	ctx := t.Context()
	cfg, err := config.FromEnv[aconfig.BillingConfig]()
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

	// Drop application queries to an unprivileged role so Postgres RLS
	// is actually enforced (the testcontainer user is a superuser which
	// bypasses FORCE ROW LEVEL SECURITY).
	require.NoError(t, rlstest.CreateRole(ctx, testDS.String()))
	rlsProv := rlstest.New()

	frameOpts = append(
		[]frame.Option{
			frame.WithName("billing tests"), frame.WithConfig(&cfg),
			frame.WithTenancyProvider(rlsProv),
			frame.WithDatastore(), frametests.WithNoopDriver()}, frameOpts...)

	ctx, svc := frame.NewServiceWithContext(ctx, frameOpts...)
	t.Cleanup(func() { svc.Stop(ctx) })

	dbManager := svc.DatastoreManager()
	dbPool := dbManager.GetPool(ctx, datastore.DefaultPoolName)
	workMan := svc.WorkManager()

	// Billing repositories
	catalogVersionRepo := repository.NewCatalogVersionRepository(ctx, dbPool, workMan)
	planRepo := repository.NewPlanRepository(ctx, dbPool, workMan)
	componentRepo := repository.NewComponentRepository(ctx, dbPool, workMan)
	tierRepo := repository.NewTierRepository(ctx, dbPool, workMan)
	subscriptionRepo := repository.NewSubscriptionRepository(ctx, dbPool, workMan)
	usageEventRepo := repository.NewUsageEventRepository(ctx, dbPool, workMan)
	meteredUsageRepo := repository.NewMeteredUsageRepository(ctx, dbPool, workMan)
	discountRepo := repository.NewDiscountRepository(ctx, dbPool, workMan)
	discountedLineRepo := repository.NewDiscountedLineRepository(ctx, dbPool, workMan)
	creditGrantRepo := repository.NewCreditGrantRepository(ctx, dbPool, workMan)
	creditEntryRepo := repository.NewCreditEntryRepository(ctx, dbPool, workMan)
	ratedLineRepo := repository.NewRatedLineRepository(ctx, dbPool, workMan)
	invoiceRepo := repository.NewInvoiceRepository(ctx, dbPool, workMan)
	invoiceLineRepo := repository.NewInvoiceLineRepository(ctx, dbPool, workMan)
	billingRunRepo := repository.NewBillingRunRepository(ctx, dbPool, workMan)

	// Ledger repos for integration
	_ = ledgerRepo.NewLedgerRepository(ctx, dbPool, workMan)
	lAccountRepo := ledgerRepo.NewAccountRepository(ctx, dbPool, workMan)
	lTransactionRepo := ledgerRepo.NewTransactionRepository(ctx, dbPool, workMan, lAccountRepo)
	ledgerTxnBusiness := ledgerBusiness.NewTransactionBusiness(workMan, lAccountRepo, lTransactionRepo)

	// Business layers
	catalogBus := business.NewCatalogBusiness(workMan, catalogVersionRepo, planRepo, componentRepo, tierRepo)
	subscriptionBus := business.NewSubscriptionBusiness(workMan, subscriptionRepo)
	usageIngestionBus := business.NewUsageIngestionBusiness(workMan, usageEventRepo)
	meteringEng := business.NewMeteringEngine(workMan, usageEventRepo, meteredUsageRepo)
	pricingEng := business.NewPricingEngine()
	discountEng := business.NewDiscountEngine(workMan, discountRepo, discountedLineRepo)
	creditEng := business.NewCreditEngine(workMan, dbPool, creditGrantRepo, creditEntryRepo)
	invoiceEng := business.NewInvoiceEngine(workMan, dbPool, invoiceRepo, invoiceLineRepo)
	ledgerInteg := business.NewLedgerIntegration(ledgerTxnBusiness)
	billingWorkflow := business.NewBillingWorkflow(
		workMan, billingRunRepo, ratedLineRepo, subscriptionBus, catalogBus, componentRepo,
		meteringEng, pricingEng, discountEng, creditEng, invoiceEng, ledgerInteg, nil)

	resources := &ServiceResources{
		CatalogBusiness:      catalogBus,
		SubscriptionBusiness: subscriptionBus,
		UsageIngestion:       usageIngestionBus,
		MeteringEngine:       meteringEng,
		PricingEngine:        pricingEng,
		DiscountEngine:       discountEng,
		CreditEngine:         creditEng,
		InvoiceEngine:        invoiceEng,
		BillingWorkflow:      billingWorkflow,
		LedgerIntegration:    ledgerInteg,
		ComponentRepo:        componentRepo,
		BillingRunRepo:       billingRunRepo,
		InvoiceRepo:          invoiceRepo,
	}

	// Run both ledger and billing migrations
	err = ledgerRepo.Migrate(ctx, dbManager, "../../ledger/migrations/0001")
	require.NoError(t, err)

	err = repository.Migrate(ctx, dbManager, "../../billing/migrations/0001")
	require.NoError(t, err)

	require.NoError(t, rlstest.GrantAll(ctx, testDS.String()))
	require.NoError(t, rlsadmin.GrantOwnership(ctx, testDS.String()))
	rlsProv.Enable()

	err = svc.Run(ctx, "")
	require.NoError(t, err)

	return ctx, svc, resources
}

func (bs *BaseTestSuite) TearDownSuite() {
	bs.FrameBaseTestSuite.TearDownSuite()
}

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
