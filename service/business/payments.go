package business

import (
	"context"
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	partitionV1 "github.com/antinvestor/apis/go/partition/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"

	"github.com/antinvestor/service-payments-v1/service/events"
	"github.com/antinvestor/service-payments-v1/service/models"
	"github.com/antinvestor/service-payments-v1/service/repository"

	"github.com/pitabwire/frame"
)

type PaymentBusiness interface {
	Send(ctx context.Context, payment *paymentV1.Payment) (*commonv1.StatusResponse, error)
	Receive(ctx context.Context, payment *paymentV1.Payment) (*commonv1.StatusResponse, error)
}

func NewPaymentBusiness(_ context.Context, service *frame.Service, profileCli *profileV1.ProfileClient, partitionCli *partitionV1.PartitionClient) (PaymentBusiness, error) {
	//initialize the service
	if service == nil || profileCli == nil || partitionCli == nil {
		return nil, ErrorInitializationFail
	}
	return &paymentBusiness{
		service:      service,
		profileCli:   profileCli,
		partitionCli: partitionCli,
	}, nil
}

type paymentBusiness struct {
	service      *frame.Service
	profileCli   *profileV1.ProfileClient
	partitionCli *partitionV1.PartitionClient
}

func (pb *paymentBusiness) Send(ctx context.Context, payment *paymentV1.Payment) (*commonv1.StatusResponse, error) {

	logger := pb.service.L().WithField("request", payment)

	authClaim := frame.ClaimsFromContext(ctx)

	logger.WithField("auth claim", authClaim).Info("handling send request")

	payment.SenderProfileType = authClaim.ProfileType
	payment.SenderProfileID = authClaim.ProfileID
	payment.SenderContactID = authClaim.ContactID

	payment.PartitionID = authClaim.PartitionID
	payment.TenantID = authClaim.TenantID
	payment.AccessID = authClaim.AccessID

	payment.DateCreated = time.Now()
	