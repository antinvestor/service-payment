package repository

import (
	"context"

	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/workerpool"

	"github.com/antinvestor/service-payments/service/models"
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
