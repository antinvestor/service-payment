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

package business

import (
	"context"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/workerpool"
)

// CatalogBusiness defines the business interface for catalog operations.
type CatalogBusiness interface {
	CreateCatalogVersion(ctx context.Context, catalogVersion *models.CatalogVersion) (*models.CatalogVersion, error)
	GetCatalogVersion(ctx context.Context, id string) (*models.CatalogVersion, error)
	GetEffectiveCatalog(ctx context.Context, catalogID string, at time.Time) (*models.CatalogVersion, error)
	PublishCatalogVersion(ctx context.Context, id string, effectiveAt time.Time) (*models.CatalogVersion, error)
	CreatePlan(ctx context.Context, plan *models.Plan) (*models.Plan, error)
	CreateComponent(ctx context.Context, component *models.Component) (*models.Component, error)
	CreateTier(ctx context.Context, tier *models.Tier) (*models.Tier, error)
}

type catalogBusiness struct {
	workMan     workerpool.Manager
	catalogRepo repository.CatalogVersionRepository
	planRepo    repository.PlanRepository
	compRepo    repository.ComponentRepository
	tierRepo    repository.TierRepository
}

func NewCatalogBusiness(
	workMan workerpool.Manager,
	catalogRepo repository.CatalogVersionRepository,
	planRepo repository.PlanRepository,
	compRepo repository.ComponentRepository,
	tierRepo repository.TierRepository,
) CatalogBusiness {
	return &catalogBusiness{
		workMan:     workMan,
		catalogRepo: catalogRepo,
		planRepo:    planRepo,
		compRepo:    compRepo,
		tierRepo:    tierRepo,
	}
}

func (b *catalogBusiness) CreateCatalogVersion(
	ctx context.Context,
	cv *models.CatalogVersion,
) (*models.CatalogVersion, error) {
	if cv.CatalogID == "" {
		return nil, ErrCatalogIDRequired
	}
	if cv.Name == "" {
		return nil, ErrCatalogNameRequired
	}
	if cv.Currency == "" {
		return nil, ErrCatalogCurrencyRequired
	}

	cv.GenID(ctx)

	if err := b.catalogRepo.Create(ctx, cv); err != nil {
		return nil, err
	}

	return cv, nil
}

func (b *catalogBusiness) GetCatalogVersion(ctx context.Context, id string) (*models.CatalogVersion, error) {
	if id == "" {
		return nil, ErrCatalogIDRequired
	}

	return b.catalogRepo.GetByID(ctx, id)
}

func (b *catalogBusiness) GetEffectiveCatalog(
	ctx context.Context,
	catalogID string,
	at time.Time,
) (*models.CatalogVersion, error) {
	if catalogID == "" {
		return nil, ErrCatalogIDRequired
	}

	return b.catalogRepo.GetEffectiveByCatalogID(ctx, catalogID, at)
}

func (b *catalogBusiness) PublishCatalogVersion(
	ctx context.Context,
	id string,
	effectiveAt time.Time,
) (*models.CatalogVersion, error) {
	cv, err := b.catalogRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if cv.RetiredAt != nil {
		return nil, apperrors.ErrCatalogVersionRetired
	}

	now := time.Now()
	cv.PublishedAt = &now
	cv.EffectiveAt = &effectiveAt

	_, err = b.catalogRepo.Update(ctx, cv)
	if err != nil {
		return nil, err
	}

	return cv, nil
}

func (b *catalogBusiness) CreatePlan(ctx context.Context, plan *models.Plan) (*models.Plan, error) {
	if plan.CatalogVersionID == "" {
		return nil, ErrCatalogIDRequired
	}
	if plan.Name == "" {
		return nil, ErrPlanNameRequired
	}

	plan.GenID(ctx)

	if err := b.planRepo.Create(ctx, plan); err != nil {
		return nil, err
	}

	return plan, nil
}

func (b *catalogBusiness) CreateComponent(ctx context.Context, comp *models.Component) (*models.Component, error) {
	if comp.PlanID == "" {
		return nil, ErrPlanIDRequired
	}
	if comp.Name == "" {
		return nil, ErrComponentNameRequired
	}
	if comp.MetricKey == "" {
		return nil, ErrMetricKeyRequired
	}

	// Validate pricing model
	switch comp.PricingModel {
	case models.PricingModelFlat, models.PricingModelPerUnit,
		models.PricingModelTiered, models.PricingModelVolume,
		models.PricingModelStairstep:
		// valid
	default:
		return nil, apperrors.ErrInvalidPricingModel
	}

	comp.GenID(ctx)

	if err := b.compRepo.Create(ctx, comp); err != nil {
		return nil, err
	}

	return comp, nil
}

func (b *catalogBusiness) CreateTier(ctx context.Context, tier *models.Tier) (*models.Tier, error) {
	if tier.ComponentID == "" {
		return nil, ErrComponentIDRequired
	}

	tier.GenID(ctx)

	if err := b.tierRepo.Create(ctx, tier); err != nil {
		return nil, err
	}

	return tier, nil
}
