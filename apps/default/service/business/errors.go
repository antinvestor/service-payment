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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInitializationFail = status.Error(codes.Internal, "Internal configuration is invalid")

	ErrInvalidPaymentRequest = status.Error(codes.InvalidArgument, "Invalid payment request")

	ErrPaymentDoesNotExist = status.Error(codes.NotFound, "Specified payment does not exist")

	ErrPaymentAlreadyProcessed = status.Error(
		codes.FailedPrecondition,
		"Specified payment has already been processed",
	)

	ErrPaymentAlreadyReleased = status.Error(codes.FailedPrecondition, "Specified payment has already been released")

	ErrPaymentAlreadyRefunded = status.Error(codes.FailedPrecondition, "Specified payment has already been refunded")

	ErrPaymentAlreadyCanceled = status.Error(codes.FailedPrecondition, "Specified payment has already been canceled")

	ErrPaymentAlreadySettled = status.Error(codes.FailedPrecondition, "Specified payment has already been settled")

	ErrPaymentAlreadyPartiallySettled = status.Error(
		codes.FailedPrecondition,
		"Specified payment has already been partially settled",
	)

	ErrPaymentAlreadyPartiallyRefunded = status.Error(
		codes.FailedPrecondition,
		"Specified payment has already been partially refunded",
	)
)
