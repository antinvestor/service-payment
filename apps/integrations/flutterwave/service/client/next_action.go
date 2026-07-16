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

package client

import "strings"

// NextActionType classifies charge.next_action for checkout UX.
type NextActionType string

const (
	NextActionNone                 NextActionType = ""
	NextActionRedirectURL          NextActionType = "redirect_url" // 3DS
	NextActionRequiresPIN          NextActionType = "requires_pin"
	NextActionRequiresOTP          NextActionType = "requires_otp"
	NextActionRequiresAVS          NextActionType = "requires_additional_fields"
	NextActionPaymentInstruction   NextActionType = "payment_instruction"
	NextActionRequiresBankTransfer NextActionType = "requires_bank_transfer"
)

// NextAction is a portable view of Flutterwave charge next steps.
// Checkout uses this without knowing Flutterwave field shapes.
type NextAction struct {
	Type        NextActionType
	RedirectURL string
	// Fields lists AVS field paths when Type is requires_additional_fields.
	Fields []string
	// Note is bank-transfer / payment instruction text.
	Note string
}

// ExtractNextAction normalises charge.NextAction for provider-agnostic consumers.
func ExtractNextAction(ch *Charge) NextAction {
	if ch == nil || ch.NextAction == nil {
		// Fallback: some responses put redirect only on top-level.
		if ch != nil && strings.TrimSpace(ch.RedirectURL) != "" {
			return NextAction{Type: NextActionRedirectURL, RedirectURL: strings.TrimSpace(ch.RedirectURL)}
		}
		return NextAction{}
	}
	t, _ := ch.NextAction["type"].(string)
	t = strings.ToLower(strings.TrimSpace(t))
	na := NextAction{Type: NextActionType(t)}
	switch na.Type {
	case NextActionRedirectURL:
		if ru, ok := ch.NextAction["redirect_url"].(map[string]any); ok {
			if u, ok := ru["url"].(string); ok {
				na.RedirectURL = u
			}
		}
	case NextActionRequiresAVS:
		if avs, ok := ch.NextAction["requires_additional_fields"].(map[string]any); ok {
			if fields, ok := avs["fields"].([]any); ok {
				for _, f := range fields {
					if s, ok := f.(string); ok {
						na.Fields = append(na.Fields, s)
					}
				}
			}
		}
	case NextActionPaymentInstruction:
		if pi, ok := ch.NextAction["payment_instruction"].(map[string]any); ok {
			if n, ok := pi["note"].(string); ok {
				na.Note = n
			}
		}
	case NextActionRequiresBankTransfer:
		// Leave details in raw charge; note optional.
	case NextActionRequiresPIN, NextActionRequiresOTP:
		// Empty object payloads — UI only needs the type.
	}
	if na.RedirectURL == "" {
		na.RedirectURL = ExtractRedirectURL(ch)
	}
	if na.RedirectURL == "" && strings.TrimSpace(ch.RedirectURL) != "" {
		na.RedirectURL = strings.TrimSpace(ch.RedirectURL)
		if na.Type == "" {
			na.Type = NextActionRedirectURL
		}
	}
	return na
}

// PaymentMethodIDFromCharge extracts pmd_* from charge response when present.
func PaymentMethodIDFromCharge(ch *Charge) string {
	if ch == nil || ch.PaymentMethod == nil {
		return ""
	}
	if id, ok := ch.PaymentMethod["id"].(string); ok {
		return id
	}
	return ""
}

// CustomerIDFromCharge returns the customer id on a charge.
func CustomerIDFromCharge(ch *Charge) string {
	if ch == nil {
		return ""
	}
	return ch.CustomerID
}
