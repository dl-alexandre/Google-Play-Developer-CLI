package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

// ============================================================================
// External Transactions Commands
// ============================================================================

// ExternalTransactionsCmd contains external transaction management commands.
// External transactions are used for alternative billing and reporting
// transactions that occur outside of Google Play's billing system.
type ExternalTransactionsCmd struct {
	Create ExternalTransactionsCreateCmd `cmd:"" help:"Create an external transaction"`
	Get    ExternalTransactionsGetCmd    `cmd:"" help:"Get an external transaction"`
	Refund ExternalTransactionsRefundCmd `cmd:"" help:"Refund an external transaction"`
}

// ============================================================================
// External Transactions Create Command
// ============================================================================

// ExternalTransactionsCreateCmd creates a new external transaction.
type ExternalTransactionsCreateCmd struct {
	ExternalTransactionID    string `help:"Unique external transaction ID (1-63 chars, a-zA-Z0-9_-)" required:""`
	TransactionTime          string `help:"Transaction time (RFC3339)" required:""`
	PriceMicros              string `help:"Price in micros (e.g., 990000 for $0.99)" required:""`
	Currency                 string `help:"Currency code (ISO 4217)" required:""`
	TaxMicros                string `help:"Tax amount in micros (default: 0)" default:"0"`
	RegionCode               string `help:"Two-letter region code (ISO-3166-1 Alpha-2) for tax address" required:""`
	AdministrativeArea       string `help:"Administrative area (required for India)"`
	ExternalTransactionToken string `help:"External transaction token from alternative billing flow"`
	OneTime                  bool   `help:"One-time transaction (default: true)" default:"true"`
	TransactionProgramCode   int64  `help:"Transaction program code for partner programs"`
	File                     string `help:"JSON file with external transaction data (alternative to flags)" type:"existingfile"`
}

// Run executes the create external transaction command.
func (cmd *ExternalTransactionsCreateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	var externalTxn *androidpublisher.ExternalTransaction

	if cmd.File != "" {
		// Load from JSON file
		fileData, err := os.ReadFile(cmd.File)
		if err != nil {
			return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
		}
		if err := json.Unmarshal(fileData, &externalTxn); err != nil {
			return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse JSON: %v", err)).
				WithHint("Ensure the file contains valid ExternalTransaction JSON")
		}
	} else {
		// Build from flags
		externalTxn = &androidpublisher.ExternalTransaction{
			PackageName:           globals.Package,
			ExternalTransactionId: cmd.ExternalTransactionID,
			TransactionTime:       cmd.TransactionTime,
		}

		// Set amounts
		externalTxn.OriginalPreTaxAmount = &androidpublisher.Price{
			PriceMicros: cmd.PriceMicros,
			Currency:    cmd.Currency,
		}
		externalTxn.OriginalTaxAmount = &androidpublisher.Price{
			PriceMicros: cmd.TaxMicros,
			Currency:    cmd.Currency,
		}

		// Set user tax address (required)
		externalTxn.UserTaxAddress = &androidpublisher.ExternalTransactionAddress{
			RegionCode: cmd.RegionCode,
		}
		if cmd.AdministrativeArea != "" {
			externalTxn.UserTaxAddress.AdministrativeArea = cmd.AdministrativeArea
		}

		// Set transaction type
		if cmd.OneTime {
			externalTxn.OneTimeTransaction = &androidpublisher.OneTimeExternalTransaction{}
			if cmd.ExternalTransactionToken != "" {
				externalTxn.OneTimeTransaction.ExternalTransactionToken = cmd.ExternalTransactionToken
			}
		}

		// Set transaction program code if provided
		if cmd.TransactionProgramCode > 0 {
			externalTxn.TransactionProgramCode = cmd.TransactionProgramCode
		}
	}

	var created *androidpublisher.ExternalTransaction
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		call := svc.Externaltransactions.Createexternaltransaction(globals.Package, externalTxn).
			ExternalTransactionId(cmd.ExternalTransactionID)
		created, callErr = call.Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create external transaction: %v", err))
	}

	return outputResult(
		output.NewResult(created).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// ============================================================================
// External Transactions Get Command
// ============================================================================

// ExternalTransactionsGetCmd gets an external transaction by ID.
type ExternalTransactionsGetCmd struct {
	ExternalTransactionID string `arg:"" help:"External transaction ID" required:""`
}

// Run executes the get external transaction command.
func (cmd *ExternalTransactionsGetCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	name := fmt.Sprintf("applications/%s/externalTransactions/%s", globals.Package, cmd.ExternalTransactionID)

	var externalTxn *androidpublisher.ExternalTransaction
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		externalTxn, callErr = svc.Externaltransactions.Getexternaltransaction(name).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get external transaction: %v", err))
	}

	return outputResult(
		output.NewResult(externalTxn).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// ============================================================================
// External Transactions Refund Command
// ============================================================================

// ExternalTransactionsRefundCmd refunds or partially refunds an external transaction.
type ExternalTransactionsRefundCmd struct {
	ExternalTransactionID string `arg:"" help:"External transaction ID" required:""`
	RefundTime            string `help:"Refund time (RFC3339, default: now)"`
	RefundID              string `help:"Unique refund ID (required for partial refund)"`
	PartialRefund         bool   `help:"Perform partial refund"`
	RefundPriceMicros     string `help:"Refund amount in micros for partial refund"`
	RefundCurrency        string `help:"Refund currency for partial refund (defaults to original)"`
}

// Run executes the refund external transaction command.
func (cmd *ExternalTransactionsRefundCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	name := fmt.Sprintf("applications/%s/externalTransactions/%s", globals.Package, cmd.ExternalTransactionID)

	refundReq := &androidpublisher.RefundExternalTransactionRequest{}

	// Handle partial refund
	if cmd.PartialRefund {
		if cmd.RefundPriceMicros == "" {
			return errors.NewAPIError(errors.CodeValidationError, "refund price is required for partial refund").
				WithHint("Provide --refund-price-micros flag")
		}
		if cmd.RefundID == "" {
			return errors.NewAPIError(errors.CodeValidationError, "refund ID is required for partial refund").
				WithHint("Provide --refund-id flag with a unique identifier")
		}

		currency := cmd.RefundCurrency
		if currency == "" {
			currency = "USD" // Default, will be validated by API
		}

		refundReq.PartialRefund = &androidpublisher.PartialRefund{
			RefundId: cmd.RefundID,
			RefundPreTaxAmount: &androidpublisher.Price{
				PriceMicros: cmd.RefundPriceMicros,
				Currency:    currency,
			},
		}
	} else {
		refundReq.FullRefund = &androidpublisher.FullRefund{}
	}

	// Set refund time (API requires this)
	if cmd.RefundTime != "" {
		refundReq.RefundTime = cmd.RefundTime
	} else {
		refundReq.RefundTime = time.Now().UTC().Format(time.RFC3339)
	}

	var refunded *androidpublisher.ExternalTransaction
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		refunded, callErr = svc.Externaltransactions.Refundexternaltransaction(name, refundReq).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to refund external transaction: %v", err))
	}

	data := map[string]interface{}{
		"externalTransactionId": cmd.ExternalTransactionID,
		"refunded":              true,
		"partialRefund":         cmd.PartialRefund,
		"transaction":           refunded,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}
