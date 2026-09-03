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

type promptHandler struct {
	statusEmitter
	credsResolver *credentials.Resolver
	ycCli         client.YellowcardClient
	catalog       *client.Catalog
	cfg           *config.YellowcardConfig
	metrics       *integrationobs.Metrics
}

// NewPromptHandler creates a queue worker for receive (collection) prompts.
// A Yellow Card mobile money receive triggers a USSD approval on the
// customer's phone; a bank receive returns account details the customer
// must transfer into.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	ycCli client.YellowcardClient,
	catalog *client.Catalog,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.YellowcardConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		statusEmitter: statusEmitter{eventsMan: eventsMan},
		credsResolver: credentials.NewResolver(settingsCli, cfg),
		ycCli:         ycCli,
		catalog:       catalog,
		cfg:           cfg,
		metrics:       integrationobs.NewMetrics(providerName),
	}
}

func (h *promptHandler) fail(ctx context.Context, promptID, externalID, reason string, extras map[string]any) {
	h.metrics.QueueFailed(ctx, "prompt", reason)
	extras["entity_type"] = "prompt"
	extras[collection.ExtraProvider] = providerName
	h.emitStatus(ctx, promptID, externalID, commonv1.STATUS_FAILED, extras)
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "yellowcard.prompt")
	defer logger.Release()
	logger.Debug("queue handler started")

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, "prompt", "unmarshal_error")
		return nil // non-retriable
	}

	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	creds, err := h.credsResolver.FromHeaders(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.fail(ctx, promptID, "", "credentials_error", failureExtras(failureCredentials, err.Error()))
		return nil
	}

	phone := normalizeMSISDN(prompt.GetRecipient().GetContactId())
	if phone == "" {
		phone = normalizeMSISDN(prompt.GetSource().GetContactId())
	}
	extra := prompt.GetExtra()

	corr, err := resolveCorridor(phone, extra, prompt.GetAmount().GetCurrencyCode(), creds)
	if err != nil {
		logger.WithError(err).Error("failed to resolve country")
		h.fail(ctx, promptID, "", "provider_error", failureExtras(failureInvalidCountry, err.Error()))
		return nil
	}

	localAmount := moneyToLocalAmount(prompt.GetAmount())
	if localAmount <= 0 {
		h.fail(ctx, promptID, "", "provider_error", failureExtras(failureInvalidAmount, "amount must be a positive whole number"))
		return nil
	}

	channels, err := h.catalog.Channels(ctx, creds, corr.Country)
	if err != nil {
		logger.WithError(err).Error("failed to load channels")
		h.fail(ctx, promptID, "", "provider_error", providerFailureExtras(err))
		return nil
	}

	channelType := resolveChannelType(extra, creds)
	if channelType == "" {
		if client.HasActiveChannel(channels, client.ChannelTypeMomo, client.RampTypeDeposit) {
			channelType = client.ChannelTypeMomo
		} else {
			channelType = client.ChannelTypeBank
		}
	}
	channel, ok := client.SelectChannel(channels, corr.Currency, channelType, client.RampTypeDeposit)
	if !ok {
		logger.WithField("country", corr.Country).WithField("channel_type", channelType).Error("no active receive channel")
		h.fail(ctx, promptID, "", "provider_error", failureExtras(failureChannelUnavailable,
			"no active "+channelType+" receive channel for "+corr.Country+" "+corr.Currency))
		return nil
	}

	source := client.Source{AccountType: channelType}
	networkName := ""
	if channelType == client.ChannelTypeMomo {
		if phone == "" {
			h.fail(ctx, promptID, "", "provider_error", failureExtras(client.TxErrInvalidRecipient, "a phone number is required for mobile money"))
			return nil
		}
		networks, netErr := h.catalog.Networks(ctx, creds, corr.Country)
		if netErr != nil {
			logger.WithError(netErr).Error("failed to load networks")
			h.fail(ctx, promptID, "", "provider_error", providerFailureExtras(netErr))
			return nil
		}
		hint := extraString(extra, collection.ExtraNetwork)
		if hint == "" {
			hint = creds.Network
		}
		network, found := client.ResolveNetwork(networks, hint, channel, client.ChannelTypeMomo)
		if !found {
			h.fail(ctx, promptID, "", "provider_error", failureExtras(failureInvalidNetwork,
				"no active mobile money network for "+corr.Country))
			return nil
		}
		source.AccountNumber = phone
		source.NetworkID = network.ID
		networkName = network.Name
	}

	redirectURL := extraString(extra, collection.ExtraRedirectURL, collection.ExtraSuccessURL)
	if redirectURL == "" && h.cfg != nil {
		redirectURL = h.cfg.DefaultRedirectURL
	}
	reason := extraString(extra, collection.ExtraReason)
	if reason == "" {
		reason = defaultReason
	}

	req := &client.ReceiveRequest{
		ChannelID:    channel.ID,
		SequenceID:   promptID,
		LocalAmount:  localAmount,
		Country:      corr.Country,
		Currency:     corr.Currency,
		Source:       source,
		Recipient:    buildParty(extra, phone, corr.Country, creds),
		CustomerType: strings.ToLower(creds.CustomerType),
		CustomerUID:  customerUID(headers, extra, prompt.GetRecipient().GetProfileId()),
		RedirectURL:  redirectURL,
		Reason:       reason,
		ForceAccept:  true,
	}

	receive, err := h.ycCli.SubmitReceive(ctx, creds, req)
	if err != nil && client.IsDuplicate(err) {
		logger.Warn("receive already submitted, looking up existing record")
		receive, err = h.ycCli.GetReceiveBySequenceID(ctx, creds, promptID)
	}
	if err != nil {
		logger.WithError(err).Error("receive submission failed")
		h.fail(ctx, promptID, "", "provider_error", providerFailureExtras(err))
		return nil
	}

	extras := receiveExtras(receive, channelType, networkName)
	if isTerminalFailure(receive.Status) {
		logger.WithField("status", receive.Status).WithField("error_code", receive.ErrorCode).Error("receive rejected")
		for k, v := range failureExtras(receive.ErrorCode, "") {
			extras[k] = v
		}
		h.fail(ctx, promptID, receive.ID, "rejected", extras)
		return nil
	}

	logger.WithField("receive_id", receive.ID).WithField("status", receive.Status).Debug("receive submitted")
	h.metrics.QueueProcessed(ctx, "prompt")
	status := commonv1.STATUS_IN_PROCESS
	if strings.EqualFold(receive.Status, client.StatusComplete) {
		status = commonv1.STATUS_SUCCESSFUL
	}
	h.emitStatus(ctx, promptID, receive.ID, status, extras)
	return nil
}
