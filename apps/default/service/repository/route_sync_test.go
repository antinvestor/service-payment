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

package repository_test

import (
	"context"
	"testing"

	"github.com/pitabwire/frame/v3"
	"github.com/pitabwire/frame/v3/datastore"
	"github.com/pitabwire/frame/v3/datastore/pool"
	"github.com/pitabwire/frame/v3/frametests"
	"github.com/pitabwire/frame/v3/frametests/definition"
	"github.com/pitabwire/frame/v3/frametests/deps/testpostgres"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/antinvestor/service-payments/apps/default/service/repository"
)

const routesJSON = `[
 {"id":"rx-ops","tenant_id":"t1","partition_id":"p1","name":"Ops","route_type":"any","mode":"rx",
  "uri":"gcppubsub://proj/ops-received"},
 {"id":"tx-mpesa","tenant_id":"t1","partition_id":"p1","name":"M-Pesa","route_type":"msisdn","mode":"tx",
  "uri":"gcppubsub://proj/mpesa-payments"}
]`

func TestParseRouteConfig(t *testing.T) {
	routes, err := repository.ParseRouteConfig(routesJSON)
	require.NoError(t, err)
	require.Len(t, routes, 2)

	none, err := repository.ParseRouteConfig("  ")
	require.NoError(t, err)
	require.Empty(t, none)

	for name, raw := range map[string]string{
		"not json":     `{`,
		"no id":        `[{"tenant_id":"t","partition_id":"p","route_type":"any","mode":"rx","uri":"x"}]`,
		"no partition": `[{"id":"a","tenant_id":"t","route_type":"any","mode":"rx","uri":"x"}]`,
		"bad mode":     `[{"id":"a","tenant_id":"t","partition_id":"p","route_type":"any","mode":"in","uri":"x"}]`,
		"long type":    `[{"id":"a","tenant_id":"t","partition_id":"p","route_type":"mobile_money","mode":"rx","uri":"x"}]`,
		"no uri":       `[{"id":"a","tenant_id":"t","partition_id":"p","route_type":"any","mode":"rx"}]`,
		"duplicate id": `[{"id":"a","tenant_id":"t","partition_id":"p","route_type":"any","mode":"rx","uri":"x"},{"id":"a","tenant_id":"t","partition_id":"p","route_type":"any","mode":"tx","uri":"y"}]`,
	} {
		_, err = repository.ParseRouteConfig(raw)
		require.ErrorIs(t, err, repository.ErrInvalidRouteConfig, name)
	}
}

type routeSyncSuite struct {
	frametests.FrameBaseTestSuite
}

func TestRouteSyncSuite(t *testing.T) {
	suite.Run(t, new(routeSyncSuite))
}

func (s *routeSyncSuite) SetupSuite() {
	s.InitResourceFunc = func(_ context.Context) []definition.TestResource {
		return []definition.TestResource{
			testpostgres.NewWithOpts("payment_routes",
				definition.WithUserName("ant"), definition.WithCredential("s3cr3t"),
				definition.WithEnableLogging(false)),
		}
	}
	s.FrameBaseTestSuite.SetupSuite()
}

func (s *routeSyncSuite) pool() (context.Context, pool.Pool) {
	ctx := s.T().Context()
	var db definition.DependancyConn
	for _, r := range s.Resources() {
		if r.Name() == testpostgres.PostgresqlDBImage && r.GetDS(ctx).IsDB() {
			db = r
		}
	}
	s.Require().NotNil(db)
	dsn, cleanup, err := db.GetRandomisedDS(ctx, util.RandomAlphaNumericString(8))
	s.Require().NoError(err)
	s.T().Cleanup(func() { cleanup(ctx) })

	ctx, svc := frame.NewServiceWithContext(ctx, frame.WithName("payment-routes-test"),
		frame.WithDatastore(pool.WithConnection(dsn.String(), false)))
	s.T().Cleanup(func() { svc.Stop(ctx) })
	svc.Init(ctx)
	s.Require().NoError(repository.Migrate(ctx, svc.DatastoreManager(), "../../migrations/0001"))
	dbPool := svc.DatastoreManager().GetPool(ctx, datastore.DefaultPoolName)
	s.Require().NotNil(dbPool)
	return ctx, dbPool
}

// Declared routes are written, a re-run changes nothing, and a changed
// declaration updates the route in place.
func (s *routeSyncSuite) TestSyncRoutesWritesAndUpdatesDeclaredRoutes() {
	ctx, dbPool := s.pool()
	routes, err := repository.ParseRouteConfig(routesJSON)
	s.Require().NoError(err)

	s.Require().NoError(repository.SyncRoutes(ctx, dbPool, routes))
	s.Require().NoError(repository.SyncRoutes(ctx, dbPool, routes))

	type row struct {
		ID, PartitionID, RouteType, Mode, URI string
		Version                               int64
	}
	var rows []row
	s.Require().NoError(dbPool.DB(ctx, true).Table("routes").Order("id").Find(&rows).Error)
	s.Require().Len(rows, 2)
	s.Equal(row{ID: "rx-ops", PartitionID: "p1", RouteType: "any", Mode: "rx",
		URI: "gcppubsub://proj/ops-received", Version: 1}, rows[0])
	s.Equal(int64(1), rows[1].Version, "an unchanged re-run writes nothing")

	routes[1].URI = "gcppubsub://proj/mpesa-payouts"
	s.Require().NoError(repository.SyncRoutes(ctx, dbPool, routes))
	var updated row
	s.Require().NoError(dbPool.DB(ctx, true).Table("routes").Where("id = ?", "tx-mpesa").Take(&updated).Error)
	s.Equal("gcppubsub://proj/mpesa-payouts", updated.URI)
	s.Equal(int64(2), updated.Version)
}
