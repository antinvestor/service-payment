package business_test

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"net"

// 	"github.com/antinvestor/apis/go/common"

// 	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
// 	tenancyv1 "buf.build/gen/go/antinvestor/tenancy/protocolbuffers/go/tenancy/v1"
// 	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
// 	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"

// 	money "google.golang.org/genproto/googleapis/type/money"

// 	"testing"
// 	"time"

// 	"github.com/antinvestor/service-payments/apps/default/config"
// 	business "github.com/antinvestor/service-payments/apps/default/service/business"
// 	"github.com/antinvestor/service-payments/apps/default/service/events"
// 	"github.com/testcontainers/testcontainers-go"
// 	"github.com/testcontainers/testcontainers-go/wait"
// 	"go.uber.org/mock/gomock"

// 	"github.com/pitabwire/frame"
// )

// func getService(serviceName string) (*ctxSrv, error) {
// 	ctx := context.Background()

// 	req := testcontainers.ContainerRequest{
// 		Image:        "postgres:latest",
// 		ExposedPorts: []string{"5432/tcp"},
// 		Env: map[string]string{
// 			"POSTGRES_USER":     "ant",
// 			"POSTGRES_PASSWORD": "secret",
// 			"POSTGRES_DB":       "service_payment",
// 		},
// 		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(5 * time.Minute),
// 	}

// 	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
// 		ContainerRequest: req,
// 		Started:          true,
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to start container: %w", err)
// 	}
// 	// return
// 	defer func() {
// 		if terminateErr := postgresC.Terminate(ctx); terminateErr != nil {
// 			// Log error but continue cleanup
// 			fmt.Printf("Error terminating postgres container: %v\n", terminateErr)
// 		}
// 	}()

// 	mappedPort, err := postgresC.MappedPort(ctx, "5432")
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get mapped port: %w", err)
// 	}

// 	hostIP, err := postgresC.Host(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get container host: %w", err)
// 	}
// 	dbURL := fmt.Sprintf(
// 		"postgres://ant:secret@%s/service_payment?sslmode=disable",
// 		net.JoinHostPort(hostIP, mappedPort.Port()),
// 	)
// 	testDB := frame.WithDatastoreConnection(dbURL, false)

// 	var pcfg config.PaymentConfig
// 	// _ = frame.ConfigFillFromEnv(&pcfg)

// 	ctx2, service := frame.NewService(serviceName, testDB, frame.WithConfig(&pcfg), frame.WithNoopDriver())

// 	m := make(map[string]string)
// 	m["sub"] = "testing"
// 	m["tenant_id"] = "test_tenant-id"
// 	m["partition_id"] = "test_partition-id"
// 	m["access_id"] = "test_access-id"

// 	claims := frame.ClaimsFromMap(m)
// 	ctx = claims.ClaimsToContext(ctx2)

// 	eventList := frame.WithRegisterEvents(
// 		&events.PaymentSave{Service: service},
// 		&events.StatusSave{Service: service})
// 	service.Init(ctx, eventList)
// 	_ = service.Run(ctx, "")
// 	return &ctxSrv{
// 		ctx,
// 		service,
// 	}, nil
// }

// type ctxSrv struct {
// 	ctx context.Context
// 	srv *frame.Service
// }

// func getProfileCli(t *testing.T) *profilev1.ProfileClient {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
// 	mockProfileService := profilev1.NewMockProfileServiceClient(ctrl)
// 	mockProfileService.EXPECT().
// 		GetById(gomock.Any(), gomock.Any()).
// 		Return(&profilev1.GetByIdResponse{
// 			Data: &profilev1.ProfileObject{
// 				Id: "test_profile-id",
// 			},
// 		}, nil).AnyTimes()
// 	mockProfileService.EXPECT().
// 		GetByContact(gomock.Any(), gomock.Any()).
// 		Return(&profilev1.GetByContactResponse{
// 			Data: &profilev1.ProfileObject{
// 				Id: "test_profile-id",
// 			},
// 		}, nil).AnyTimes()

// 	profileCli := profilev1.Init(&common.GrpcClientBase{}, mockProfileService)
// 	return profileCli
// }

// func getTenancyCli(t *testing.T) *tenancyV1.TenancyClient {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
// 	mockTenancyService := tenancyV1.NewMockTenancyServiceClient(ctrl)

// 	mockTenancyService.EXPECT().
// 		GetAccess(gomock.Any(), gomock.Any()).
// 		Return(&tenancyV1.GetAccessResponse{Data: &tenancyV1.AccessObject{
// 			AccessId: "test_access-id",
// 			Partition: &tenancyV1.PartitionObject{
// 				Id:       "test_partition-id",
// 				TenantId: "test_tenant-id",
// 			},
// 		}}, nil).AnyTimes()

// 	profileCli := tenancyV1.Init(&common.GrpcClientBase{}, mockTenancyService)
// 	return profileCli
// }

// func TestNewPaymentBusiness_Success(t *testing.T) {
// 	profileCli := getProfileCli(t)
// 	tenancyCli := getTenancyCli(t)

// 	type args struct {
// 		ctxService   *ctxSrv
// 		profileCli   *profilev1.ProfileClient
// 		tenancyCli *tenancyV1.TenancyClient
// 	}
// 	tests := []struct {
// 		name      string
// 		args      args
// 		want      business.PaymentBusiness
// 		expectErr bool
// 	}{
// 		{
// 			name: "NewPaymentBusiness",
// 			args: args{
// 				ctxService:   nil,
// 				profileCli:   profileCli,
// 				tenancyCli: tenancyCli},
// 			expectErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			service, err := getService(tt.name)
// 			if err != nil {
// 				t.Errorf("failed to get service: %v", err)
// 			}

// 			pb, err := business.NewPaymentBusiness(service.ctx, service.srv, tt.args.profileCli, tt.args.tenancyCli)

// 			if err != nil {
// 				t.Errorf("expected no error, got %v", err)
// 			}

// 			if pb == nil {
// 				t.Errorf("expected payment business, got nil")
// 			}
// 		})
// 	}
// }

// func TestNewPaymentBusinessWithNils(t *testing.T) {
// 	profileCli := getProfileCli(t)
// 	tenancyCli := getTenancyCli(t)

// 	type args struct {
// 		ctxService   *ctxSrv
// 		profileCli   *profilev1.ProfileClient
// 		tenancyCli *tenancyV1.TenancyClient
// 	}
// 	tests := []struct {
// 		name      string
// 		args      args
// 		want      business.PaymentBusiness
// 		expectErr bool
// 	}{
// 		{
// 			name: "NewPaymentBusinessWithNils",
// 			args: args{
// 				ctxService:   nil,
// 				profileCli:   nil,
// 				tenancyCli: nil},
// 			expectErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			service, err := getService(tt.name)
// 			if err != nil {
// 				t.Errorf("failed to get service: %v", err)
// 			}
// 			pb, err := business.NewPaymentBusiness(service.ctx, nil, profileCli, tenancyCli)

// 			if !errors.Is(err, business.ErrInitializationFail) {
// 				t.Errorf("expected ErrInitializationFail, got %v", err)
// 			}

// 			if pb != nil {
// 				t.Errorf("expected nil PaymentBusiness instance, got %v", pb)
// 			}
// 		})
// 	}
// }

// func TestSendPaymentWithValidData(t *testing.T) {
// 	profileCli := getProfileCli(t)
// 	tenancyCli := getTenancyCli(t)

// 	type fields struct {
// 		ctxService   *ctxSrv
// 		profileCli   *profilev1.ProfileClient
// 		tenancyCli *tenancyV1.TenancyClient
// 	}

// 	type args struct {
// 		ctx     context.Context
// 		message *paymentv1.Payment
// 	}

// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    *commonv1.StatusResponse
// 		wantErr bool
// 	}{
// 		{
// 			name: "Send",
// 			fields: fields{
// 				ctxService:   nil,
// 				profileCli:   profileCli,
// 				tenancyCli: tenancyCli,
// 			},
// 			args: args{
// 				ctx: nil,
// 				message: &paymentv1.Payment{
// 					Id: "c2f4j7au6s7f91uqnojg",
// 					Recipient: &commonv1.ContactLink{
// 						ContactId: "test_contact-id",
// 					},
// 					Amount: &money.Money{
// 						CurrencyCode: "USD",
// 						Units:        1000.00,
// 						Nanos:        0,
// 					},
// 					Cost: &money.Money{
// 						CurrencyCode: "USD",
// 						Units:        200,
// 						Nanos:        0,
// 					},
// 					ReferenceId:           "test_reference-id",
// 					BatchId:               "test_batch-id",
// 					ExternalTransactionId: "test_external-transaction-id",
// 					Outbound:              true,
// 				},
// 			},
// 			want: &commonv1.StatusResponse{
// 				Id:     "c2f4j7au6s7f91uqnojg",
// 				State:  commonv1.STATE_CREATED,
// 				Status: commonv1.STATUS_QUEUED,
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctxService, err := getService(tt.name)
// 			// log ctxService
// 			if err != nil {
// 				t.Errorf("getService() error = %v", err)
// 				return
// 			}

// 			pb, err := business.NewPaymentBusiness(
// 				ctxService.ctx,
// 				ctxService.srv,
// 				tt.fields.profileCli,
// 				tt.fields.tenancyCli,
// 			)

// 			if err != nil {
// 				t.Errorf("NewPaymentBusiness() error = %v", err)
// 				return
// 			}

// 			status, err := pb.Send(ctxService.ctx, tt.args.message)
// 			if err != nil {
// 				t.Errorf("Dispatch() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			// log status			log.Printf("Dispatch() status = %v", status)

// 			if status.GetId() != tt.want.GetId() {
// 				t.Errorf("Dispatch() status.Id = %v, want %v", status.GetId(), tt.want.GetId())
// 			}

// 			if status.GetState() != tt.want.GetState() {
// 				t.Errorf("Dispatch() status.State = %v, want %v", status.GetState(), tt.want.GetState())
// 			}

// 			if status.GetStatus() != tt.want.GetStatus() {
// 				t.Errorf("Dispatch() status.Status = %v, want %v", status.GetStatus(), tt.want.GetStatus())
// 			}
// 		})
// 	}
// }

// func TestSendPaymentWithAmountMissing(t *testing.T) {
// 	profileCli := getProfileCli(t)
// 	tenancyCli := getTenancyCli(t)

// 	type fields struct {
// 		ctxService   *ctxSrv
// 		profileCli   *profilev1.ProfileClient
// 		tenancyCli *tenancyV1.TenancyClient
// 	}

// 	type args struct {
// 		ctx     context.Context
// 		message *paymentv1.Payment
// 	}

// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    *commonv1.StatusResponse
// 		wantErr bool
// 	}{
// 		{
// 			name: "SendWithAmountMissing",
// 			fields: fields{
// 				ctxService:   nil,
// 				profileCli:   profileCli,
// 				tenancyCli: tenancyCli,
// 			},
// 			args: args{
// 				ctx: nil,
// 				message: &paymentv1.Payment{
// 					Id: "c2f4j7au6s7f91uqnojz",
// 					Recipient: &commonv1.ContactLink{
// 						ContactId: "test_contact-id",
// 					},
// 					Amount: &money.Money{
// 						CurrencyCode: "",
// 						Units:        0,
// 						Nanos:        0,
// 					},
// 					Cost: &money.Money{
// 						CurrencyCode: "",
// 						Units:        0,
// 						Nanos:        0,
// 					},
// 					ReferenceId:           "test_reference-id",
// 					BatchId:               "test_batch-id",
// 					ExternalTransactionId: "test_external-transaction-id",
// 					Outbound:              true,
// 				},
// 			},
// 			want: &commonv1.StatusResponse{
// 				Id:     "c2f4j7au6s7f91uqnojz",
// 				State:  commonv1.STATE_CREATED,
// 				Status: commonv1.STATUS_QUEUED,
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctxService, err := getService(tt.name)
// 			if err != nil {
// 				t.Errorf("getService() error = %v", err)
// 				return
// 			}

// 			pb, err := business.NewPaymentBusiness(
// 				ctxService.ctx,
// 				ctxService.srv,
// 				tt.fields.profileCli,
// 				tt.fields.tenancyCli,
// 			)

// 			if err != nil {
// 				t.Errorf("NewPaymentBusiness() error = %v", err)
// 				return
// 			}

// 			status, err := pb.Send(ctxService.ctx, tt.args.message)

// 			if err != nil {
// 				t.Errorf("Dispatch() error = %v, wantErr %v", err, tt.wantErr)
// 				// return
// 			}

// 			if status == nil {
// 				t.Errorf("Dispatch() status = %v, want %v", status, nil)
// 				// return
// 			}
// 		})
// 	}
// }
