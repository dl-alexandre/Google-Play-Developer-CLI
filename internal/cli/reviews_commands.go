// Package cli provides reviews commands for gpd.
package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) addReviewsCommands() {
	reviewsCmd := &cobra.Command{
		Use:   "reviews",
		Short: "Review management commands",
		Long:  "List and reply to user reviews.",
	}

	reviewsCmd.AddCommand(
		c.newReviewsListCommand(),
		c.newReviewsReplyCommand(),
		c.newReviewsGetCommand(),
		c.newReviewsResponseCommand(),
		c.newReviewsCapabilitiesCommand(),
	)
	c.rootCmd.AddCommand(reviewsCmd)
}

func (c *CLI) newReviewsListCommand() *cobra.Command {
	var (
		minRating       int
		maxRating       int
		language        string
		startDate       string
		endDate         string
		scanLimit       int
		includeText     bool
		translationLang string
		pageSize        int64
		pageToken       string
		all             bool
	)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List user reviews",
		Long:  "List user reviews with optional filtering.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.reviewsList(cmd.Context(), minRating, maxRating, language, startDate, endDate,
				scanLimit, includeText, translationLang, pageSize, pageToken, all)
		},
	}
	listCmd.Flags().IntVar(&minRating, "min-rating", 0, "Minimum rating filter (1-5)")
	listCmd.Flags().IntVar(&maxRating, "max-rating", 0, "Maximum rating filter (1-5)")
	listCmd.Flags().StringVar(&language, "language", "", "Filter by review language")
	listCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (ISO 8601)")
	listCmd.Flags().StringVar(&endDate, "end-date", "", "End date (ISO 8601)")
	listCmd.Flags().IntVar(&scanLimit, "scan-limit", 100, "Maximum reviews to scan")
	listCmd.Flags().BoolVar(&includeText, "include-review-text", false, "Include review text in output")
	listCmd.Flags().StringVar(&translationLang, "translation-language", "", "Language for translated reviews")
	listCmd.Flags().Int64Var(&pageSize, "page-size", 50, "Results per page")
	listCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	addPaginationFlags(listCmd, &all)

	return listCmd
}

func (c *CLI) newReviewsReplyCommand() *cobra.Command {
	var (
		reviewID     string
		replyText    string
		templateFile string
		maxActions   int
		rateLimit    string
		dryRun       bool
	)

	replyCmd := &cobra.Command{
		Use:   "reply",
		Short: "Reply to a review",
		Long:  "Reply to a user review with optional templating.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.reviewsReply(cmd.Context(), reviewID, replyText, templateFile, maxActions, rateLimit, dryRun)
		},
	}
	replyCmd.Flags().StringVar(&reviewID, "review-id", "", "Review ID to reply to")
	replyCmd.Flags().StringVar(&replyText, "text", "", "Reply text")
	replyCmd.Flags().StringVar(&templateFile, "template-file", "", "Template file for reply")
	replyCmd.Flags().IntVar(&maxActions, "max-actions", 10, "Maximum replies per execution")
	replyCmd.Flags().StringVar(&rateLimit, "rate-limit", "5s", "Rate limit between replies")
	replyCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	return replyCmd
}

func (c *CLI) newReviewsGetCommand() *cobra.Command {
	var (
		getReviewID string
		getInclude  bool
		getLanguage string
	)

	getCmd := &cobra.Command{
		Use:   "get [review-id]",
		Short: "Get a review by ID",
		Long:  "Fetch a single review by ID with optional translated text.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reviewID := getReviewID
			if len(args) > 0 {
				reviewID = args[0]
			}
			if strings.TrimSpace(reviewID) == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError, "review ID is required").
					WithHint("Provide a review ID argument or use --review-id")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.reviewsGet(cmd.Context(), reviewID, getInclude, getLanguage)
		},
	}
	getCmd.Flags().StringVar(&getReviewID, "review-id", "", "Review ID to fetch")
	getCmd.Flags().BoolVar(&getInclude, "include-review-text", false, "Include review text in output")
	getCmd.Flags().StringVar(&getLanguage, "translation-language", "", "Language for translated review")

	return getCmd
}

func (c *CLI) newReviewsResponseCommand() *cobra.Command {
	responseCmd := &cobra.Command{
		Use:   "response",
		Short: "Review response commands",
		Long:  "Get or delete a developer response for a review.",
	}

	var responseReviewID string
	responseGetCmd := &cobra.Command{
		Use:   "get [review-id]",
		Short: "Get a review response",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reviewID := responseReviewID
			if len(args) > 0 {
				reviewID = args[0]
			}
			if strings.TrimSpace(reviewID) == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError, "review ID is required").
					WithHint("Provide a review ID argument or use --review-id")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.reviewsResponseGet(cmd.Context(), reviewID)
		},
	}
	responseGetCmd.Flags().StringVar(&responseReviewID, "review-id", "", "Review ID to fetch response for")

	responseForReviewCmd := &cobra.Command{
		Use:   "for-review [review-id]",
		Short: "Get a review response for a review",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reviewID := responseReviewID
			if len(args) > 0 {
				reviewID = args[0]
			}
			if strings.TrimSpace(reviewID) == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError, "review ID is required").
					WithHint("Provide a review ID argument or use --review-id")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.reviewsResponseGet(cmd.Context(), reviewID)
		},
	}
	responseForReviewCmd.Flags().StringVar(&responseReviewID, "review-id", "", "Review ID to fetch response for")

	responseDeleteCmd := &cobra.Command{
		Use:   "delete [review-id]",
		Short: "Delete a review response",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reviewID := responseReviewID
			if len(args) > 0 {
				reviewID = args[0]
			}
			if strings.TrimSpace(reviewID) == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError, "review ID is required").
					WithHint("Provide a review ID argument or use --review-id")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.reviewsResponseDelete(cmd.Context(), reviewID)
		},
	}
	responseDeleteCmd.Flags().StringVar(&responseReviewID, "review-id", "", "Review ID to delete response for")

	responseCmd.AddCommand(responseGetCmd, responseForReviewCmd, responseDeleteCmd)
	return responseCmd
}

func (c *CLI) newReviewsCapabilitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "capabilities",
		Short: "List review capabilities",
		Long:  "List review API capabilities and limitations.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.reviewsCapabilities(cmd.Context())
		},
	}
}

// reviewsListParams holds parameters for listing reviews.
type reviewsListParams struct {
	minRating       int
	maxRating       int
	scanLimit       int
	includeText     bool
	translationLang string
	pageSize        int64
	pageToken       string
	all             bool
}

// passesRatingFilter checks if a review passes the rating filter criteria.
func passesRatingFilter(review *androidpublisher.Review, minRating, maxRating int) bool {
	if minRating == 0 && maxRating == 0 {
		return true
	}
	if len(review.Comments) == 0 {
		return true
	}
	rating := int(review.Comments[0].UserComment.StarRating)
	if minRating > 0 && rating < minRating {
		return false
	}
	if maxRating > 0 && rating > maxRating {
		return false
	}
	return true
}

// buildReviewOutput creates the output map for a single review.
func buildReviewOutput(review *androidpublisher.Review, includeText bool) map[string]interface{} {
	reviewOutput := map[string]interface{}{
		"reviewId": review.ReviewId,
	}

	if len(review.Comments) == 0 {
		return reviewOutput
	}

	userComment := review.Comments[0].UserComment
	reviewOutput["rating"] = userComment.StarRating
	reviewOutput["language"] = userComment.ReviewerLanguage
	reviewOutput["lastModified"] = userComment.LastModified.Seconds

	if includeText {
		reviewOutput["text"] = userComment.Text
	}

	// Check for developer reply
	if len(review.Comments) > 1 && review.Comments[1].DeveloperComment != nil {
		reviewOutput["developerComment"] = map[string]interface{}{
			"text":         review.Comments[1].DeveloperComment.Text,
			"lastModified": review.Comments[1].DeveloperComment.LastModified.Seconds,
		}
	}

	return reviewOutput
}

// getNextPageToken extracts the next page token from a reviews response.
func getNextPageToken(resp *androidpublisher.ReviewsListResponse) string {
	if resp.TokenPagination != nil {
		return resp.TokenPagination.NextPageToken
	}
	return ""
}

func (c *CLI) reviewsList(ctx context.Context, minRating, maxRating int, _, _, _ string,
	scanLimit int, includeText bool, translationLang string, pageSize int64, pageToken string, all bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	params := reviewsListParams{
		minRating:       minRating,
		maxRating:       maxRating,
		scanLimit:       scanLimit,
		includeText:     includeText,
		translationLang: translationLang,
		pageSize:        pageSize,
		pageToken:       pageToken,
		all:             all,
	}

	return c.fetchAndOutputReviews(ctx, &params)
}

func (c *CLI) reviewsGet(ctx context.Context, reviewID string, includeText bool, translationLang string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	call := publisher.Reviews.Get(c.packageName, reviewID)
	if translationLang != "" {
		call = call.TranslationLanguage(translationLang)
	}

	resp, err := call.Context(ctx).Do()
	if err != nil {
		apiErr := errors.ClassifyAuthError(err)
		if apiErr == nil {
			apiErr = errors.NewAPIError(errors.CodeGeneralError, err.Error())
		}
		if apiErr.Code == errors.CodeValidationError || strings.Contains(apiErr.Message, "wrong format") {
			apiErr = apiErr.WithHint("Review IDs look like 'gp:AOqpTO...'. Use 'gpd reviews list' to retrieve a valid ID.")
		}
		result := output.NewErrorResult(apiErr).WithServices("androidpublisher")
		return c.Output(result)
	}

	result := output.NewResult(buildReviewOutput(resp, includeText))
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) reviewsResponseGet(ctx context.Context, reviewID string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	resp, err := publisher.Reviews.Get(c.packageName, reviewID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	if len(resp.Comments) > 1 && resp.Comments[1].DeveloperComment != nil {
		result := output.NewResult(map[string]interface{}{
			"reviewId": resp.ReviewId,
			"text":     resp.Comments[1].DeveloperComment.Text,
			"updated":  resp.Comments[1].DeveloperComment.LastModified.Seconds,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	result := output.NewErrorResult(errors.NewAPIError(errors.CodeNotFound, "review response not found").
		WithDetails(map[string]interface{}{"reviewId": reviewID})).WithServices("androidpublisher")
	return c.Output(result)
}

func (c *CLI) reviewsResponseDelete(_ context.Context, reviewID string) error {
	result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError, "review response deletion is not supported by the Google Play API").
		WithHint("Use gpd reviews reply to overwrite an existing response if needed").
		WithDetails(map[string]interface{}{"reviewId": reviewID})).WithServices("androidpublisher")
	return c.Output(result)
}

func (c *CLI) fetchAndOutputReviews(ctx context.Context, params *reviewsListParams) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	req := c.buildReviewsRequest(publisher, params)
	allReviews, scannedCount, filteredCount, nextToken, err := c.collectReviews(ctx, req, params)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	return c.outputReviewsResult(allReviews, scannedCount, filteredCount, params.pageToken, nextToken)
}

func (c *CLI) buildReviewsRequest(publisher *androidpublisher.Service, params *reviewsListParams) *androidpublisher.ReviewsListCall {
	req := publisher.Reviews.List(c.packageName).MaxResults(params.pageSize)
	if params.pageToken != "" {
		req = req.Token(params.pageToken)
	}
	if params.translationLang != "" {
		req = req.TranslationLanguage(params.translationLang)
	}
	return req
}

func (c *CLI) collectReviews(ctx context.Context, req *androidpublisher.ReviewsListCall, params *reviewsListParams) (reviews []map[string]interface{}, scanned, filtered int, nextToken string, err error) {
	var allReviews []map[string]interface{}
	scannedCount := 0
	filteredCount := 0

	for scannedCount < params.scanLimit {
		resp, err := req.Context(ctx).Do()
		if err != nil {
			return nil, 0, 0, "", err
		}

		for _, review := range resp.Reviews {
			scannedCount++
			if !passesRatingFilter(review, params.minRating, params.maxRating) {
				filteredCount++
				continue
			}
			allReviews = append(allReviews, buildReviewOutput(review, params.includeText))
		}

		nextToken = getNextPageToken(resp)
		if nextToken == "" || !params.all {
			break
		}
		req = req.Token(nextToken)
	}

	return allReviews, scannedCount, filteredCount, nextToken, nil
}

func (c *CLI) outputReviewsResult(reviews []map[string]interface{}, scanned, filtered int, pageToken, nextToken string) error {
	result := output.NewResult(reviews)
	result.WithServices("androidpublisher")
	result.WithPartial(scanned, filtered, 0)
	result.WithPagination(pageToken, nextToken)
	if strings.EqualFold(c.outputFormat, string(output.FormatTable)) {
		if err := c.renderReviewsTable(reviews); err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				"failed to render reviews table: "+err.Error()))
		}
		return nil
	}

	if len(reviews) == 0 {
		if scanned == 0 {
			result.WithWarnings("No reviews returned. This can mean the app has no reviews yet or access to reviews is restricted.")
		} else {
			result.WithWarnings("No reviews matched the current filters. Try adjusting rating filters or scan limits.")
		}
	} else if filtered > 0 {
		result.WithWarnings("Some reviews were filtered out by rating or date.")
	}

	return c.Output(result)
}

func (c *CLI) renderReviewsTable(reviews []map[string]interface{}) error {
	table := tablewriter.NewWriter(c.stdout)
	table.Header([]string{"reviewId", "rating", "language", "lastModified", "hasReply"})

	for _, review := range reviews {
		hasReply := "false"
		if _, ok := review["developerComment"]; ok {
			hasReply = "true"
		}

		if err := table.Append([]string{
			stringValue(review["reviewId"], "-"),
			stringValue(review["rating"], "-"),
			stringValue(review["language"], "-"),
			stringValue(review["lastModified"], "-"),
			hasReply,
		}); err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to append table row: %v", err))
		}
	}

	if err := table.Render(); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to render table: %v", err))
	}
	return nil
}

func (c *CLI) reviewsReply(ctx context.Context, reviewID, replyText, templateFile string, _ int, rateLimit string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		result := output.NewErrorResult(err.(*errors.APIError)).WithServices("androidpublisher")
		return c.Output(result)
	}

	// Parse rate limit
	rateDuration, err := time.ParseDuration(rateLimit)
	if err != nil {
		result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("invalid rate limit: %s", rateLimit))).WithServices("androidpublisher")
		return c.Output(result)
	}

	// Process template if provided
	if templateFile != "" {
		data, err := os.ReadFile(templateFile)
		if err != nil {
			result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
				fmt.Sprintf("failed to read template file: %v", err))).WithServices("androidpublisher")
			return c.Output(result)
		}
		replyText = string(data)
	}

	if replyText == "" {
		result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
			"reply text is required").WithHint("Provide --text or --template-file")).WithServices("androidpublisher")
		return c.Output(result)
	}

	// Process template variables
	replyText, err = processTemplate(replyText, map[string]string{
		"appName": c.packageName, // Would be actual app name
		"rating":  "5",           // Would come from review
		"locale":  "en-US",       // Would come from review
	})
	if err != nil {
		result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError, err.Error())).
			WithServices("androidpublisher")
		return c.Output(result)
	}

	// Check idempotency
	idempotencyKey := hashReply(reviewID, replyText)

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":         true,
			"action":         "reply",
			"reviewId":       reviewID,
			"text":           replyText,
			"idempotencyKey": idempotencyKey,
			"rateLimit":      rateDuration.String(),
			"package":        c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	// Get API client
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	// Apply rate limiting
	time.Sleep(rateDuration)

	// Build the reply request
	replyRequest := &androidpublisher.ReviewsReplyRequest{
		ReplyText: replyText,
	}

	// Post reply
	replyResp, err := publisher.Reviews.Reply(c.packageName, reviewID, replyRequest).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to reply: %v", err)))
	}
	_ = replyResp // Contains the result with lastEdited timestamp

	result := output.NewResult(map[string]interface{}{
		"success":        true,
		"reviewId":       reviewID,
		"text":           replyText,
		"idempotencyKey": idempotencyKey,
		"action":         "created",
		"package":        c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) reviewsCapabilities(_ context.Context) error {
	result := output.NewResult(map[string]interface{}{
		"requiredScopes":   []string{"https://www.googleapis.com/auth/androidpublisher"},
		"reviewWindowDays": 90,
		"maxReplyLength":   350,
		"supportedFilters": map[string]interface{}{
			"serverSide": []string{"translationLanguage"},
			"clientSide": []string{"rating", "dateRange", "language"},
		},
		"defaultScanLimit":  100,
		"templateVariables": []string{"{{appName}}", "{{rating}}", "{{locale}}"},
		"apiLimitations": []string{
			"Date range filtering is client-side only",
			"Review window limited to recent reviews",
			"Server-side filtering limited to translation language",
		},
	})
	return c.Output(result.WithServices("androidpublisher"))
}

// processTemplate processes template variables in the text.
func processTemplate(text string, vars map[string]string) (string, error) {
	// Find all template variables
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	matches := re.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		varName := match[1]
		value, ok := vars[varName]
		if !ok {
			return "", fmt.Errorf("missing template variable: {{%s}}", varName)
		}
		text = strings.ReplaceAll(text, match[0], value)
	}

	return text, nil
}

// hashReply creates an idempotency key for a reply.
func hashReply(reviewID, text string) string {
	h := sha256.New()
	h.Write([]byte(reviewID))
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))[:16]
}
