package repository

import (
	"context"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
)

type AccountRepository interface {
	GetByID(ctx context.Context, id string) (*models.Account, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error)
	Save(ctx context.Context, account *models.Account) error
}

type accountRepository struct {
	abstractRepository
}

func NewAccountRepository(ctx context.Context, service *frame.Service) AccountRepository {
	return &accountRepository{abstractRepository{service: service}}
}

func (repo *accountRepository) GetByID(ctx context.Context, id string) (*models.Account, error) {
	account := models.Account{}
	err := repo.readDb(ctx).First(&account, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (repo *accountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error) {
	account := models.Account{}
	err := repo.readDb(ctx).First(&account, "account_number = ? ", accountNumber).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (repo *accountRepository) Save(ctx context.Context, account *models.Account) error {
	return repo.writeDb(ctx).Save(account).Error
}
