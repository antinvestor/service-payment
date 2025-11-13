package repository

import (
	"context"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

type accountRepository struct {
	datastore.BaseRepository[*models.Account]
}

func NewAccountRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) AccountRepository {
	repo := accountRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Account](
			ctx, dbPool, workMan, func() *models.Account { return &models.Account{} },
		),
	}
	return &repo
}

func (ar *accountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error) {
	account := &models.Account{}
	err := ar.Pool().DB(ctx, true).First(account, "account_number = ?", accountNumber).Error
	return account, err
}
