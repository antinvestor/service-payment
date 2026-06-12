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

	aconfig "github.com/antinvestor/service-payments/apps/checkout/config"
	"github.com/antinvestor/service-payments/apps/checkout/service/repository"
	"github.com/antinvestor/service-payments/internal/rlsadmin"

	// Register PostgreSQL driver for database connections.
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame"
	"github.com/pitabwire/frame/config"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/frametests"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/pitabwire/frame/frametests/deps/testpostgres"
	"github.com/pitabwire/frame/frametests/rlstest"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/require"
)

const (
	DefaultRandomStringLength = 8
)

// ServiceResources holds the shared service dependencies for the test suite.
type ServiceResources struct {
	SessionRepository repository.SessionRepository
	LinkRepository    repository.LinkRepository
}

// BaseTestSuite is the frametests-backed suite for checkout repository tests.
type BaseTestSuite struct {
	frametests.FrameBaseTestSuite
	ctx       context.Context
	resources *ServiceResources
}

// ServiceResources returns cached service dependencies, creating them once.
func (bs *BaseTestSuite) ServiceResources() *ServiceResources {
	if bs.resources == nil {
		ctx, _, resources := bs.CreateService(bs.T(), nil)
		bs.ctx = ctx
		bs.resources = resources
	}
	return bs.resources
}

func initResources(_ context.Context) []definition.TestResource {
	pg := testpostgres.NewWithOpts("service_checkout", definition.WithUserName("ant"))
	return []definition.TestResource{pg}
}

func (bs *BaseTestSuite) SetupSuite() {
	bs.InitResourceFunc = initResources
	bs.FrameBaseTestSuite.SetupSuite()
}

func (bs *BaseTestSuite) CreateService(
	t *testing.T,
	depOpts *definition.DependencyOption,
	frameOpts ...frame.Option,
) (context.Context, *frame.Service, *ServiceResources) {
	ctx := t.Context()
	cfg, err := config.FromEnv[aconfig.CheckoutConfig]()
	require.NoError(t, err)

	cfg.LogLevel = "debug"
	cfg.RunServiceSecurely = false
	cfg.ServerPort = ""
	cfg.DatabaseMigrate = true
	cfg.DatabaseTraceQueries = true

	if depOpts == nil {
		depOpts = definition.NewDependancyOption(
			"default",
			util.RandomAlphaNumericString(DefaultRandomStringLength),
			bs.Resources(),
		)
	}
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
			frame.WithName("checkout tests"), frame.WithConfig(&cfg),
			frame.WithTenancyProvider(rlsProv),
			frame.WithDatastore(), frametests.WithNoopDriver()}, frameOpts...)

	ctx, svc := frame.NewServiceWithContext(ctx, frameOpts...)
	t.Cleanup(func() { svc.Stop(ctx) })

	dbManager := svc.DatastoreManager()
	dbPool := dbManager.GetPool(ctx, datastore.DefaultPoolName)
	workMan := svc.WorkManager()

	sessionRepo := repository.NewSessionRepository(ctx, dbPool, workMan)
	linkRepo := repository.NewLinkRepository(ctx, dbPool, workMan)

	resources := &ServiceResources{
		SessionRepository: sessionRepo,
		LinkRepository:    linkRepo,
	}

	err = repository.Migrate(ctx, dbManager, "../../migrations/0001")
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

// WithTestDependencies runs testFn for each known dependency option.
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
