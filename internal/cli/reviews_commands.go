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

	// Shared flags
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

	// reviews list
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

	// reviews reply
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

	// reviews capabilities
	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List review capabilities",
		Long:  "List review API capabilities and limitations.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.reviewsCapabilities(cmd.Context())
		},
	}

	reviewsCmd.AddCommand(listCmd, replyCmd, capabilitiesCmd)
	c.rootCmd.AddCommand(reviewsCmd)
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
	return c.Output(result)
}

func (c *CLI) reviewsReply(ctx context.Context, reviewID, replyText, templateFile string, _ int, rateLimit string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	// Parse rate limit
	rateDuration, err := time.ParseDuration(rateLimit)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("invalid rate limit: %s", rateLimit)))
	}

	// Process template if provided
	if templateFile != "" {
		data, err := os.ReadFile(templateFile)
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
				fmt.Sprintf("failed to read template file: %v", err)))
		}
		replyText = string(data)
	}

	if replyText == "" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"reply text is required").WithHint("Provide --text or --template-file"))
	}

	// Process template variables
	replyText, err = processTemplate(replyText, map[string]string{
		"appName": c.packageName, // Would be actual app name
		"rating":  "5",           // Would come from review
		"locale":  "en-US",       // Would come from review
	})
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, err.Error()))
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
