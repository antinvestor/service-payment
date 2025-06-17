package business_test

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/antinvestor/apis/go/common"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	partitionV1 "github.com/antinvestor/apis/go/partition/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"

	money "google.golang.org/genproto/googleapis/type/money"

	"testing"
	"time"

	"github.com/antinvestor/service-payments/config"
	business "github.com/antinvestor/service-payments/service/business"
	"github.com/antinvestor/service-payments/service/events"
	"github.com/pitabwire/frame"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/mock/gomock"
)

func getService(serviceName string) (*ctxSrv, error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "ant",
			"POSTGRES_PASSWORD": "secret",
			"POSTGRES_DB":       "service_payment",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(5 * time.Minute),
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}
	// Ensure the container is shut down at the end
	defer func() {
		fmt.Println("Shutting down Postgres container...")
		if terminateErr := postgresC.Terminate(ctx); terminateErr != nil {
			fmt.Printf("failed to terminate container: %s\n", err.Error())
		}
	}()

	mappedPort, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	hostIP, err := postgresC.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	dbURL := fmt.Sprintf("postgres://ant:secret@%s:%s/service_payment?sslmode=disable", hostIP, mappedPort.Port())
	testDb := frame.WithDatastoreConnection(dbURL, false)

	var pcfg config.PaymentConfig
	//_ = frame.ConfigFillFromEnv(&pcfg)

	ctx, service := frame.NewService(serviceName, testDb, frame.WithConfig(&pcfg), frame.WithNoopDriver())
	log.Printf("New Service = %v", ctx)

	m := make(map[string]string)
	m["sub"] = "testing"
	m["tenant_id"] = "test_tenant-id"
	m["partition_id"] = "test_partition-id"
	m["access_id"] = "test_access-id"

	claims := frame.ClaimsFromMap(m)
	ctx = claims.ClaimsToContext(ctx)

	eventList := frame.WithRegisterEvents(
		&events.PaymentSave{Service: service},
		&events.PaymentStatusSave{Service: service})
	service.Init(ctx, eventList)
	_ = service.Run(ctx, "")
	return &ctxSrv{
		ctx,
		service,
	}, nil
}

type ctxSrv struct {
	ctx context.Context
	srv *frame.Service
}

func getProfileCli(t *testing.T) *profileV1.ProfileClient {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProfileService := profileV1.NewMockProfileServiceClient(ctrl)
	mockProfileService.EXPECT().
		GetById(gomock.Any(), gomock.Any()).
		Return(&profileV1.GetByIdResponse{
			Data: &profileV1.ProfileObject{
				Id: "test_profile-id",
			},
		}, nil).AnyTimes()
	mockProfileService.EXPECT().
		GetByContact(gomock.Any(), gomock.Any()).
		Return(&profileV1.GetByContactResponse{
			Data: &profileV1.ProfileObject{
				Id: "test_profile-id",
			},
		}, nil).AnyTimes()

	profileCli := profileV1.Init(&common.GrpcClientBase{}, mockProfileService)
	return profileCli
}

func getPartitionCli(t *testing.T) *partitionV1.PartitionClient {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPartitionService := partitionV1.NewMockPartitionServiceClient(ctrl)

	mockPartitionService.EXPECT().
		GetAccess(gomock.Any(), gomock.Any()).
		Return(&partitionV1.GetAccessResponse{Data: &partitionV1.AccessObject{
			AccessId: "test_access-id",
			Partition: &partitionV1.PartitionObject{
				Id:       "test_partition-id",
				TenantId: "test_tenant-id",
			},
		}}, nil).AnyTimes()

	profileCli := partitionV1.Init(&common.GrpcClientBase{}, mockPartitionService)
	return profileCli
}

func TestNewPaymentBusiness_Success(t *testing.T) {
	profileCli := getProfileCli(t)
	partitionCli := getPartitionCli(t)

	type args struct {
		ctxService   *ctxSrv
		profileCli   *profileV1.ProfileClient
		partitionCli *partitionV1.PartitionClient
	}
	tests := []struct {
		name      string
		args      args
		want      business.PaymentBusiness
		expectErr bool
	}{
		{
			name: "NewPaymentBusiness",
			args: args{
				ctxService:   nil,
				profileCli:   profileCli,
				partitionCli: partitionCli},
			expectErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := getService(tt.name)
			log.Printf("ctxService = %v", service.ctx)
			if err != nil {
				t.Errorf("failed to get service: %v", err)
			}

			pb, err := business.NewPaymentBusiness(service.ctx, service.srv, tt.args.profileCli, tt.args.partitionCli)

			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if pb == nil {
				t.Errorf("expected payment business, got nil")
			}
		})
	}
}

func TestNewPaymentBusinessWithNils(t *testing.T) {
	profileCli := getProfileCli(t)
	partitionCli := getPartitionCli(t)

	type args struct {
		ctxService   *ctxSrv
		profileCli   *profileV1.ProfileClient
		partitionCli *partitionV1.PartitionClient
	}
	tests := []struct {
		name      string
		args      args
		want      business.PaymentBusiness
		expectErr bool
	}{
		{
			name: "NewPaymentBusinessWithNils",
			args: args{
				ctxService:   nil,
				profileCli:   nil,
				partitionCli: nil},
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := getService(tt.name)
			if err != nil {
				t.Errorf("failed to get service: %v", err)
			}
			pb, err := business.NewPaymentBusiness(service.ctx, nil, profileCli, partitionCli)

			if !errors.Is(err, business.ErrorInitializationFail) {
				t.Errorf("expected ErrorInitializationFail, got %v", err)
			}

			if pb != nil {
				t.Errorf("expected nil PaymentBusiness instance, got %v", pb)
			}
		})
	}
}

func TestSendPaymentWithValidData(t *testing.T) {
	profileCli := getProfileCli(t)
	partitionCli := getPartitionCli(t)

	type fields struct {
		ctxService   *ctxSrv
		profileCli   *profileV1.ProfileClient
		partitionCli *partitionV1.PartitionClient
	}

	type args struct {
		ctx     context.Context
		message *paymentV1.Payment
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *commonv1.StatusResponse
		wantErr bool
	}{
		{
			name: "Send",
			fields: fields{
				ctxService:   nil,
				profileCli:   profileCli,
				partitionCli: partitionCli,
			},
			args: args{
				ctx: nil,
				message: &paymentV1.Payment{
					Id: "c2f4j7au6s7f91uqnojg",
					Recipient: &commonv1.ContactLink{
						ContactId: "test_contact-id",
					},
					Amount: &money.Money{
						CurrencyCode: "USD",
						Units:        1000.00,
						Nanos:        0,
					},
					Cost: &money.Money{
						CurrencyCode: "USD",
						Units:        200,
						Nanos:        0,
					},
					ReferenceId:           "test_reference-id",
					BatchId:               "test_batch-id",
					ExternalTransactionId: "test_external-transaction-id",
					Outbound:              true,
				},
			},
			want: &commonv1.StatusResponse{
				Id:     "c2f4j7au6s7f91uqnojg",
				State:  commonv1.STATE_CREATED,
				Status: commonv1.STATUS_QUEUED,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctxService, err := getService(tt.name)
			//log ctxService
			log.Printf("ctxService = %v", ctxService.ctx)
			if err != nil {
				t.Errorf("getService() error = %v", err)
				return
			}

			pb, err := business.NewPaymentBusiness(ctxService.ctx, ctxService.srv, tt.fields.profileCli, tt.fields.partitionCli)

			if err != nil {
				t.Errorf("NewPaymentBusiness() error = %v", err)
				return
			}

			status, err := pb.Send(ctxService.ctx, tt.args.message)
			if err != nil {
				t.Errorf("Dispatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//log status

			log.Printf("Dispatch() status = %v", status)

			if status.Id != tt.want.Id {
				t.Errorf("Dispatch() status.Id = %v, want %v", status.Id, tt.want.Id)
			}

			if status.State != tt.want.State {
				t.Errorf("Dispatch() status.State = %v, want %v", status.State, tt.want.State)
			}

			if status.Status != tt.want.Status {
				t.Errorf("Dispatch() status.Status = %v, want %v", status.Status, tt.want.Status)
			}
		})
	}
}

func TestSendPaymentWithAmountMissing(t *testing.T) {
	profileCli := getProfileCli(t)
	partitionCli := getPartitionCli(t)

	type fields struct {
		ctxService   *ctxSrv
		profileCli   *profileV1.ProfileClient
		partitionCli *partitionV1.PartitionClient
	}

	type args struct {
		ctx     context.Context
		message *paymentV1.Payment
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *commonv1.StatusResponse
		wantErr bool
	}{
		{
			name: "SendWithAmountMissing",
			fields: fields{
				ctxService:   nil,
				profileCli:   profileCli,
				partitionCli: partitionCli,
			},
			args: args{
				ctx: nil,
				message: &paymentV1.Payment{
					Id: "c2f4j7au6s7f91uqnojz",
					Recipient: &commonv1.ContactLink{
						ContactId: "test_contact-id",
					},
					Amount: &money.Money{
						CurrencyCode: "",
						Units:        0,
						Nanos:        0,
					},
					Cost: &money.Money{
						CurrencyCode: "",
						Units:        0,
						Nanos:        0,
					},
					ReferenceId:           "test_reference-id",
					BatchId:               "test_batch-id",
					ExternalTransactionId: "test_external-transaction-id",
					Outbound:              true,
				},
			},
			want: &commonv1.StatusResponse{
				Id:     "c2f4j7au6s7f91uqnojz",
				State:  commonv1.STATE_CREATED,
				Status: commonv1.STATUS_QUEUED,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctxService, err := getService(tt.name)
			if err != nil {
				t.Errorf("getService() error = %v", err)
				return
			}

			pb, err := business.NewPaymentBusiness(ctxService.ctx, ctxService.srv, tt.fields.profileCli, tt.fields.partitionCli)

			if err != nil {
				t.Errorf("NewPaymentBusiness() error = %v", err)
				return
			}

			status, err := pb.Send(ctxService.ctx, tt.args.message)

			if err != nil {
				t.Errorf("Dispatch() error = %v, wantErr %v", err, tt.wantErr)
				//return
			}

			if status == nil {
				t.Errorf("Dispatch() status = %v, want %v", status, nil)
				//return
			}
		})
	}
}
