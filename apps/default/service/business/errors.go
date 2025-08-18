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
