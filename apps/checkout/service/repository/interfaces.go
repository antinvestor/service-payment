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

	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/pitabwire/frame/datastore"
)

// SessionRepository manages CheckoutSession persistence.
type SessionRepository interface {
	datastore.BaseRepository[*models.CheckoutSession]
	GetByRef(ctx context.Context, ref string) (*models.CheckoutSession, error)
	ListByStatus(ctx context.Context, status string, limit int) ([]*models.CheckoutSession, error)
}

// LinkRepository manages CheckoutLink persistence.
type LinkRepository interface {
	datastore.BaseRepository[*models.CheckoutLink]
	GetByRef(ctx context.Context, ref string) (*models.CheckoutLink, error)
}
