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

package queue

import (
	"context"
	"strconv"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/config"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/credentials"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v2/events"
	"github.com/pitabwire/frame/v2/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type paymentHandler struct {
	statusEmitter
	credsResolver *credentials.Resolver
	ycCli         client.YellowcardClient
	catalog       *client.Catalog
	metrics       *integrationobs.Metrics
}

// NewPaymentHandler creates a queue worker for send (disbursement) payments.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	ycCli client.YellowcardClient,
	catalog *client.Catalog,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.YellowcardConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		statusEmitter: statusEmitter{eventsMan: eventsMan},
		credsResolver: credentials.NewResolver(settingsCli, cfg),
		ycCli:         ycCli,
		catalog:       catalog,
		metrics:       integrationobs.NewMetrics(providerName),
	}
}

func (h *paymentHandler) fail(ctx context.Context, paymentID, externalID, reason string, extras map[string]any) {
	h.metrics.QueueFailed(ctx, "payment", reason)
	extras["entity_type"] = "payment"
	extras[collection.ExtraProvider] = providerName
	h.emitStatus(ctx, paymentID, externalID, commonv1.STATUS_FAILED, extras)
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "yellowcard.payment")
	defer logger.Release()
	logger.Debug("queue handler started")

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, "payment", "unmarshal_error")
		return nil // non-retriable
	}

	paymentID := payment.GetId()
	logger = logger.WithField("payment_id", paymentID)

	creds, err := h.credsResolver.FromHeaders(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.fail(ctx, paymentID, "", "credentials_error", failureExtras(failureCredentials, err.Error()))
		return nil
	}

	extra := payment.GetExtra()
	phone := normalizeMSISDN(payment.GetRecipient().GetContactId())
	// Bank destinations arrive through extras; the Payment proto only carries
	// a contact link for the recipient.
	accountNumber := extraString(extra, "account_number", "recipient_account_number")

	corr, err := resolveCorridor(phone, extra, payment.GetAmount().GetCurrencyCode(), creds)
	if err != nil {
		logger.WithError(err).Error("failed to resolve country")
		h.fail(ctx, paymentID, "", "provider_error", failureExtras(failureInvalidCountry, err.Error()))
		return nil
	}

	localAmount := moneyToLocalAmount(payment.GetAmount())
	if localAmount <= 0 {
		h.fail(ctx, paymentID, "", "provider_error", failureExtras(failureInvalidAmount, "amount must be a positive whole number"))
		return nil
	}

	channelType := resolveChannelType(extra, creds)
	if channelType == "" {
		switch {
		case accountNumber != "" && accountNumber != strings.TrimPrefix(phone, "+") && accountNumber != phone:
			channelType = client.ChannelTypeBank
		case phone != "":
			channelType = client.ChannelTypeMomo
		default:
			channelType = client.ChannelTypeBank
		}
	}

	channels, err := h.catalog.Channels(ctx, creds, corr.Country)
	if err != nil {
		logger.WithError(err).Error("failed to load channels")
		h.fail(ctx, paymentID, "", "provider_error", providerFailureExtras(err))
		return nil
	}
	channel, ok := client.SelectChannel(channels, corr.Currency, channelType, client.RampTypeWithdraw)
	if !ok {
		h.fail(ctx, paymentID, "", "provider_error", failureExtras(failureChannelUnavailable,
			"no active "+channelType+" send channel for "+corr.Country+" "+corr.Currency))
		return nil
	}

	networks, err := h.catalog.Networks(ctx, creds, corr.Country)
	if err != nil {
		logger.WithError(err).Error("failed to load networks")
		h.fail(ctx, paymentID, "", "provider_error", providerFailureExtras(err))
		return nil
	}
	hint := extraString(extra, collection.ExtraNetwork, "bank_code", "bank_name")
	if hint == "" && channelType == client.ChannelTypeMomo {
		hint = creds.Network
	}
	network, found := client.ResolveNetwork(networks, hint, channel, channelType)
	if !found {
		h.fail(ctx, paymentID, "", "provider_error", failureExtras(failureInvalidNetwork,
			"no active "+channelType+" network for "+corr.Country))
		return nil
	}

	dest := client.Destination{
		AccountType: channelType,
		NetworkID:   network.ID,
		AccountName: extraString(extra, "recipient_name", "account_name", collection.ExtraCustomerName, collection.ExtraDisplayName),
		Country:     corr.Country,
	}
	if channelType == client.ChannelTypeMomo {
		if phone == "" {
			h.fail(ctx, paymentID, "", "provider_error", failureExtras(client.TxErrInvalidRecipient, "a phone number is required for mobile money"))
			return nil
		}
		dest.AccountNumber = phone
		dest.PhoneNumber = phone
	} else {
		if accountNumber == "" {
			h.fail(ctx, paymentID, "", "provider_error", failureExtras(client.TxErrInvalidRecipient, "a bank account number is required"))
			return nil
		}
		dest.AccountNumber = accountNumber
		dest.AccountBank = network.Name
	}

	reason := extraString(extra, collection.ExtraReason)
	if reason == "" {
		reason = defaultReason
	}

	req := &client.SendRequest{
		ChannelID:    channel.ID,
		SequenceID:   paymentID,
		LocalAmount:  localAmount,
		Country:      corr.Country,
		Currency:     corr.Currency,
		Reason:       reason,
		Sender:       buildParty(extra, normalizeMSISDN(payment.GetSource().GetContactId()), corr.Country, creds),
		Destination:  dest,
		CustomerType: strings.ToLower(creds.CustomerType),
		CustomerUID:  customerUID(headers, extra, payment.GetSource().GetProfileId()),
		ForceAccept:  true,
	}

	send, err := h.ycCli.SubmitSend(ctx, creds, req)
	if err != nil && client.IsDuplicate(err) {
		logger.Warn("send already submitted, looking up existing record")
		send, err = h.ycCli.GetSendBySequenceID(ctx, creds, paymentID)
	}
	if err != nil {
		logger.WithError(err).Error("send submission failed")
		h.fail(ctx, paymentID, "", "provider_error", providerFailureExtras(err))
		return nil
	}

	extras := sendExtras(send, channelType, network.Name)
	if isTerminalFailure(send.Status) {
		logger.WithField("status", send.Status).WithField("error_code", send.ErrorCode).Error("send rejected")
		for k, v := range failureExtras(send.ErrorCode, "") {
			extras[k] = v
		}
		h.fail(ctx, paymentID, send.ID, "rejected", extras)
		return nil
	}

	logger.WithField("send_id", send.ID).WithField("status", send.Status).Debug("send submitted")
	h.metrics.QueueProcessed(ctx, "payment")
	status := commonv1.STATUS_IN_PROCESS
	if strings.EqualFold(send.Status, client.StatusComplete) {
		status = commonv1.STATUS_SUCCESSFUL
	}
	h.emitStatus(ctx, paymentID, send.ID, status, extras)
	return nil
}

// sendExtras builds the extras for an in-flight send.
func sendExtras(s *client.Send, channelType, networkName string) map[string]any {
	return map[string]any{
		"entity_type":             "payment",
		collection.ExtraProvider:  providerName,
		"send_id":                 s.ID,
		"sequence_id":             s.SequenceID,
		"status":                  s.Status,
		"channel_type":            channelType,
		"channel_id":              s.ChannelID,
		"network_id":              s.Destination.NetworkID,
		"network":                 networkName,
		"country":                 s.Country,
		"currency":                s.Currency,
		"local_amount":            strconv.FormatFloat(s.ConvertedAmount, 'f', -1, 64),
		"rate":                    strconv.FormatFloat(s.Rate, 'f', -1, 64),
		"expires_at":              s.ExpiresAt,
		"provider_transaction_id": s.Reference,
	}
}
