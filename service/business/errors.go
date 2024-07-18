package business

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrorInitializationFail = status.Error(codes.Internal, "Internal configuration is invalid")

	ErrorInvalidPaymentRequest = status.Error(codes.InvalidArgument, "Invalid payment request")

	ErrorPaymentDoesNotExist = status.Error(codes.NotFound, "Specified payment does not exist")

	ErrorPaymentAlreadyProcessed = status.Error(codes.FailedPrecondition, "Specified payment has already been processed")

	ErrorPaymentAlreadyReleased = status.Error(codes.FailedPrecondition, "Specified payment has already been released")

	ErrorPaymentAlreadyRefunded = status.Error(codes.FailedPrecondition, "Specified payment has already been refunded")

	ErrorPaymentAlreadyCanceled = status.Error(codes.FailedPrecondition, "Specified payment has already been canceled")

	ErrorPaymentAlreadySettled = status.Error(codes.FailedPrecondition, "Specified payment has already been settled")

	ErrorPaymentAlreadyPartiallySettled = status.Error(codes.FailedPrecondition, "Specified payment has already been partially settled")

	ErrorPaymentAlreadyPartiallyRefunded = status.Error(codes.FailedPrecondition, "Specified payment has already been partially refunded")
	
)
