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
	"fmt"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"connectrpc.com/connect/v2"
	"github.com/pitabwire/frame/v3/data"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/pkg/collection"
)

// continuationExtras are the only extras a continuation may carry to the
// provider. Everything else (route, amount, payer, tenancy) comes from the
// stored prompt, so a continuation cannot redirect or resize the charge.
var continuationExtras = []string{
	collection.ExtraAction,
	collection.ExtraAuthorizationType,
	collection.ExtraChargeID,
	collection.ExtraPIN,
	collection.ExtraEncryptedPIN,
	collection.ExtraNonce,
	collection.ExtraOTP,
	"avs_city",
	"avs_country",
	"avs_line1",
	"avs_line2",
	"avs_postal_code",
	"avs_state",
}

// transientExtras reach the provider with the prompt but are never stored:
// card data and authorization secrets must not rest in the prompt's row.
var transientExtras = []string{
	collection.ExtraEncryptedCardNumber,
	collection.ExtraEncryptedExpiryMonth,
	collection.ExtraEncryptedExpiryYear,
	collection.ExtraEncryptedCVV,
	collection.ExtraCardNonce,
	collection.ExtraNonce,
	collection.ExtraEncryptedPIN,
	collection.ExtraPIN,
	collection.ExtraOTP,
	"card_number",
	"cvv",
	"expiry_month",
	"expiry_year",
}

// isContinuation reports whether a prompt request asks the provider to
// complete a charge it already started (PIN / OTP / AVS) rather than
// start one.
func isContinuation(req *paymentv1.InitiatePromptRequest) bool {
	action := req.GetExtra().GetFields()[collection.ExtraAction].GetStringValue()
	return strings.EqualFold(strings.TrimSpace(action), collection.ActionAuthorize)
}

func isTerminal(status commonv1.STATUS) bool {
	return status == commonv1.STATUS_SUCCESSFUL || status == commonv1.STATUS_FAILED
}

// promptStatus is where a prompt stands: its latest status record. The
// prompt row keeps the status it was raised with, so it is only the answer
// when no status was ever recorded.
func (pb *paymentBusiness) promptStatus(ctx context.Context, prompt *models.Prompt) *commonv1.StatusResponse {
	status, err := pb.statusRepo.GetByEntity(ctx, prompt.ID, "prompt")
	if err != nil || status == nil {
		return prompt.ToAPIStatus()
	}
	return status.ToAPI()
}

// continuePrompt forwards an authorization step (PIN / OTP / AVS) for a
// prompt already raised to the provider that holds its charge. Nothing is
// stored: the step's secrets travel only on the queue, and the prompt's
// status moves only when the provider answers.
func (pb *paymentBusiness) continuePrompt(
	ctx context.Context,
	prompt *models.Prompt,
	req *paymentv1.InitiatePromptRequest,
) (*commonv1.StatusResponse, error) {
	logger := util.Log(ctx).WithField("prompt_id", prompt.ID)

	current := pb.promptStatus(ctx, prompt)
	if isTerminal(current.GetStatus()) {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errPromptSettled(prompt.ID, current.GetStatus()))
	}
	// Only the prompt's own payer authorizes its charge.
	if strings.TrimSpace(req.GetSource().GetProfileId()) != prompt.SourceID {
		return nil, connect.NewError(connect.CodePermissionDenied, errPromptPayer(prompt.ID))
	}
	if route := strings.TrimSpace(req.GetRoute()); route != "" && !strings.EqualFold(route, prompt.Route) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errPromptRoute(prompt.ID))
	}

	var requested data.JSONMap
	requested = requested.FromProtoStruct(req.GetExtra())

	// The charge being authorized is the one the provider reported for this
	// prompt, never one the caller names: no reported charge, nothing to
	// authorize.
	var reported data.JSONMap
	reported = reported.FromProtoStruct(current.GetExtras())
	knownCharge := strings.TrimSpace(reported.GetString(collection.ExtraChargeID))
	if knownCharge == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errPromptNoCharge(prompt.ID))
	}
	askedCharge := strings.TrimSpace(requested.GetString(collection.ExtraChargeID))
	if askedCharge != "" && askedCharge != knownCharge {
		return nil, connect.NewError(connect.CodeInvalidArgument, errPromptCharge(prompt.ID))
	}

	step := map[string]string{}
	for _, key := range continuationExtras {
		if v := strings.TrimSpace(requested.GetString(key)); v != "" {
			step[key] = v
		}
	}
	step[collection.ExtraAction] = collection.ActionAuthorize
	step[collection.ExtraChargeID] = knownCharge

	apiPrompt := prompt.ToAPI(step)
	dial, err := payerNumber(ctx, pb.profileCli, apiPrompt.GetSource())
	if err != nil {
		return nil, err
	}
	if dial != "" {
		apiPrompt.Source.ContactId = dial
	}

	payload, err := proto.Marshal(apiPrompt)
	if err != nil {
		return nil, err
	}
	topic := pb.resolvePromptTopic(prompt.Route)
	if err = pb.qMan.Publish(ctx, topic, payload); err != nil {
		logger.WithError(err).WithField("prompt_topic", topic).Warn("could not publish prompt continuation")
		return nil, err
	}
	logger.WithField("authorization_type", step[collection.ExtraAuthorizationType]).
		Debug("prompt continuation sent to provider")
	return current, nil
}

// storedExtras drops the transient extras from what is kept on the prompt.
func storedExtras(extras data.JSONMap) data.JSONMap {
	for _, key := range transientExtras {
		delete(extras, key)
	}
	return extras
}

// withTransientExtras puts the request's transient extras back on the
// prompt sent to the provider.
func withTransientExtras(apiPrompt *paymentv1.InitiatePromptRequest, req *structpb.Struct) {
	if apiPrompt.GetExtra() == nil {
		apiPrompt.Extra = &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	for _, key := range transientExtras {
		if v, ok := req.GetFields()[key]; ok {
			apiPrompt.Extra.Fields[key] = v
		}
	}
}

func errPromptSettled(id string, status commonv1.STATUS) string {
	return fmt.Sprintf("prompt %s is already %s; nothing left to authorize", id, status.String())
}

func errPromptRoute(id string) string {
	return fmt.Sprintf("prompt %s continues only on the route it was raised on", id)
}

func errPromptCharge(id string) string {
	return fmt.Sprintf("prompt %s authorizes only the charge its provider reported", id)
}

func errPromptNoCharge(id string) string {
	return fmt.Sprintf("prompt %s has no charge waiting for authorization", id)
}

func errPromptPayer(id string) string {
	return fmt.Sprintf("prompt %s is authorized only by its payer", id)
}
