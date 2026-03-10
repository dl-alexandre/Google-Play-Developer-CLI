package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

// OrdersCmd contains order management commands.
type OrdersCmd struct {
	Get      OrdersGetCmd      `cmd:"" help:"Get order details"`
	Refund   OrdersRefundCmd   `cmd:"" help:"Refund an order"`
	BatchGet OrdersBatchGetCmd `cmd:"" name:"batch-get" help:"Batch get orders"`
}

// OrdersGetCmd gets order details by order ID.
type OrdersGetCmd struct {
	OrderID string `arg:"" help:"Order ID to fetch" required:""`
}

// orderData represents a simplified order for output.
type orderData struct {
	OrderID       string         `json:"orderId"`
	PackageName   string         `json:"packageName"`
	State         string         `json:"state,omitempty"`
	CreateTime    time.Time      `json:"createTime,omitempty"`
	LastEventTime time.Time      `json:"lastEventTime,omitempty"`
	TotalAmount   string         `json:"totalAmount,omitempty"`
	TaxAmount     string         `json:"taxAmount,omitempty"`
	Revenue       string         `json:"revenue,omitempty"`
	SalesChannel  string         `json:"salesChannel,omitempty"`
	PurchaseToken string         `json:"purchaseToken,omitempty"`
	LineItems     []lineItemData `json:"lineItems,omitempty"`
}

// lineItemData represents a simplified line item for output.
type lineItemData struct {
	ProductID    string `json:"productId,omitempty"`
	ProductTitle string `json:"productTitle,omitempty"`
	Type         string `json:"type,omitempty"`
	Total        string `json:"total,omitempty"`
	ListingPrice string `json:"listingPrice,omitempty"`
}

// Run executes the orders get command.
func (cmd *OrdersGetCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	if cmd.OrderID == "" {
		return errors.NewAPIError(errors.CodeValidationError, "order ID is required")
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

	var order *androidpublisher.Order
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		order, callErr = svc.Orders.Get(globals.Package, cmd.OrderID).Context(ctx).Do()
		return callErr
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get order: %v", err))
	}

	data := convertOrder(order, globals.Package)

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// convertOrder converts an API Order to orderData.
func convertOrder(order *androidpublisher.Order, packageName string) orderData {
	data := orderData{
		OrderID:       order.OrderId,
		PackageName:   packageName,
		State:         order.State,
		SalesChannel:  order.SalesChannel,
		PurchaseToken: order.PurchaseToken,
	}

	// Parse times
	if order.CreateTime != "" {
		t, err := time.Parse(time.RFC3339, order.CreateTime)
		if err == nil {
			data.CreateTime = t
		}
	}
	if order.LastEventTime != "" {
		t, err := time.Parse(time.RFC3339, order.LastEventTime)
		if err == nil {
			data.LastEventTime = t
		}
	}

	// Format money amounts
	if order.Total != nil {
		data.TotalAmount = formatMoney(order.Total)
	}
	if order.Tax != nil {
		data.TaxAmount = formatMoney(order.Tax)
	}
	if order.DeveloperRevenueInBuyerCurrency != nil {
		data.Revenue = formatMoney(order.DeveloperRevenueInBuyerCurrency)
	}

	// Convert line items
	if order.LineItems != nil {
		for _, item := range order.LineItems {
			if item == nil {
				continue
			}
			itemData := lineItemData{
				ProductID:    item.ProductId,
				ProductTitle: item.ProductTitle,
			}
			if item.Total != nil {
				itemData.Total = formatMoney(item.Total)
			}
			if item.ListingPrice != nil {
				itemData.ListingPrice = formatMoney(item.ListingPrice)
			}
			// Determine type from details
			if item.SubscriptionDetails != nil {
				itemData.Type = "subscription"
			} else if item.PaidAppDetails != nil {
				itemData.Type = "paid_app"
			} else if item.OneTimePurchaseDetails != nil {
				itemData.Type = "one_time"
			}
			data.LineItems = append(data.LineItems, itemData)
		}
	}

	return data
}

// formatMoney formats a Money object as string.
func formatMoney(m *androidpublisher.Money) string {
	if m == nil {
		return ""
	}
	// Convert nanos to cents
	cents := float64(m.Nanos) / 10000000
	units := float64(m.Units)
	total := units + cents/100
	return fmt.Sprintf("%s %.2f", m.CurrencyCode, total)
}

// OrdersRefundCmd refunds an order.
type OrdersRefundCmd struct {
	OrderID string `arg:"" help:"Order ID to refund" required:""`
	Revoke  bool   `help:"Revoke the purchase (grant refund but cancel the item/subscription)"`
}

// refundResult represents a refund operation result.
type refundResult struct {
	OrderID   string    `json:"orderId"`
	Refunded  bool      `json:"refunded"`
	Revoked   bool      `json:"revoked,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Run executes the orders refund command.
func (cmd *OrdersRefundCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	if cmd.OrderID == "" {
		return errors.NewAPIError(errors.CodeValidationError, "order ID is required")
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

	err = client.DoWithRetry(ctx, func() error {
		call := svc.Orders.Refund(globals.Package, cmd.OrderID).Context(ctx)
		if cmd.Revoke {
			call = call.Revoke(true)
		}
		return call.Do()
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to refund order: %v", err))
	}

	data := refundResult{
		OrderID:   cmd.OrderID,
		Refunded:  true,
		Revoked:   cmd.Revoke,
		Timestamp: time.Now(),
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// OrdersBatchGetCmd batch gets orders by order IDs.
type OrdersBatchGetCmd struct {
	OrderIDs []string `arg:"" help:"Order IDs to fetch (comma-separated or multiple)" optional:""`
	FromFile string   `help:"File containing order IDs (one per line)" type:"existingfile"`
	All      bool     `help:"Fetch all pages (if paginated)"`
}

// batchOrderResult represents the batch get response.
type batchOrderResult struct {
	Orders     []orderData `json:"orders"`
	TotalCount int         `json:"totalCount"`
	NotFound   []string    `json:"notFound,omitempty"`
}

// Run executes the orders batch get command.
func (cmd *OrdersBatchGetCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	// Collect order IDs from args and/or file
	var orderIDs []string

	if cmd.FromFile != "" {
		fileIDs, err := readLinesFromFile(cmd.FromFile)
		if err != nil {
			return err
		}
		orderIDs = append(orderIDs, fileIDs...)
	}

	// Add any order IDs from command arguments
	for _, id := range cmd.OrderIDs {
		// Split by comma if multiple IDs passed as single arg
		parts := strings.Split(id, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				orderIDs = append(orderIDs, part)
			}
		}
	}

	if len(orderIDs) == 0 {
		return errors.NewAPIError(errors.CodeValidationError, "at least one order ID is required (provide as arguments or --from-file)")
	}

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	var batchResp *androidpublisher.BatchGetOrdersResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		batchResp, callErr = svc.Orders.Batchget(globals.Package).OrderIds(orderIDs...).Context(ctx).Do()
		return callErr
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to batch get orders: %v", err))
	}

	// Process response
	var orders []orderData
	for _, order := range batchResp.Orders {
		if order == nil {
			continue
		}
		orders = append(orders, convertOrder(order, globals.Package))
	}

	// Determine which order IDs were not found
	foundIDs := make(map[string]bool)
	for i := range orders {
		order := orders[i]
		foundIDs[order.OrderID] = true
	}
	var notFound []string
	for _, id := range orderIDs {
		if !foundIDs[id] {
			notFound = append(notFound, id)
		}
	}

	result := batchOrderResult{
		Orders:     orders,
		TotalCount: len(orders),
		NotFound:   notFound,
	}

	return outputResult(
		output.NewResult(result).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// readLinesFromFile reads a file and returns each non-empty line as a string.
func readLinesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to open file: %v", err))
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			// Log close error but don't override original error
			_ = cerr
		}
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	return lines, nil
}
