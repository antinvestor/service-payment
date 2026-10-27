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

package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/pitabwire/frame/v3/datastore/pool"

	"github.com/antinvestor/service-payments/apps/default/service/models"
)

// RouteConfig is one payment route as deployment config declares it. Routes
// are config, not data: the setup Job writes them, so an environment's routes
// live with the rest of its configuration instead of being inserted by hand.
type RouteConfig struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	PartitionID string `json:"partition_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	RouteType   string `json:"route_type"`
	Mode        string `json:"mode"`
	URI         string `json:"uri"`
}

// routeColumnLimit is the width of the routes table's short text columns.
const routeColumnLimit = 10

// ErrInvalidRouteConfig is returned for a route config that cannot be written.
var ErrInvalidRouteConfig = errors.New("invalid payment route config")

// ParseRouteConfig reads the PAYMENT_ROUTES JSON list. An empty value declares
// no routes.
func ParseRouteConfig(raw string) ([]RouteConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var routes []RouteConfig
	if err := json.Unmarshal([]byte(raw), &routes); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRouteConfig, err)
	}
	seen := map[string]bool{}
	for i, r := range routes {
		if err := r.validate(); err != nil {
			return nil, fmt.Errorf("%w: route %d: %w", ErrInvalidRouteConfig, i, err)
		}
		if seen[r.ID] {
			return nil, fmt.Errorf("%w: route id %q is declared twice", ErrInvalidRouteConfig, r.ID)
		}
		seen[r.ID] = true
	}
	return routes, nil
}

func (r RouteConfig) validate() error {
	switch {
	case strings.TrimSpace(r.ID) == "":
		return errors.New("id is required")
	case strings.TrimSpace(r.TenantID) == "" || strings.TrimSpace(r.PartitionID) == "":
		return errors.New("tenant_id and partition_id are required")
	case strings.TrimSpace(r.URI) == "":
		return errors.New("uri is required")
	case r.RouteType == "" || len(r.RouteType) > routeColumnLimit:
		return fmt.Errorf("route_type %q must be 1 to %d characters", r.RouteType, routeColumnLimit)
	}
	switch r.Mode {
	case models.RouteModeReceive, models.RouteModeTransmit, models.RouteModeTransceive:
		return nil
	default:
		return fmt.Errorf("mode %q must be %s, %s or %s", r.Mode,
			models.RouteModeReceive, models.RouteModeTransmit, models.RouteModeTransceive)
	}
}

// SyncRoutes writes the declared routes, creating or updating each by id.
// Routes not declared are left as they are.
func SyncRoutes(ctx context.Context, dbPool pool.Pool, routes []RouteConfig) error {
	db := dbPool.DB(ctx, false)
	for _, r := range routes {
		if err := db.Exec(`
INSERT INTO routes (id, tenant_id, partition_id, name, description, route_type, mode, uri, version)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)
ON CONFLICT (id) DO UPDATE SET
    tenant_id = EXCLUDED.tenant_id, partition_id = EXCLUDED.partition_id,
    name = EXCLUDED.name, description = EXCLUDED.description,
    route_type = EXCLUDED.route_type, mode = EXCLUDED.mode, uri = EXCLUDED.uri,
    deleted_at = NULL, modified_at = CURRENT_TIMESTAMP, version = routes.version + 1
WHERE (routes.tenant_id, routes.partition_id, routes.name, routes.description, routes.route_type,
       routes.mode, routes.uri, routes.deleted_at IS NULL)
   IS DISTINCT FROM (EXCLUDED.tenant_id, EXCLUDED.partition_id, EXCLUDED.name, EXCLUDED.description,
       EXCLUDED.route_type, EXCLUDED.mode, EXCLUDED.uri, true)`,
			r.ID, r.TenantID, r.PartitionID, r.Name, r.Description, r.RouteType, r.Mode, r.URI,
		).Error; err != nil {
			return fmt.Errorf("write payment route %s: %w", r.ID, err)
		}
	}
	return nil
}
