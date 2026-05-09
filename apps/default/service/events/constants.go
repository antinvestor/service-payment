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

package events

const (
	EventNameAccountSave     = "account.save"
	EventNameCostSave        = "cost.save"
	EventNamePaymentSave     = "payment.save"
	EventNameStatusSave      = "status.save"
	EventNamePromptSave      = "prompt.save"
	EventNamePaymentLinkSave = "payment_link.save"
	EventNamePaymentInQueue  = "payment.in.queue"
	EventNamePaymentInRoute  = "payment.in.route"
	EventNamePaymentOutQueue = "payment.out.queue"
	EventNamePaymentOutRoute = "payment.out.route"
)
