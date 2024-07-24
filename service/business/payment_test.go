package business

import (
	"context"
	"fmt"
	"github.com/antinvestor/apis/go/common"
	commonv1 "github.com/antinvestor/apis/go/common/v1"
	partitionV1 "github.com/antinvestor/apis/go/partition/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments-v1/service/config"

	"github.com/antinvestor/service-payments-v1/service/events"
	"github.com/pitabwire/frame"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
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

	mappedPort, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	hostIP, err := postgresC.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	dbURL := fmt.Sprintf("postgres://ant:secret@%s:%s/service_notification?sslmode=disable", hostIP, mappedPort.Port())
	testDb := frame.DatastoreCon(dbURL, false)

	var pcfg config.PaymentConfig
	_ = frame.ConfigProcess("", &pcfg)

	ctx, service := frame.NewService(serviceName, testDb, frame.Config(&pcfg), frame.NoopDriver())

	m := make(map[string]string)
	m["sub"] = "testing"
	m["tenant_id"] = "test_tenant-id"
	m["partition_id"] = "test_partition-id"
	m["access_id"] = "test_access-id"

	//claims := frame.ClaimsFromMap(m)
	//ctx = claims.ClaimsToContext(ctx)

	eventList := frame.RegisterEvents(
		&events.PaymentSave{Service: service},
		&events.PaymentStatusSave{Service: service})
	service.Init(eventList)
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

func TestNewPaymentBusiness(t *testing.T) {
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
		want      PaymentBusiness
		expectErr bool
	}{
		{
			name: "NewPaymentBusiness",
			args: args{
				ctxService:  nil,
				profileCli:   profileCli,
				partitionCli: partitionCli},
			expectErr: false,
		},
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
			ctxService, err := getService(tt.name)
			if (err != nil) != tt.expectErr {
				t.Errorf("getService() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if err == nil {
				tt.args.ctxService = ctxService
			}

			_, err = NewPaymentBusiness(tt.args.ctxService.ctx, tt.args.ctxService.srv, tt.args.profileCli, tt.args.partitionCli)
			if (err != nil) != tt.expectErr {
				t.Errorf("NewPaymentBusiness() error = %v, wantErr %v", err, tt.expectErr)
				return
			}

		})


	}

}

func TestPaymentBusiness_Dispatch(t *testing.T) {

	profileCli := getProfileCli(t)
	partitionCli := getPartitionCli(t)

	type fields struct {
		ctxService   *ctxSrv
		profileCli   *profileV1.ProfileClient
		partitionCli *partitionV1.PartitionClient
	}

	type args struct {
		ctx context.Context
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
			name: "Dispatch",
			fields: fields{
				ctxService:   getService("Dispatch"),
				profileCli:   profileCli,
				partitionCli: partitionCli,
			},
			args: args{
				ctx: context.Background(),
				message: &paymentV1.Payment{
					Id: "test_payment-id",

					}
				},

