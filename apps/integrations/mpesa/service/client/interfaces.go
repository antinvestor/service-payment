package client

import "context"

// MpesaClient defines the interface for M-Pesa Daraja API operations.
type MpesaClient interface {
	// STKPush initiates an STK Push (Lipa Na M-Pesa) request to the customer's phone
	STKPush(ctx context.Context, creds *MpesaCredentials, req *STKPushRequest) (*STKPushResponse, error)

	// B2CPayment initiates a Business-to-Customer payment (disbursement)
	B2CPayment(ctx context.Context, creds *MpesaCredentials, req *B2CRequest) (*B2CResponse, error)
}
