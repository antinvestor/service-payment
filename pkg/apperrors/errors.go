package apperrors

import (
	"fmt"
	"strings"
)

// Error code constants for different categories.
const (
	// System error codes (1-10).
	ErrorCodeSystemFailure        = 1
	ErrorCodeUnspecifiedID        = 2
	ErrorCodeUnspecifiedReference = 3
	ErrorCodeBadDataSupplied      = 4

	// Ledger error codes (11-20).
	ErrorCodeLedgerNotFound = 11

	// Account error codes (21-30).
	ErrorCodeAccountNotFound            = 21
	ErrorCodeAccountsNotFound           = 22
	ErrorCodeAccountsCurrencyUnknown    = 23
	ErrorCodeAccountWithReferenceExists = 24

	// Transaction error codes (31-60).
	ErrorCodeTransactionNotFound               = 31
	ErrorCodeTransactionEntriesNotFound        = 32
	ErrorCodeTransactionEntryHasZeroAmount     = 33
	ErrorCodeTransactionAccountsDifferCurrency = 34
	ErrorCodeTransactionAlreadyExists          = 35
	ErrorCodeTransactionHasNonZeroSum          = 36
	ErrorCodeTransactionHasInvalidDrCrEntry    = 37
	ErrorCodeTransactionIsConflicting          = 38
	ErrorCodeTransactionTypeNotReversible      = 39

	// Search error codes (61-70).
	ErrorCodeSearchNamespaceUnknown       = 61
	ErrorCodeSearchQueryHasInvalidFormat  = 62
	ErrorCodeSearchQueryHasInvalidKeys    = 63
	ErrorCodeSearchQueryResultsNotCasting = 64

	// Billing catalog error codes (71-80).
	ErrorCodeCatalogVersionNotFound = 71
	ErrorCodeCatalogVersionRetired  = 72
	ErrorCodeCatalogNotPublished    = 73
	ErrorCodePlanNotFound           = 74
	ErrorCodeComponentNotFound      = 75
	ErrorCodeInvalidPricingModel    = 76

	// Billing subscription error codes (91-100).
	ErrorCodeSubscriptionNotFound  = 91
	ErrorCodeSubscriptionNotActive = 92

	// Billing usage error codes (101-110).
	ErrorCodeUsageEventDuplicate = 101
	ErrorCodeUsageEventInvalid   = 102

	// Billing rating/discount error codes (111-120).
	ErrorCodeRatingFailed    = 111
	ErrorCodeDiscountInvalid = 112

	// Billing invoice error codes (121-130).
	ErrorCodeInvoiceNotFound      = 121
	ErrorCodeInvoiceAlreadyIssued = 122
	ErrorCodeInvoiceNotIssuable   = 123
	ErrorCodeInvoiceNotVoidable   = 124
	ErrorCodeInvoiceNotPayable    = 125

	// Billing credit error codes (131-140).
	ErrorCodeCreditGrantNotFound     = 131
	ErrorCodeCreditInsufficientFunds = 132
	ErrorCodeCreditExpired           = 133

	// Billing run error codes (141-150).
	ErrorCodeBillingRunNotFound      = 141
	ErrorCodeBillingRunAlreadyExists = 142
	ErrorCodeBillingRunFailed        = 143
)

type ApplicationError interface {
	error
	ErrorCode() int32
	String() string
	Extend(message string) ApplicationError
	Override(errs ...error) ApplicationError
}

type applicationLedgerError struct {
	Code         int32
	CodeOffset   int32
	Message      string
	ExtraMessage string
}

// DefaultCodeOffset is added to error codes to produce the final ErrorCode().
const DefaultCodeOffset int32 = 200

func NewApplicationError(code int32, message string) ApplicationError {
	return &applicationLedgerError{code, DefaultCodeOffset, message, ""}
}

func (e applicationLedgerError) Error() string {
	if e.ExtraMessage != "" {
		return fmt.Sprintf("%d  : - %s  \n extra info : %s", e.ErrorCode(), e.Message, e.ExtraMessage)
	}
	return fmt.Sprintf("%d  : - %s  ", e.ErrorCode(), e.Message)
}

// ErrorCode returns the unique Code of the error.
func (e applicationLedgerError) ErrorCode() int32 {
	return e.CodeOffset + e.Code
}

// String implementation supports logging.
func (e applicationLedgerError) String() string {
	return e.Error()
}

// Extend default Message.
func (e applicationLedgerError) Extend(message string) ApplicationError {
	return &applicationLedgerError{e.Code, e.CodeOffset, e.Message, message}
}

// Override default Message.
func (e applicationLedgerError) Override(errs ...error) ApplicationError {
	errorStrings := make([]string, len(errs))

	for i, err := range errs {
		errorStrings[i] = err.Error()
	}
	return &applicationLedgerError{e.Code, e.CodeOffset, e.Message, strings.Join(errorStrings, "\n")}
}

var (
	ErrSystemFailure        = NewApplicationError(ErrorCodeSystemFailure, "Internal System failure")
	ErrUnspecifiedID        = NewApplicationError(ErrorCodeUnspecifiedID, "No ID was supplied")
	ErrUnspecifiedReference = NewApplicationError(ErrorCodeUnspecifiedReference, "No reference was supplied")
	ErrBadDataSupplied      = NewApplicationError(ErrorCodeBadDataSupplied, "Invalid data format was supplied")

	ErrLedgerNotFound = NewApplicationError(ErrorCodeLedgerNotFound, "Ledger with reference/id not found")

	ErrAccountNotFound  = NewApplicationError(ErrorCodeAccountNotFound, "Account with reference/id not found")
	ErrAccountsNotFound = NewApplicationError(
		ErrorCodeAccountsNotFound,
		"Accounts with references/ids were not found",
	)
	ErrAccountsCurrencyUnknown = NewApplicationError(
		ErrorCodeAccountsCurrencyUnknown,
		"Supplied account currency is unknown",
	)
	ErrAccountWithReferenceExists = NewApplicationError(
		ErrorCodeAccountWithReferenceExists,
		"An account with the given reference exists",
	)

	ErrTransactionNotFound = NewApplicationError(
		ErrorCodeTransactionNotFound,
		"Transaction with reference/id not found",
	)
	ErrTransactionEntriesNotFound = NewApplicationError(
		ErrorCodeTransactionEntriesNotFound,
		"Transaction no entries found",
	)
	ErrTransactionEntryHasZeroAmount = NewApplicationError(
		ErrorCodeTransactionEntryHasZeroAmount,
		"Transaction entry has zero amount",
	)
	ErrTransactionAccountsDifferCurrency = NewApplicationError(
		ErrorCodeTransactionAccountsDifferCurrency,
		"Transaction accounts have different currencies",
	)
	ErrTransactionAlreadyExists = NewApplicationError(
		ErrorCodeTransactionAlreadyExists,
		"Transaction with reference/id already exists",
	)
	ErrTransactionHasNonZeroSum = NewApplicationError(
		ErrorCodeTransactionHasNonZeroSum,
		"Transaction has a non zero sum",
	)
	ErrTransactionHasInvalidDrCrEntry = NewApplicationError(
		ErrorCodeTransactionHasInvalidDrCrEntry,
		"Transaction has a invalid count of dr/cr entries",
	)
	ErrTransactionIsConflicting = NewApplicationError(
		ErrorCodeTransactionIsConflicting,
		"Transaction is conflicting",
	)
	ErrTransactionTypeNotReversible = NewApplicationError(
		ErrorCodeTransactionTypeNotReversible,
		"Transaction type is not reversible",
	)

	ErrSearchNamespaceUnknown = NewApplicationError(
		ErrorCodeSearchNamespaceUnknown,
		"Search namespace provided is unknown",
	)
	ErrSearchQueryHasInvalidFormat = NewApplicationError(
		ErrorCodeSearchQueryHasInvalidFormat,
		"Search query has invalid format",
	)
	ErrSearchQueryHasInvalidKeys = NewApplicationError(
		ErrorCodeSearchQueryHasInvalidKeys,
		"Search query has invalid keys",
	)
	ErrSearchQueryResultsNotCasting = NewApplicationError(
		ErrorCodeSearchQueryResultsNotCasting,
		"Search query results not casting",
	)

	// ErrCatalogVersionNotFound indicates the requested catalog version does not exist.
	ErrCatalogVersionNotFound = NewApplicationError(ErrorCodeCatalogVersionNotFound, "Catalog version not found")
	// ErrCatalogVersionRetired indicates the catalog version has been retired.
	ErrCatalogVersionRetired = NewApplicationError(ErrorCodeCatalogVersionRetired, "Catalog version is retired")
	// ErrCatalogNotPublished indicates the catalog version is not yet published.
	ErrCatalogNotPublished = NewApplicationError(ErrorCodeCatalogNotPublished, "Catalog version is not published")
	// ErrPlanNotFound indicates the requested plan does not exist.
	ErrPlanNotFound = NewApplicationError(ErrorCodePlanNotFound, "Plan not found")
	// ErrComponentNotFound indicates the requested component does not exist.
	ErrComponentNotFound = NewApplicationError(ErrorCodeComponentNotFound, "Component not found")
	// ErrInvalidPricingModel indicates an unsupported pricing model was specified.
	ErrInvalidPricingModel = NewApplicationError(ErrorCodeInvalidPricingModel, "Invalid pricing model")
	// ErrSubscriptionNotFound indicates the requested subscription does not exist.
	ErrSubscriptionNotFound = NewApplicationError(ErrorCodeSubscriptionNotFound, "Subscription not found")
	// ErrSubscriptionNotActive indicates the subscription is not in an active state.
	ErrSubscriptionNotActive = NewApplicationError(ErrorCodeSubscriptionNotActive, "Subscription is not active")
	// ErrUsageEventDuplicate indicates a duplicate usage event was submitted.
	ErrUsageEventDuplicate = NewApplicationError(ErrorCodeUsageEventDuplicate, "Duplicate usage event")
	// ErrUsageEventInvalid indicates the usage event data is invalid.
	ErrUsageEventInvalid = NewApplicationError(ErrorCodeUsageEventInvalid, "Invalid usage event")
	// ErrRatingFailed indicates a rating calculation error.
	ErrRatingFailed = NewApplicationError(ErrorCodeRatingFailed, "Rating calculation failed")
	// ErrDiscountInvalid indicates an invalid discount configuration.
	ErrDiscountInvalid = NewApplicationError(ErrorCodeDiscountInvalid, "Invalid discount configuration")
	// ErrInvoiceNotFound indicates the requested invoice does not exist.
	ErrInvoiceNotFound = NewApplicationError(ErrorCodeInvoiceNotFound, "Invoice not found")
	// ErrInvoiceAlreadyIssued indicates the invoice has already been issued.
	ErrInvoiceAlreadyIssued = NewApplicationError(ErrorCodeInvoiceAlreadyIssued, "Invoice has already been issued")
	// ErrInvoiceNotIssuable indicates the invoice cannot be issued in its current state.
	ErrInvoiceNotIssuable = NewApplicationError(
		ErrorCodeInvoiceNotIssuable, "Invoice is not in a state that can be issued")
	// ErrInvoiceNotVoidable indicates the invoice cannot be voided in its current state.
	ErrInvoiceNotVoidable = NewApplicationError(
		ErrorCodeInvoiceNotVoidable, "Invoice cannot be voided in its current state")
	// ErrInvoiceNotPayable indicates the invoice must be issued before payment can be recorded.
	ErrInvoiceNotPayable = NewApplicationError(
		ErrorCodeInvoiceNotPayable, "Invoice must be in issued state to record payment")
	// ErrCreditGrantNotFound indicates the requested credit grant does not exist.
	ErrCreditGrantNotFound = NewApplicationError(ErrorCodeCreditGrantNotFound, "Credit grant not found")
	// ErrCreditInsufficientFunds indicates insufficient credit balance.
	ErrCreditInsufficientFunds = NewApplicationError(ErrorCodeCreditInsufficientFunds, "Insufficient credit balance")
	// ErrCreditExpired indicates the credit grant has expired.
	ErrCreditExpired = NewApplicationError(ErrorCodeCreditExpired, "Credit grant has expired")
	// ErrBillingRunNotFound indicates the requested billing run does not exist.
	ErrBillingRunNotFound = NewApplicationError(ErrorCodeBillingRunNotFound, "Billing run not found")
	// ErrBillingRunAlreadyExists indicates a billing run already exists for the given period.
	ErrBillingRunAlreadyExists = NewApplicationError(
		ErrorCodeBillingRunAlreadyExists, "Billing run already exists for this period")
	// ErrBillingRunFailed indicates the billing run has failed.
	ErrBillingRunFailed = NewApplicationError(ErrorCodeBillingRunFailed, "Billing run failed")
)
