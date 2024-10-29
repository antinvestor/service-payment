package client

import (
	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentv1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/shopspring/decimal"
	money "google.golang.org/genproto/googleapis/type/money"
)

type ContactLink struct {
	ProfileID       string `json:"profile_id"`
	ProfileName     string `json:"profile_name"`
	ProfileType     string `json:"profile_type"`
	ContactID       string `json:"contact_id"`
	ContactDetail   string `json:"contact_detail"`
}



func (c *ContactLink) Populate(link *commonv1.ContactLink) ContactLink {

	c.ProfileID = link.GetProfileId()
	c.ProfileName = link.GetProfileName()
	c.ProfileType = link.GetProfileType()
	c.ContactID = link.GetContactId()
	c.ContactDetail = link.GetDetail()
	return *c
}

type PaymentRequest struct {
	ID          string            `json:"id"`
	Sender      ContactLink       `json:"sender"`
	Recipient   ContactLink       `json:"recipient"`
	Amount        decimal.NullDecimal `gorm:"type:numeric" json:"amount"`
	TransactionId string              `gorm:"type:varchar(50)"`
	ReferenceId   string              `gorm:"type:varchar(50)"`
	Currency      string              `gorm:"type:varchar(10)"`
	PaymentType   string              `gorm:"type:varchar(10)"`
	MetaData    map[string]string `json:"meta_data"`
}

func (p *PaymentRequest) SenderLink() *commonv1.ContactLink {
	return &commonv1.ContactLink{
		ProfileId:   p.Sender.ProfileID,
		ProfileName: p.Sender.ProfileName,
		ProfileType: p.Sender.ProfileType,
		ContactId:   p.Sender.ContactID,
		Detail:      p.Sender.ContactDetail,
	}
}

func (p *PaymentRequest) RecipientLink() *commonv1.ContactLink {
	return &commonv1.ContactLink{
		ProfileId:   p.Recipient.ProfileID,
		ProfileName: p.Recipient.ProfileName,
		ProfileType: p.Recipient.ProfileType,
		ContactId:   p.Recipient.ContactID,
		Detail:      p.Recipient.ContactDetail,
	}
}

// convertDecimalToMoney converts a decimal.NullDecimal to *money.Money
func convertDecimalToMoney(amount decimal.NullDecimal, currency string) *money.Money {
	if !amount.Valid {
		return nil
	}
	return & money.Money{
		Units: amount.Decimal.IntPart(),
		CurrencyCode: currency,
	}
}
		

func (p *PaymentRequest) ToApiObject() *paymentv1.Payment {
	return &paymentv1.Payment{
		Id:            p.ID,
		Source:        p.SenderLink(),
		Recipient:     p.RecipientLink(),
		Amount:        convertDecimalToMoney(p.Amount, p.Currency),
		TransactionId: p.TransactionId,
		ReferenceId:   p.ReferenceId,
		
	}
}
	

