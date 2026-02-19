package client

import "context"

// AirtelClient defines the interface for Airtel Money API operations.
type AirtelClient interface {
	// CollectionPush initiates a USSD push collection request
	CollectionPush(ctx context.Context, creds *AirtelCredentials, req *CollectionRequest) (*CollectionResponse, error)

	// Disburse initiates a disbursement to a customer
	Disburse(ctx context.Context, creds *AirtelCredentials, req *DisbursementRequest) (*DisbursementResponse, error)

	// TransactionStatus checks the status of a transaction
	TransactionStatus(ctx context.Context, creds *AirtelCredentials, transactionID string) (*StatusResponse, error)
}
