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

//nolint:dupl // Similar pattern for different models is acceptable
package events

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/util"
)

type PromptSave struct {
	promptRepo repository.PromptRepository
}

// NewPromptSave creates a new PromptSave event handler with the required dependencies.
func NewPromptSave(promptRepo repository.PromptRepository) *PromptSave {
	return &PromptSave{
		promptRepo: promptRepo,
	}
}

func (e *PromptSave) Name() string {
	return EventNamePromptSave
}

func (e *PromptSave) PayloadType() any {
	return &models.Prompt{}
}

func (e *PromptSave) Validate(_ context.Context, payload any) error {
	prompt, ok := payload.(*models.Prompt)
	if !ok {
		return errors.New("payload is not of type models.Prompt")
	}

	if prompt.GetID() == "" {
		if prompt.ID != "" {
			return nil
		}
		return errors.New("prompt Id should already have been set")
	}

	return nil
}

func (e *PromptSave) Execute(ctx context.Context, payload any) error {
	prompt, ok := payload.(*models.Prompt)
	if !ok {
		return errors.New("payload is not of type models.Prompt")
	}

	logger := util.Log(ctx).WithFields(map[string]any{"prompt_id": prompt.ID, "type": e.Name()})
	logger.Debug("handling event")

	// Attempt to save to database
	err := e.promptRepo.Create(ctx, prompt)
	if err != nil {
		if data.ErrorIsDuplicateKey(err) {
			logger.Debug("record already exists, skipping duplicate")
			return nil
		}
		logger.WithError(err).Warn("could not save prompt to db")
		return err
	}

	logger.Debug("successfully saved record to db")
	return nil
}
