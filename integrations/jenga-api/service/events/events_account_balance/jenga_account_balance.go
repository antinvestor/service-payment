package events_account_balance

import (
	"context"
	"errors"
	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/pitabwire/frame"
)

type JengaAccountBalance struct {
	Service     *frame.Service
	Client      coreapi.JengaApiClient

}

func (event *JengaAccountBalance) Name() string {
	return "jenga.account.balance"
}

func (event *JengaAccountBalance) PayloadType() any {
	return &models.AccountBalanceRequest{}
}

func (event *JengaAccountBalance) Validate(ctx context.Context, payload any) error {
	request := payload.(*models.AccountBalanceRequest)

	if request.CountryCode == "" {
		return errors.New("country code is required")
	}

	if request.AccountId == "" {
		return errors.New("account is required")
	}		

	return nil
}


func (event *JengaAccountBalance) Execute(ctx context.Context, payload any) error {

	if event.Client == nil {
		return errors.New("jenga client not initialized")
	}	

	request := payload.(*models.AccountBalanceRequest)

	logger := event.Service.L(ctx).WithField("type", event.Name()).WithField("AccountBalanceRequest", request)
	logger.WithField("request", request).Debug("processing  account balance")



	// Generate bearer token for authorization
	token, err := event.Client.GenerateBearerToken()
	//convert token to string

	if err != nil {
		logger.WithError(err).Error("failed to generate bearer token")
		return err
	}
	//log token
	logger.WithField("token", token.AccessToken).Info("---------------generated token--------------------")
    
	// Get account balance
	balance, err := event.Client.InitiateAccountBalance(request.CountryCode, request.AccountId, token.AccessToken)
	if err != nil {
		logger.WithError(err).Error(err)
		return err
	}

	logger.WithField("balance", balance).Debug("got account balance")


	return nil
}

