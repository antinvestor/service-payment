package models

import (
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/pitabwire/frame"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	money "google.golang.org/genproto/googleapis/type/money"
)

// Payment Table holds the payment details
type Payment struct {
	frame.BaseModel

	SenderProfileID   string `gorm:"type:varchar(250)"`
	SenderProfileType string `gorm:"type:varchar(50)"`
	SenderContactID   string `gorm:"type:varchar(50)"`

	RecipientProfileID   string `gorm:"type:varchar(250)"`
	RecipientProfileType string `gorm:"type:varchar(50)"`
	RecipientContactID   string `gorm:"type:varchar(50)"`

	Id                    string `gorm:"type:varchar(50)"`
	TransactionId         string `gorm:"type:varchar(50)"`
	ReferenceId           string `gorm:"type:varchar(50)"`
	BatchId               string `gorm:"type:varchar(50)"`
	ExternalTransactionId string `gorm:"type:varchar(50)"`
	Route                 string `gorm:"type:varchar(50)"`

	Source        *commonv1.ContactLink
	Recipient     *commonv1.ContactLink
	Amount        decimal.NullDecimal `gorm:"type:numeric" json:"amount"`
	Cost          decimal.NullDecimal `gorm:"type:numeric" json:"cost"`
	Currency      string              `gorm:"type:varchar(10)" json:"currency"`
	State         commonv1.STATE
	Status        commonv1.STATUS
	DateCreated   *time.Time
	DateProcessed *time.Time
	Outbound      bool
	Extra         datatypes.JSONMap
}

func (model *Payment) IsProcessed() bool {
	return model.DateProcessed != nil && !model.DateProcessed.IsZero()
}

func (model *Payment) ToApi() *paymentV1.Payment {

	extra := make(map[string]string)
	extra["tenant_id"] = model.TenantID
	extra["partition_id"] = model.PartitionID
	extra["access_id"] = model.AccessID

	if model.IsProcessed() {
		extra["ProcessedDate"] = model.DateProcessed.String()
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
		Id:                    model.ID,
		Source:                source,
		Recipient:             recipient,
		Amount:                &money.Money{CurrencyCode: model.Currency, Units: model.Amount.Decimal.CoefficientInt64()},
		Cost:                  &money.Money{CurrencyCode: model.Currency, Units: model.Cost.Decimal.CoefficientInt64()},
		State:                 model.State,
		Status:                model.Status,
		TransactionId:         model.TransactionId,
		ReferenceId:           model.ReferenceId,
		BatchId:               model.BatchId,
		ExternalTransactionId: model.ExternalTransactionId,
		Route:                 model.Route,
		Outbound:              model.Outbound,
		DateCreated:           model.DateCreated.Format(time.RFC3339),
		DateProcessed:         model.DateProcessed.Format(time.RFC3339),
		Extra:                 extra,
	}

	return &payment
}

type PaymentStatus struct {
	frame.BaseModel
	PaymentID   string `gorm:"type:varchar(50)"`
	Extra       datatypes.JSONMap
	State       int32
	Status      int32
}

func (model *PaymentStatus) ToStatusAPI() *commonv1.StatusResponse {
	extra := frame.DBPropertiesToMap(model.Extra)
	extra["CreatedAt"] = model.CreatedAt.String()
	extra["StatusID"] = model.PaymentID

	status := commonv1.StatusResponse{
		Id:          model.PaymentID,
		State:       commonv1.STATE(model.State),
		Status:      commonv1.STATUS(model.Status),
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
