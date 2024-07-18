package models

import (
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	money "google.golang.org/genproto/googleapis/type/money"
	"github.com/pitabwire/frame"
	"gorm.io/datatypes"
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

	Id                    string            `gorm:"type:varchar(50)"`
	TransactionId         string            `gorm:"type:varchar(50)"`
	ReferenceId           string            `gorm:"type:varchar(50)"`
	BatchId               string            `gorm:"type:varchar(50)"`
	ExternalTransactionId string            `gorm:"type:varchar(50)"`
	Route                 string            `gorm:"type:varchar(50)"`

	Source                *commonv1.ContactLink
	Recipient             *commonv1.ContactLink
	Amount                *money.Money
	Cost                  *money.Money
	State                 commonv1.STATE
	Status                commonv1.STATUS
	DateCreated           *time.Time
	DateProcessed         *time.Time
	Outbound              bool
	Extra                 datatypes.JSONMap
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
		Amount:                model.Amount,
		Cost:                  model.Cost,
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
