package models

import (
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/pitabwire/frame"
	"github.com/shopspring/decimal"
	money "google.golang.org/genproto/googleapis/type/money"
	"gorm.io/datatypes"
)

const (
	RouteModeTransmit   = "tx"
	RouteModeReceive    = "rx"
	RouteModeTransceive = "trx"

	RouteTypeAny       = "any"
	RouteTypeLongForm  = "l"
	RouteTypeShortForm = "s"
)

// Payment Table holds the payment details.
type Payment struct {
	frame.BaseModel

	SenderProfileID   string `gorm:"type:varchar(250)"`
	SenderProfileType string `gorm:"type:varchar(50)"`
	SenderContactID   string `gorm:"type:varchar(50)"`

	RecipientProfileID   string `gorm:"type:varchar(250)"`
	RecipientProfileType string `gorm:"type:varchar(50)"`
	RecipientContactID   string `gorm:"type:varchar(50)"`

	Amount        decimal.NullDecimal `gorm:"type:numeric" json:"amount"`
	TransactionId string              `gorm:"type:varchar(50)"`
	ReferenceId   string              `gorm:"type:varchar(50)"`
	BatchId       string              `gorm:"type:varchar(50)"`
	RouteID       string              `gorm:"type:varchar(50)"`
	Currency      string              `gorm:"type:varchar(10)"`
	PaymentType   string              `gorm:"type:varchar(10)"`
	Cost          *Cost
	ReleasedAt    *time.Time
	OutBound      bool
	Extra         datatypes.JSONMap `gorm:"index:,type:gin,option:jsonb_path_ops" json:"extra"`
}

func (model *Payment) IsReleased() bool {
	return model.ReleasedAt != nil && !model.ReleasedAt.IsZero()
}
func (model *Payment) ToApi(status *PaymentStatus, message map[string]string) *paymentV1.Payment {
	extra := make(map[string]string)
	extra["tenant_id"] = model.TenantID
	extra["partition_id"] = model.PartitionID
	extra["access_id"] = model.AccessID
	if model.IsReleased() {
		extra["ReleaseDate"] = model.ReleasedAt.String()
	}
	if len(message) != 0 {
		for key, val := range message {
			extra[key] = val
		}
	}

	source := &commonv1.ContactLink{
		ProfileType: model.SenderProfileType,
		ProfileId:   model.SenderProfileID,
		ContactId:   model.SenderContactID,
	}

	recipient := &commonv1.ContactLink{
		ProfileType: model.RecipientProfileType,
		ProfileId:   model.RecipientProfileID,
		ContactId:   model.RecipientContactID,
	}

	payment := paymentV1.Payment{
		Id:            model.ID,
		Source:        source,
		Recipient:     recipient,
		Amount:        &money.Money{CurrencyCode: model.Currency, Units: model.Amount.Decimal.CoefficientInt64()},
		TransactionId: model.TransactionId,
		ReferenceId:   model.ReferenceId,
		BatchId:       model.BatchId,
		Route:         model.RouteID,
		Status:        commonv1.STATUS(status.Status),
		Outbound:      model.OutBound,
		Extra:         extra,
		Cost: &money.Money{
			CurrencyCode: model.Cost.Currency,
			Units:        model.Cost.Amount.Decimal.CoefficientInt64(),
		},
	}

	return &payment
}

type Cost struct {
	frame.BaseModel
	PaymentID string              `gorm:"type:varchar(50)"`
	Amount    decimal.NullDecimal `gorm:"type:numeric" json:"amount"`
	Currency  string
	Extra     datatypes.JSONMap `gorm:"index:,type:gin,option:jsonb_path_ops" json:"extra"`
}

type PaymentStatus struct {
	frame.BaseModel
	PaymentID string            `gorm:"type:varchar(50)"`
	Extra     datatypes.JSONMap `gorm:"index:,type:gin,option:jsonb_path_ops" json:"extra"`
	State     int32
	Status    int32
}

func (model *PaymentStatus) ToStatusAPI() *commonv1.StatusResponse {
	extra := frame.DBPropertiesToMap(model.Extra)
	extra["CreatedAt"] = model.CreatedAt.String()
	extra["StatusID"] = model.PaymentID

	status := commonv1.StatusResponse{
		Id:     model.PaymentID,
		State:  commonv1.STATE(model.State),
		Status: commonv1.STATUS(model.Status),
		Extras: extra,
	}
	return &status
}

type Route struct {
	frame.BaseModel

	CounterID   string `gorm:"type:varchar(50)"`
	Name        string `gorm:"type:varchar(50)"`
	Description string `gorm:"type:text"`
	RouteType   string `gorm:"type:varchar(10)"`
	Mode        string `gorm:"type:varchar(10)"`
	Uri         string `gorm:"type:varchar(255)"`
}



type Prompt struct {
	frame.BaseModel

	SourceID string `gorm:"type:varchar(50)"`
	SourceProfileType string `gorm:"type:varchar(50)"`
	SourceContactID string `gorm:"type:varchar(50)"`

	RecipientID string `gorm:"type:varchar(50)"`
	RecipientProfileType string `gorm:"type:varchar(50)"`
	RecipientContactID string `gorm:"type:varchar(50)"`

	Amount *decimal.Decimal `gorm:"type:numeric"`
	DateCreated string `gorm:"type:varchar(50)"`
	DeviceID string `gorm:"type:varchar(50)"`
	State int32 `gorm:"type:integer"`
	Status int32 `gorm:"type:integer"`
	Route string `gorm:"type:varchar(50)"`
	Metadata map[string]string `gorm:"type:jsonb"`
	
}

func (model *Prompt) ToApi() *paymentV1.InitiatePromptRequest {
	return &paymentV1.InitiatePromptRequest{
		Source: &commonv1.ContactLink{
			ProfileType: model.SourceProfileType,
			ProfileId:   model.SourceID,
			ContactId:   model.SourceContactID,
		},
		Recipient: &commonv1.ContactLink{
			ProfileType: model.RecipientProfileType,
			ProfileId:   model.RecipientID,
			ContactId:   model.RecipientContactID,
		},
		Amount: &money.Money{Units: model.Amount.CoefficientInt64()},
		DateCreated: model.DateCreated,
		DeviceId:    model.DeviceID,
		State:       commonv1.STATE(model.State),
		Status:      commonv1.STATUS(model.Status),
		Route:       model.Route,
		Metadata:    model.Metadata,
	}
}

func (model *Prompt) ToApiStatus() *commonv1.StatusResponse {
	return &commonv1.StatusResponse{
		Id:     model.ID,
		State:  commonv1.STATE(model.State),
		Status: commonv1.STATUS(model.Status),
	}
}


