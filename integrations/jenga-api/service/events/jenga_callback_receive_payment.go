// Copyright 2023-2024 Ant Investor Ltd
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

package events

import (
	"context"
	"encoding/json"
	"errors"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/go-redis/redis"
	"github.com/pitabwire/frame"
)

type JengaCallbackReceivePayment struct {
	Service       *frame.Service
	RedisClient   *redis.Client
	PaymentClient *paymentV1.PaymentClient
}

func (event *JengaCallbackReceivePayment) Name() string {
	return "jenga.callback.receive.payment"
}

func (event *JengaCallbackReceivePayment) PayloadType() any {
	return &models.CallbackRequest{}
}

func (event *JengaCallbackReceivePayment) Validate(ctx context.Context, payload any) error {
	request := payload.(*models.CallbackRequest)

	if request.Transaction.Reference == "" {
		return errors.New("transaction reference is required")
	}

	if request.Transaction.ID == "" {
		return errors.New("transaction ID is required")
	}

	return nil
}

func (event *JengaCallbackReceivePayment) Execute(ctx context.Context, payload any) error {

	if event.PaymentClient == nil {
		return errors.New("payment client not initialized")
	}

	callback := payload.(*models.CallbackRequest)

	// Extract relevant information from callback
	payment := &paymentV1.Payment{
		Id:         callback.Transaction.Reference,
		ExternalId: callback.Transaction.ID,
		Amount:     callback.Transaction.Amount,
		Currency:   callback.Transaction.Currency,
	}

	// Add any additional information from callback to extras
	extras := make(map[string]string)
	extras["status"] = callback.Transaction.Status
	extras["statusDescription"] = callback.Transaction.StatusDescription
	extras["transactionDate"] = callback.Transaction.TransactionDate

	// Marshal the full callback to JSON and store it in extras
	callbackJSON, err := json.Marshal(callback)
	if err == nil {
		extras["raw_callback"] = string(callbackJSON)
	}

	payment.Extras = extras

	// Invoke the GRPC receive method
	_, err = event.PaymentClient.Receive(ctx, payment)
	return err
}
