package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/api"
	"github.com/dl-alexandre/gpd/internal/auth"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
	"github.com/dl-alexandre/gpd/internal/storage"
)

// ReviewsCmd contains review management commands.
type ReviewsCmd struct {
	List           ReviewsListCmd           `cmd:"" help:"List user reviews"`
	Get            ReviewsGetCmd            `cmd:"" help:"Get a review by ID"`
	Reply          ReviewsReplyCmd          `cmd:"" help:"Reply to a review"`
	ResponseGet    ReviewsResponseGetCmd    `cmd:"" name:"response-get" help:"Get response for a review"`
	ResponseDelete ReviewsResponseDeleteCmd `cmd:"" name:"response-delete" help:"Delete response for a review"`
}

// ReviewsListCmd lists user reviews with optional filtering.
type ReviewsListCmd struct {
	MinRating           int    `help:"Minimum rating filter (1-5)"`
	MaxRating           int    `help:"Maximum rating filter (1-5)"`
	Language            string `help:"Filter by review language"`
	StartDate           string `help:"Start date (ISO 8601)"`
	EndDate             string `help:"End date (ISO 8601)"`
	ScanLimit           int    `help:"Maximum reviews to scan" default:"100"`
	IncludeReviewText   bool   `help:"Include review text in output"`
	TranslationLanguage string `help:"Language for translated reviews"`
	PageSize            int64  `help:"Results per page" default:"50"`
	PageToken           string `help:"Pagination token"`
	All                 bool   `help:"Fetch all pages"`
}

// reviewData represents a simplified review for output.
type reviewData struct {
	ReviewID     string    `json:"reviewId"`
	AuthorName   string    `json:"authorName"`
	Rating       int       `json:"rating"`
	Language     string    `json:"language"`
	ReviewText   string    `json:"reviewText,omitempty"`
	ReplyText    string    `json:"replyText,omitempty"`
	LastModified time.Time `json:"lastModified"`
	HasReply     bool      `json:"hasReply"`
	VersionCode  int64     `json:"versionCode,omitempty"`
	DeviceModel  string    `json:"deviceModel,omitempty"`
}

// reviewsListResponse represents the list response data.
type reviewsListResponse struct {
	Reviews       []reviewData `json:"reviews"`
	TotalCount    int          `json:"totalCount"`
	NextPageToken string       `json:"nextPageToken,omitempty"`
}

// createAPIClient creates a new API client with authentication.
func createAPIClient(ctx context.Context, globals *Globals) (*api.Client, error) {
	authMgr := auth.NewManager(storage.New())
	authMgr.SetStoreTokens(globals.StoreTokens)
	authMgr.SetActiveProfile(globals.Profile)

	creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
	if err != nil {
		return nil, err
	}

	return api.NewClient(ctx, creds.TokenSource, api.WithTimeout(globals.Timeout))
}

// reviewsPageResponse wraps the reviews list response to implement PageResponse.
type reviewsPageResponse struct {
	resp *androidpublisher.ReviewsListResponse
}

func (r reviewsPageResponse) GetNextPageToken() string {
	if r.resp.TokenPagination != nil {
		return r.resp.TokenPagination.NextPageToken
	}
	return ""
}

func (r reviewsPageResponse) GetItems() []*androidpublisher.Review {
	return r.resp.Reviews
}

// Run executes the reviews list command.
func (cmd *ReviewsListCmd) Run(globals *Globals) error {
	ctx := context.Background()

	if globals.Package == "" {
		return errors.ErrPackageRequired
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

	var allReviews []*androidpublisher.Review
	var nextPageToken string

	// First page query with retry
	err = client.DoWithRetry(ctx, func() error {
		call := svc.Reviews.List(globals.Package)
		if cmd.PageSize > 0 {
			call = call.MaxResults(int64(cmd.PageSize))
		}
		if cmd.TranslationLanguage != "" {
			call = call.TranslationLanguage(cmd.TranslationLanguage)
		}
		resp, err := call.Do()
		if err != nil {
			return err
		}

		allReviews = append(allReviews, resp.Reviews...)
		if resp.TokenPagination != nil {
			nextPageToken = resp.TokenPagination.NextPageToken
		}

		// Fetch additional pages if requested
		if cmd.All && nextPageToken != "" {
			query := func(pageToken string) (reviewsPageResponse, error) {
				pageCall := svc.Reviews.List(globals.Package).
					Token(pageToken)
				if cmd.PageSize > 0 {
					pageCall = pageCall.MaxResults(int64(cmd.PageSize))
				}
				if cmd.TranslationLanguage != "" {
					pageCall = pageCall.TranslationLanguage(cmd.TranslationLanguage)
				}
				pageResp, err := pageCall.Do()
				return reviewsPageResponse{resp: pageResp}, err
			}

			// Use scan limit as max results if set, otherwise unlimited
			maxResults := 0
			if cmd.ScanLimit > 0 {
				// Calculate remaining items needed
				maxResults = cmd.ScanLimit - len(allReviews)
				if maxResults <= 0 {
					return nil
				}
			}

			additionalReviews, remainingToken, err := fetchAllPages(ctx, query, nextPageToken, maxResults)
			if err != nil {
				return err
			}
			allReviews = append(allReviews, additionalReviews...)
			nextPageToken = remainingToken
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list reviews: %v", err))
	}

	// Apply filtering
	filtered := cmd.filterReviews(allReviews)

	// Convert to output format
	reviews := make([]reviewData, 0, len(filtered))
	for _, r := range filtered {
		reviews = append(reviews, cmd.convertReview(r))
	}

	data := reviewsListResponse{
		Reviews:       reviews,
		TotalCount:    len(reviews),
		NextPageToken: nextPageToken,
	}

	result := output.NewResult(data).
		WithServices("androidpublisher").
		WithPartial(len(allReviews), len(filtered), len(allReviews))

	if nextPageToken != "" {
		result = result.WithPagination(cmd.PageToken, nextPageToken)
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

// filterReviews applies rating, date, and language filters.
func (cmd *ReviewsListCmd) filterReviews(reviews []*androidpublisher.Review) []*androidpublisher.Review {
	if len(reviews) == 0 {
		return reviews
	}

	var filtered []*androidpublisher.Review
	for _, r := range reviews {
		if cmd.matchesFilters(r) {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

// matchesFilters checks if a review matches all specified filters.
func (cmd *ReviewsListCmd) matchesFilters(r *androidpublisher.Review) bool {
	// Get the user comment for filtering
	userComment := cmd.getUserComment(r)
	if userComment == nil {
		return false
	}

	// Rating filter
	if cmd.MinRating > 0 && userComment.StarRating < int64(cmd.MinRating) {
		return false
	}
	if cmd.MaxRating > 0 && userComment.StarRating > int64(cmd.MaxRating) {
		return false
	}

	// Language filter
	if cmd.Language != "" && !strings.EqualFold(userComment.ReviewerLanguage, cmd.Language) {
		return false
	}

	// Date range filter
	if cmd.StartDate != "" || cmd.EndDate != "" {
		lastModified := cmd.parseTimestamp(userComment.LastModified)
		if lastModified.IsZero() {
			return false
		}

		if cmd.StartDate != "" {
			start, err := time.Parse(time.RFC3339, cmd.StartDate)
			if err == nil && lastModified.Before(start) {
				return false
			}
		}

		if cmd.EndDate != "" {
			end, err := time.Parse(time.RFC3339, cmd.EndDate)
			if err == nil && lastModified.After(end) {
				return false
			}
		}
	}

	return true
}

// getUserComment extracts the user comment from a review.
func (cmd *ReviewsListCmd) getUserComment(r *androidpublisher.Review) *androidpublisher.UserComment {
	if r.Comments == nil {
		return nil
	}
	for _, c := range r.Comments {
		if c.UserComment != nil {
			return c.UserComment
		}
	}
	return nil
}

// parseTimestamp converts API timestamp to time.Time.
func (cmd *ReviewsListCmd) parseTimestamp(ts *androidpublisher.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return time.Unix(ts.Seconds, int64(ts.Nanos))
}

// convertReview converts API review to output format.
func (cmd *ReviewsListCmd) convertReview(r *androidpublisher.Review) reviewData {
	data := reviewData{
		ReviewID: r.ReviewId,
	}

	if r.AuthorName != "" {
		data.AuthorName = r.AuthorName
	}

	if r.Comments == nil {
		return data
	}

	for _, c := range r.Comments {
		if c.UserComment != nil {
			uc := c.UserComment
			data.Rating = int(uc.StarRating)
			data.Language = uc.ReviewerLanguage
			data.VersionCode = uc.AppVersionCode
			if uc.DeviceMetadata != nil {
				data.DeviceModel = uc.DeviceMetadata.ProductName
			}
			data.LastModified = cmd.parseTimestamp(uc.LastModified)

			if cmd.IncludeReviewText && uc.Text != "" {
				data.ReviewText = uc.Text
			}
		}

		if c.DeveloperComment != nil {
			data.HasReply = true
			if cmd.IncludeReviewText {
				data.ReplyText = c.DeveloperComment.Text
			}
		}
	}

	return data
}

// ReviewsGetCmd gets a single review by ID.
type ReviewsGetCmd struct {
	ReviewID            string `arg:"" optional:"" help:"Review ID to fetch"`
	IncludeReviewText   bool   `help:"Include review text in output"`
	TranslationLanguage string `help:"Language for translated review"`
}

// ReviewsReplyCmd replies to a review.
type ReviewsReplyCmd struct {
	ReviewID     string `arg:"" optional:"" help:"Review ID to reply to"`
	Text         string `help:"Reply text"`
	TemplateFile string `help:"Template file for reply" type:"existingfile"`
	MaxActions   int    `help:"Maximum replies per execution" default:"10"`
	RateLimit    string `help:"Rate limit between replies" default:"5s"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// ReviewsResponseGetCmd gets a review response.
type ReviewsResponseGetCmd struct {
	ReviewID string `arg:"" optional:"" help:"Review ID to fetch response for"`
}

// ReviewsResponseDeleteCmd deletes a review response.
type ReviewsResponseDeleteCmd struct {
	ReviewID string `arg:"" optional:"" help:"Review ID to delete response for"`
}

// Run executes the reviews get command.
func (cmd *ReviewsGetCmd) Run(globals *Globals) error {
	ctx := context.Background()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	if cmd.ReviewID == "" {
		return errors.NewAPIError(errors.CodeValidationError, "review ID is required")
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

	var review *androidpublisher.Review
	err = client.DoWithRetry(ctx, func() error {
		call := svc.Reviews.Get(globals.Package, cmd.ReviewID)
		if cmd.TranslationLanguage != "" {
			call = call.TranslationLanguage(cmd.TranslationLanguage)
		}
		var err error
		review, err = call.Do()
		return err
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get review: %v", err))
	}

	converter := &ReviewsListCmd{IncludeReviewText: cmd.IncludeReviewText}
	data := converter.convertReview(review)

	result := output.NewResult(data).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// replyResult represents a reply operation result.
type replyResult struct {
	ReviewID   string    `json:"reviewId"`
	ReplyText  string    `json:"replyText"`
	LastEdited time.Time `json:"lastEdited"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
}

// Run executes the reviews reply command.
func (cmd *ReviewsReplyCmd) Run(globals *Globals) error {
	ctx := context.Background()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	if cmd.ReviewID == "" {
		return errors.NewAPIError(errors.CodeValidationError, "review ID is required")
	}

	// Get reply text from file or direct input
	replyText, err := cmd.getReplyText()
	if err != nil {
		return err
	}
	if replyText == "" {
		return errors.NewAPIError(errors.CodeValidationError, "reply text is required")
	}

	// Validate reply text length (API limit is ~350 characters)
	if len(replyText) > 350 {
		return errors.NewAPIError(errors.CodeValidationError, "reply text exceeds 350 character limit").
			WithHint("Shorten your reply to be under 350 characters")
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"reviewId":  cmd.ReviewID,
			"replyText": replyText,
			"dryRun":    true,
		}).WithNoOp("dry run mode enabled")
		return outputResult(result, globals.Output, globals.Pretty)
	}

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service")
	}

	// Parse rate limit
	rateLimit, err := time.ParseDuration(cmd.RateLimit)
	if err != nil {
		rateLimit = 5 * time.Second
	}

	// Apply rate limiting
	if rateLimit > 0 {
		time.Sleep(rateLimit)
	}

	req := &androidpublisher.ReviewsReplyRequest{
		ReplyText: replyText,
	}

	var resp *androidpublisher.ReviewsReplyResponse
	err = client.DoWithRetry(ctx, func() error {
		var err error
		resp, err = svc.Reviews.Reply(globals.Package, cmd.ReviewID, req).Do()
		return err
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reply to review: %v", err))
	}

	result := replyResult{
		ReviewID:  cmd.ReviewID,
		ReplyText: replyText,
		Success:   true,
	}

	if resp.Result != nil && resp.Result.LastEdited != nil {
		result.LastEdited = time.Unix(resp.Result.LastEdited.Seconds, int64(resp.Result.LastEdited.Nanos))
	}

	return outputResult(output.NewResult(result).WithServices("androidpublisher"), globals.Output, globals.Pretty)
}

// getReplyText retrieves reply text from template file or direct input.
func (cmd *ReviewsReplyCmd) getReplyText() (string, error) {
	if cmd.TemplateFile != "" {
		content, err := os.ReadFile(cmd.TemplateFile)
		if err != nil {
			return "", errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read template file: %v", err))
		}
		return strings.TrimSpace(string(content)), nil
	}
	return cmd.Text, nil
}

// responseResult represents a review response.
type responseResult struct {
	ReviewID     string    `json:"reviewId"`
	HasResponse  bool      `json:"hasResponse"`
	ReplyText    string    `json:"replyText,omitempty"`
	LastModified time.Time `json:"lastModified,omitempty"`
}

// Run executes the reviews response get command.
func (cmd *ReviewsResponseGetCmd) Run(globals *Globals) error {
	ctx := context.Background()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	if cmd.ReviewID == "" {
		return errors.NewAPIError(errors.CodeValidationError, "review ID is required")
	}

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service")
	}

	var review *androidpublisher.Review
	err = client.DoWithRetry(ctx, func() error {
		var err error
		review, err = svc.Reviews.Get(globals.Package, cmd.ReviewID).Do()
		return err
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get review: %v", err))
	}

	result := responseResult{
		ReviewID: cmd.ReviewID,
	}

	// Check for developer comment
	if review.Comments != nil {
		for _, c := range review.Comments {
			if c.DeveloperComment != nil {
				result.HasResponse = true
				result.ReplyText = c.DeveloperComment.Text
				if c.DeveloperComment.LastModified != nil {
					result.LastModified = time.Unix(c.DeveloperComment.LastModified.Seconds, int64(c.DeveloperComment.LastModified.Nanos))
				}
				break
			}
		}
	}

	return outputResult(output.NewResult(result).WithServices("androidpublisher"), globals.Output, globals.Pretty)
}

// Run executes the reviews response delete command.
func (cmd *ReviewsResponseDeleteCmd) Run(globals *Globals) error {
	// Note: Google Play API doesn't support deleting review responses directly.
	// The workaround is to send an empty reply or update the reply to empty text.
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	if cmd.ReviewID == "" {
		return errors.NewAPIError(errors.CodeValidationError, "review ID is required")
	}

	return errors.NewAPIError(errors.CodeGeneralError, "delete response not supported by Google Play API").
		WithHint("The API does not support deleting responses. You can update a response to empty text using 'gpd reviews reply <review-id> --text \"\"'")
}
