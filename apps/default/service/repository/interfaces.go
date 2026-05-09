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

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/workerpool"
)

type PaymentRepository interface {
	datastore.BaseRepository[*models.Payment]
	Search(ctx context.Context, query *data.SearchQuery) (workerpool.JobResultPipe[[]*models.Payment], error)
	GetByProfileID(ctx context.Context, profileID string) ([]*models.Payment, error)
}

type AccountRepository interface {
	datastore.BaseRepository[*models.Account]
	GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error)
}

type CostRepository interface {
	datastore.BaseRepository[*models.Cost]
	GetByPaymentID(ctx context.Context, paymentID string) ([]*models.Cost, error)
}

type StatusRepository interface {
	datastore.BaseRepository[*models.Status]
	GetByEntity(ctx context.Context, entityID string, entityType string) (*models.Status, error)
	GetByEntityIDList(ctx context.Context, entityIDs []string, entityType string) (map[string]*models.Status, error)
}

type PromptRepository interface {
	datastore.BaseRepository[*models.Prompt]
	GetByProfileID(ctx context.Context, profileID string) ([]*models.Prompt, error)
}

type PaymentLinkRepository interface {
	datastore.BaseRepository[*models.PaymentLink]
	GetByProfileID(ctx context.Context, profileID string) ([]*models.PaymentLink, error)
}

type RouteRepository interface {
	datastore.BaseRepository[*models.Route]
	GetByMode(ctx context.Context, mode string) ([]*models.Route, error)
	GetByModeTypeAndPartitionID(
		ctx context.Context,
		mode string,
		routeType string,
		partitionID string,
	) ([]*models.Route, error)
}
