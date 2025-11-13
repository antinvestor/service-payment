package business

import (
	"context"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/payment/v1"
	"github.com/pitabwire/frame/workerpool"

	"github.com/antinvestor/service-payments/service/models"
)

type PaymentBusiness interface {
	Send(ctx context.Context, payment *paymentv1.Payment) (*commonv1.StatusResponse, error)
	Receive(ctx context.Context, payment *paymentv1.Payment) (*commonv1.StatusResponse, error)
	Status(ctx context.Context, status *commonv1.StatusRequest) (*commonv1.StatusResponse, error)
	StatusUpdate(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error)
	Release(ctx context.Context, status *paymentv1.ReleaseRequest) (*commonv1.StatusResponse, error)
	Search(ctx context.Context, query *commonv1.SearchRequest) (workerpool.JobResultPipe[[]*paymentv1.Payment], error)
	InitiatePrompt(ctx context.Context, req *paymentv1.InitiatePromptRequest) (*commonv1.StatusResponse, error)
	CreatePaymentLink(ctx context.Context, req *paymentv1.CreatePaymentLinkRequest) (*commonv1.StatusResponse, error)
	ToAPI(ctx context.Context, payment *models.Payment) (*paymentv1.Payment, error)
	Reconcile(ctx context.Context, msg *paymentv1.ReconcileRequest) (*paymentv1.ReconcileResponse, error)
}
