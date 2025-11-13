package models

import (
	"encoding/json"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/payment/v1"
	"gorm.io/datatypes"

	"github.com/antinvestor/service-payments/service/utility"
	"github.com/pitabwire/frame/data"
	"github.com/shopspring/decimal"
)

const (
	RouteModeTransmit   = "tx"
	RouteModeReceive    = "rx"
	RouteModeTransceive = "trx"
)

// Payment Table holds the payment details.
type Payment struct {
	data.BaseModel

	SenderProfileID   string `gorm:"type:varchar(250)"`
	SenderProfileType string `gorm:"type:varchar(50)"`
	SenderContactID   string `gorm:"type:varchar(50)"`

	RecipientProfileID   string `gorm:"type:varchar(250)"`
	RecipientProfileType string `gorm:"type:varchar(50)"`
	RecipientContactID   string `gorm:"type:varchar(50)"`

	Amount        decimal.NullDecimal `gorm:"type:numeric"                          json:"amount"`
	TransactionID string              `gorm:"type:varchar(50)"`
	ReferenceID   string              `gorm:"type:varchar(50)"`
	BatchID       string              `gorm:"type:varchar(50)"`
	RouteID       string              `gorm:"type:varchar(50)"`
	Currency      string              `gorm:"type:varchar(10)"`
	PaymentType   string              `gorm:"type:varchar(10)"`
	CostIDs       []string            `gorm:"type:text[]"                           json:"cost_ids"`
	ReleasedAt    *time.Time
	OutBound      bool
	Extra         data.JSONMap `gorm:"index:,type:gin;option:jsonb_path_ops" json:"extra"`
}

func (model *Payment) IsReleased() bool {
	return model.ReleasedAt != nil && !model.ReleasedAt.IsZero()
}
func (model *Payment) ToAPI(status *Status, message map[string]string) *paymentv1.Payment {
	extra := data.JSONMap{}
	extra["tenant_id"] = model.TenantID
	extra["partition_id"] = model.PartitionID
	extra["access_id"] = model.AccessID
	if model.IsReleased() {
		extra["ReleaseDate"] = model.ReleasedAt.String()
	}
	for k, v := range message {
		extra[k] = v
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

	amountMoney := utility.ToMoney(model.Currency, model.Amount.Decimal)

	payment := paymentv1.Payment{
		Id:            model.ID,
		Source:        source,
		Recipient:     recipient,
		Amount:        &amountMoney,
		TransactionId: model.TransactionID,
		ReferenceId:   model.ReferenceID,
		BatchId:       model.BatchID,
		Route:         model.RouteID,
		Status:        commonv1.STATUS(status.Status),
		Outbound:      model.OutBound,
		Extra:         extra.ToProtoStruct(),
	}

	// Costs will be added separately after fetching from repository

	return &payment
}

type Cost struct {
	data.BaseModel
	PaymentID string              `gorm:"type:varchar(50)"`
	Amount    decimal.NullDecimal `gorm:"type:numeric"                          json:"amount"`
	Currency  string
	Extra     data.JSONMap `gorm:"index:,type:gin;option:jsonb_path_ops" json:"extra"`
}

// Unified Status model for all entities
// Replaces PaymentStatus, PromptStatus, PaymentLinkStatus

type Status struct {
	data.BaseModel
	EntityID   string       `gorm:"type:varchar(50)"`
	EntityType string       `gorm:"type:varchar(50)"`
	Extra      data.JSONMap `gorm:"index:,type:gin;option:jsonb_path_ops" json:"extra"`
	State      int32
	Status     int32
}

func (s *Status) ToAPI() *commonv1.StatusResponse {
	return &commonv1.StatusResponse{
		Id:     s.EntityID,
		State:  commonv1.STATE(s.State),
		Status: commonv1.STATUS(s.Status),
		Extras: s.Extra.ToProtoStruct(),
	}
}

// Deprecated: Use Status instead
// type PaymentStatus struct { ... }
// type PromptStatus struct { ... }
// type PaymentLinkStatus struct { ... }

type Route struct {
	data.BaseModel

	CounterID   string `gorm:"type:varchar(50)"`
	Name        string `gorm:"type:varchar(50)"`
	Description string `gorm:"type:text"`
	RouteType   string `gorm:"type:varchar(10)"`
	Mode        string `gorm:"type:varchar(10)"`
	URI         string `gorm:"type:varchar(255)"`
}

type Account struct {
	data.BaseModel
	AccountNumber string `gorm:"type:varchar(50)"`
	CountryCode   string `gorm:"type:varchar(50)"`
	Name          string `gorm:"type:varchar(50)"`
}

func (a *Account) ToAPI() *paymentv1.Account {
	return &paymentv1.Account{
		AccountNumber: a.AccountNumber,
		CountryCode:   a.CountryCode,
		Name:          a.Name,
	}
}

type Prompt struct {
	data.BaseModel
	ID                string `gorm:"type:varchar(50)"`
	SourceID          string `gorm:"type:varchar(50)"`
	SourceProfileType string `gorm:"type:varchar(50)"`
	SourceContactID   string `gorm:"type:varchar(50)"`

	RecipientID          string              `gorm:"type:varchar(50)"`
	RecipientProfileType string              `gorm:"type:varchar(50)"`
	RecipientContactID   string              `gorm:"type:varchar(50)"`
	Amount               decimal.NullDecimal `gorm:"type:numeric"                          json:"amount"`
	DateCreated          string              `gorm:"type:varchar(50)"`
	DeviceID             string              `gorm:"type:varchar(50)"`
	State                int32               `gorm:"type:integer"`
	Status               int32               `gorm:"type:integer"`
	Route                string              `gorm:"type:varchar(50)"`
	AccountID            string              `gorm:"type:varchar(50)"`
	Account              Account             `gorm:"foreignKey:AccountID;references:ID"`
	Extra                data.JSONMap        `gorm:"index:,type:gin;option:jsonb_path_ops" json:"extra"`
}

func (model *Prompt) getRecipientAccount() *paymentv1.Account {
	// This function no longer fetches from the database. Ensure the Account field is preloaded if needed.
	if model.AccountID != "" && model.Account.ID != "" {
		return &paymentv1.Account{
			AccountNumber: model.Account.AccountNumber,
			CountryCode:   model.Account.CountryCode,
			Name:          model.Account.Name,
		}
	}
	return &paymentv1.Account{}
}

func (model *Prompt) ToAPI(message map[string]string) *paymentv1.InitiatePromptRequest {
	extra := make(data.JSONMap)
	extra["tenant_id"] = model.TenantID
	extra["partition_id"] = model.PartitionID
	extra["access_id"] = model.AccessID
	extra["PromptID"] = model.ID

	for k, v := range message {
		extra[k] = v
	}

	amountMoney := utility.ToMoney(extra.GetString("currency"), model.Amount.Decimal)

	prompt := paymentv1.InitiatePromptRequest{
		Id: model.ID,
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
		Amount:           &amountMoney,
		DateCreated:      model.DateCreated,
		DeviceId:         model.DeviceID,
		State:            commonv1.STATE(model.State),
		Status:           commonv1.STATUS(model.Status),
		Route:            model.Route,
		RecipientAccount: model.getRecipientAccount(),
		Extra:            extra.ToProtoStruct(),
	}

	return &prompt
}

func (model *Prompt) ToAPIStatus() *commonv1.StatusResponse {
	return &commonv1.StatusResponse{
		Id:     model.ID,
		State:  commonv1.STATE(model.State),
		Status: commonv1.STATUS(model.Status),
	}
}

// PaymentLink represents a payment link with associated customers and notifications.
type PaymentLink struct {
	data.BaseModel

	ExpiryDate      time.Time       `gorm:"type:date"         json:"expiryDate"`
	SaleDate        time.Time       `gorm:"type:date"         json:"saleDate"`
	PaymentLinkType string          `gorm:"type:varchar(20)"  json:"paymentLinkType"`
	SaleType        string          `gorm:"type:varchar(20)"  json:"saleType"`
	Name            string          `gorm:"type:varchar(100)" json:"name"`
	Description     string          `gorm:"type:text"         json:"description"`
	ExternalRef     string          `gorm:"type:varchar(50)"  json:"externalRef"`
	PaymentLinkRef  string          `gorm:"type:varchar(50)"  json:"paymentLinkRef"`
	RedirectURL     string          `gorm:"type:varchar(255)" json:"redirectURL"`
	AmountOption    string          `gorm:"type:varchar(20)"  json:"amountOption"`
	Amount          decimal.Decimal `gorm:"type:numeric"      json:"amount"`
	Currency        string          `gorm:"type:varchar(10)"  json:"currency"`
	Customers       datatypes.JSON  `gorm:"type:jsonb"        json:"customers"` // stores []Customer as JSON
	Notifications   datatypes.JSON  `gorm:"type:jsonb"        json:"notifications"`
}

// Customer represents a customer for a payment link.
type Customer struct {
	FirstName           string `json:"firstName"`
	LastName            string `json:"lastName"`
	Email               string `json:"email"`
	PhoneNumber         string `json:"phoneNumber"`
	FirstAddress        string `json:"firstAddress"`
	CountryCode         string `json:"countryCode"`
	PostalOrZipCode     string `json:"postalOrZipCode"`
	CustomerExternalRef string `json:"customerExternalRef"`
}

// enum type for notification types

// NotificationType is an enum for notification types.
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "EMAIL"
	NotificationTypeSMS   NotificationType = "SMS"
)

// String returns the string representation of the NotificationType.
func (n NotificationType) String() string {
	return string(n)
}

// AllNotificationTypes returns all valid NotificationType values.
func AllNotificationTypes() []NotificationType {
	return []NotificationType{
		NotificationTypeEmail,
		NotificationTypeSMS,
	}
}

// IsValid checks if the NotificationType is valid.
func (n NotificationType) IsValid() bool {
	switch n {
	case NotificationTypeEmail, NotificationTypeSMS:
		return true
	default:
		return false
	}
}

func (model *PaymentLink) ToAPI(message map[string]string) *paymentv1.CreatePaymentLinkRequest {
	extra := make(data.JSONMap)
	extra["tenant_id"] = model.TenantID
	extra["partition_id"] = model.PartitionID
	extra["access_id"] = model.AccessID
	extra["PaymentLinkID"] = model.ID

	for k, v := range message {
		extra[k] = v
	}

	amountMoney := utility.ToMoney(model.Currency, model.Amount)

	paymentLink := paymentv1.PaymentLink{
		Id:              model.ID,
		ExpiryDate:      model.ExpiryDate.String(),
		SaleDate:        model.SaleDate.String(),
		PaymentLinkType: model.PaymentLinkType,
		SaleType:        model.SaleType,
		Name:            model.Name,
		Description:     model.Description,
		ExternalRef:     model.ExternalRef,
		PaymentLinkRef:  model.PaymentLinkRef,
		RedirectUrl:     model.RedirectURL,
		AmountOption:    model.AmountOption,
		Amount:          &amountMoney,
		Currency:        model.Currency,
	}

	customers := make([]*paymentv1.Customer, 0)
	if len(model.Customers) > 0 {
		var customerList []Customer
		err := json.Unmarshal(model.Customers, &customerList)
		if err == nil {
			for _, customer := range customerList {
				extraC := data.JSONMap{"email": customer.Email}
				customers = append(customers, &paymentv1.Customer{
					Source: &commonv1.ContactLink{
						ProfileName: customer.FirstName + " " + customer.LastName,
						ContactId:   customer.PhoneNumber,
						Extras:      extraC.ToProtoStruct(),
					},
					FirstAddress:        customer.FirstAddress,
					CountryCode:         customer.CountryCode,
					PostalOrZipCode:     customer.PostalOrZipCode,
					CustomerExternalRef: customer.CustomerExternalRef,
				})
			}
		}
	}

	createPaymentLinkRequest := &paymentv1.CreatePaymentLinkRequest{
		PaymentLink:   &paymentLink,
		Customers:     customers,
		Notifications: make([]paymentv1.NotificationType, 0),
	}
	if len(model.Notifications) > 0 {
		var notificationTypes []NotificationType
		err := json.Unmarshal(model.Notifications, &notificationTypes)
		if err == nil {
			for _, notificationType := range notificationTypes {
				if notificationType.IsValid() {
					createPaymentLinkRequest.Notifications = append(
						createPaymentLinkRequest.Notifications,
						topaymentv1NotificationType(notificationType.String()),
					)
				}
			}
		}
	}
	return createPaymentLinkRequest
}

// Helper to map string to paymentv1.NotificationType enum.
func topaymentv1NotificationType(s string) paymentv1.NotificationType {
	switch s {
	case "email":
		return paymentv1.NotificationType_NOTIFICATION_TYPE_EMAIL
	case "sms":
		return paymentv1.NotificationType_NOTIFICATION_TYPE_SMS
	default:
		return paymentv1.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED
	}
}
