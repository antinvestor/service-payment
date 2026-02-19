package client

import "context"

// MtnClient defines the interface for MTN MoMo API operations.
type MtnClient interface {
	// RequestToPay initiates a request-to-pay (collection) from a customer
	RequestToPay(ctx context.Context, creds *MtnCredentials, req *RequestToPayRequest) error

	// GetRequestToPayStatus checks the status of a request-to-pay
	GetRequestToPayStatus(ctx context.Context, creds *MtnCredentials, referenceID string) (*RequestToPayStatus, error)

	// Transfer initiates a disbursement transfer to a customer
	Transfer(ctx context.Context, creds *MtnCredentials, req *TransferRequest) error

	// GetTransferStatus checks the status of a transfer
	GetTransferStatus(ctx context.Context, creds *MtnCredentials, referenceID string) (*TransferStatus, error)
}
