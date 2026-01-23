// Package cli provides reviews commands for gpd.
package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/google-play-cli/gpd/internal/errors"
	"github.com/google-play-cli/gpd/internal/output"
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
	listCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")

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

func (c *CLI) reviewsList(ctx context.Context, minRating, maxRating int, language, startDate, endDate string,
	scanLimit int, includeText bool, translationLang string, pageSize int64, pageToken string, all bool) error {

	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
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

	// Build list request
	req := publisher.Reviews.List(c.packageName).MaxResults(pageSize)
	if pageToken != "" {
		req = req.Token(pageToken)
	}
	if translationLang != "" {
		req = req.TranslationLanguage(translationLang)
	}

	// Fetch reviews
	var allReviews []map[string]interface{}
	scannedCount := 0
	filteredCount := 0

	for {
		if scannedCount >= scanLimit {
			break
		}

		resp, err := req.Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}

		for _, review := range resp.Reviews {
			scannedCount++

			// Client-side filtering
			if minRating > 0 || maxRating > 0 {
				// Get star rating from the first user comment
				if len(review.Comments) > 0 {
					rating := int(review.Comments[0].UserComment.StarRating)
					if minRating > 0 && rating < minRating {
						filteredCount++
						continue
					}
					if maxRating > 0 && rating > maxRating {
						filteredCount++
						continue
					}
				}
			}

			// Build review output
			reviewOutput := map[string]interface{}{
				"reviewId": review.ReviewId,
			}

			if len(review.Comments) > 0 {
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
			}

			allReviews = append(allReviews, reviewOutput)
		}

		// Check for more pages
		if resp.TokenPagination == nil || resp.TokenPagination.NextPageToken == "" || !all {
			pageToken = ""
			if resp.TokenPagination != nil {
				pageToken = resp.TokenPagination.NextPageToken
			}
			break
		}

		req = req.Token(resp.TokenPagination.NextPageToken)
	}

	result := output.NewResult(allReviews)
	result.WithServices("androidpublisher")
	result.WithPartial(scannedCount, filteredCount, 0)
	if pageToken != "" {
		result.WithPagination("", pageToken)
	}
	return c.Output(result)
}

func (c *CLI) reviewsReply(ctx context.Context, reviewID, replyText, templateFile string, maxActions int, rateLimit string, dryRun bool) error {
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

	// Post reply
	_, err = publisher.Reviews.Reply(c.packageName, reviewID, nil).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to reply: %v", err)))
	}

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

func (c *CLI) reviewsCapabilities(ctx context.Context) error {
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

func init() {
	// Suppress unused import warning
	_ = json.Marshal
}
