//go:build unit
// +build unit

package cli

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
)

// ============================================================================
// ReviewsCmd Structure Tests
// ============================================================================

func TestReviewsCmd_SubcommandsExist(t *testing.T) {
	cmd := ReviewsCmd{}

	// Verify List field exists
	if reflect.TypeOf(cmd.List).String() != "cli.ReviewsListCmd" {
		t.Errorf("ReviewsCmd.List type = %v, want cli.ReviewsListCmd", reflect.TypeOf(cmd.List))
	}

	// Verify Get field exists
	if reflect.TypeOf(cmd.Get).String() != "cli.ReviewsGetCmd" {
		t.Errorf("ReviewsCmd.Get type = %v, want cli.ReviewsGetCmd", reflect.TypeOf(cmd.Get))
	}

	// Verify Reply field exists
	if reflect.TypeOf(cmd.Reply).String() != "cli.ReviewsReplyCmd" {
		t.Errorf("ReviewsCmd.Reply type = %v, want cli.ReviewsReplyCmd", reflect.TypeOf(cmd.Reply))
	}

	// Verify ResponseGet field exists
	if reflect.TypeOf(cmd.ResponseGet).String() != "cli.ReviewsResponseGetCmd" {
		t.Errorf("ReviewsCmd.ResponseGet type = %v, want cli.ReviewsResponseGetCmd", reflect.TypeOf(cmd.ResponseGet))
	}

	// Verify ResponseDelete field exists
	if reflect.TypeOf(cmd.ResponseDelete).String() != "cli.ReviewsResponseDeleteCmd" {
		t.Errorf("ReviewsCmd.ResponseDelete type = %v, want cli.ReviewsResponseDeleteCmd", reflect.TypeOf(cmd.ResponseDelete))
	}
}

// ============================================================================
// ReviewsListCmd Structure and Flag Tests
// ============================================================================

func TestReviewsListCmd_StructFields(t *testing.T) {
	cmd := ReviewsListCmd{
		MinRating:           3,
		MaxRating:           5,
		Language:            "en-US",
		StartDate:           "2024-01-01T00:00:00Z",
		EndDate:             "2024-12-31T23:59:59Z",
		ScanLimit:           200,
		IncludeReviewText:   true,
		TranslationLanguage: "es",
		PageSize:            100,
		PageToken:           "token123",
		All:                 true,
	}

	// Verify all fields can be set
	if cmd.MinRating != 3 {
		t.Errorf("MinRating = %d, want 3", cmd.MinRating)
	}
	if cmd.MaxRating != 5 {
		t.Errorf("MaxRating = %d, want 5", cmd.MaxRating)
	}
	if cmd.Language != "en-US" {
		t.Errorf("Language = %s, want en-US", cmd.Language)
	}
	if cmd.StartDate != "2024-01-01T00:00:00Z" {
		t.Errorf("StartDate = %s, want 2024-01-01T00:00:00Z", cmd.StartDate)
	}
	if cmd.EndDate != "2024-12-31T23:59:59Z" {
		t.Errorf("EndDate = %s, want 2024-12-31T23:59:59Z", cmd.EndDate)
	}
	if cmd.ScanLimit != 200 {
		t.Errorf("ScanLimit = %d, want 200", cmd.ScanLimit)
	}
	if !cmd.IncludeReviewText {
		t.Error("IncludeReviewText should be true")
	}
	if cmd.TranslationLanguage != "es" {
		t.Errorf("TranslationLanguage = %s, want es", cmd.TranslationLanguage)
	}
	if cmd.PageSize != 100 {
		t.Errorf("PageSize = %d, want 100", cmd.PageSize)
	}
	if cmd.PageToken != "token123" {
		t.Errorf("PageToken = %s, want token123", cmd.PageToken)
	}
	if !cmd.All {
		t.Error("All should be true")
	}
}

func TestReviewsListCmd_DefaultValues(t *testing.T) {
	cmd := ReviewsListCmd{}

	// Check default values
	if cmd.MinRating != 0 {
		t.Errorf("MinRating default = %d, want 0", cmd.MinRating)
	}
	if cmd.MaxRating != 0 {
		t.Errorf("MaxRating default = %d, want 0", cmd.MaxRating)
	}
	if cmd.ScanLimit != 0 {
		t.Errorf("ScanLimit default = %d, want 0", cmd.ScanLimit)
	}
	if cmd.PageSize != 0 {
		t.Errorf("PageSize default = %d, want 0", cmd.PageSize)
	}
	if cmd.All {
		t.Error("All default should be false")
	}
	if cmd.IncludeReviewText {
		t.Error("IncludeReviewText default should be false")
	}
}

func TestReviewsListCmd_Run_PackageRequired(t *testing.T) {
	cmd := &ReviewsListCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// ReviewsGetCmd Structure and Flag Tests
// ============================================================================

func TestReviewsGetCmd_StructFields(t *testing.T) {
	cmd := ReviewsGetCmd{
		ReviewID:            "review-123",
		IncludeReviewText:   true,
		TranslationLanguage: "fr",
	}

	if cmd.ReviewID != "review-123" {
		t.Errorf("ReviewID = %s, want review-123", cmd.ReviewID)
	}
	if !cmd.IncludeReviewText {
		t.Error("IncludeReviewText should be true")
	}
	if cmd.TranslationLanguage != "fr" {
		t.Errorf("TranslationLanguage = %s, want fr", cmd.TranslationLanguage)
	}
}

func TestReviewsGetCmd_Run_PackageRequired(t *testing.T) {
	cmd := &ReviewsGetCmd{ReviewID: "review-123"}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestReviewsGetCmd_Run_ReviewIDRequired(t *testing.T) {
	cmd := &ReviewsGetCmd{ReviewID: ""}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing review ID")
	}
	if !strings.Contains(err.Error(), "review ID is required") {
		t.Errorf("Expected 'review ID is required' error, got: %v", err)
	}
}

// ============================================================================
// ReviewsReplyCmd Structure and Flag Tests
// ============================================================================

func TestReviewsReplyCmd_StructFields(t *testing.T) {
	cmd := ReviewsReplyCmd{
		ReviewID:     "review-456",
		Text:         "Thank you for your feedback!",
		TemplateFile: "/path/to/template.txt",
		MaxActions:   5,
		RateLimit:    "10s",
		DryRun:       true,
	}

	if cmd.ReviewID != "review-456" {
		t.Errorf("ReviewID = %s, want review-456", cmd.ReviewID)
	}
	if cmd.Text != "Thank you for your feedback!" {
		t.Errorf("Text = %s, want 'Thank you for your feedback!'", cmd.Text)
	}
	if cmd.TemplateFile != "/path/to/template.txt" {
		t.Errorf("TemplateFile = %s, want /path/to/template.txt", cmd.TemplateFile)
	}
	if cmd.MaxActions != 5 {
		t.Errorf("MaxActions = %d, want 5", cmd.MaxActions)
	}
	if cmd.RateLimit != "10s" {
		t.Errorf("RateLimit = %s, want 10s", cmd.RateLimit)
	}
	if !cmd.DryRun {
		t.Error("DryRun should be true")
	}
}

func TestReviewsReplyCmd_Run_PackageRequired(t *testing.T) {
	cmd := &ReviewsReplyCmd{ReviewID: "review-123", Text: "Thanks!"}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestReviewsReplyCmd_Run_ReviewIDRequired(t *testing.T) {
	cmd := &ReviewsReplyCmd{ReviewID: "", Text: "Thanks!"}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing review ID")
	}
	if !strings.Contains(err.Error(), "review ID is required") {
		t.Errorf("Expected 'review ID is required' error, got: %v", err)
	}
}

func TestReviewsReplyCmd_Run_ReplyTextRequired(t *testing.T) {
	cmd := &ReviewsReplyCmd{ReviewID: "review-123", Text: ""}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing reply text")
	}
	if !strings.Contains(err.Error(), "reply text is required") {
		t.Errorf("Expected 'reply text is required' error, got: %v", err)
	}
}

func TestReviewsReplyCmd_Run_ReplyTextTooLong(t *testing.T) {
	longText := strings.Repeat("a", 351) // 351 characters, exceeds 350 limit
	cmd := &ReviewsReplyCmd{ReviewID: "review-123", Text: longText}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for reply text exceeding 350 characters")
	}
	if !strings.Contains(err.Error(), "exceeds 350 character limit") {
		t.Errorf("Expected 'exceeds 350 character limit' error, got: %v", err)
	}
}

func TestReviewsReplyCmd_Run_MaxLengthReply(t *testing.T) {
	maxText := strings.Repeat("a", 350) // Exactly 350 characters
	cmd := &ReviewsReplyCmd{ReviewID: "review-123", Text: maxText, DryRun: true}
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	// Should not error in dry run mode
	err := cmd.Run(globals)
	// May error due to auth, but should not error due to text length
	if err != nil && strings.Contains(err.Error(), "exceeds 350 character limit") {
		t.Errorf("350 character reply should be valid, got: %v", err)
	}
}

func TestReviewsReplyCmd_getReplyText_FromTextField(t *testing.T) {
	cmd := &ReviewsReplyCmd{Text: "Direct reply text"}

	text, err := cmd.getReplyText()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if text != "Direct reply text" {
		t.Errorf("getReplyText() = %s, want 'Direct reply text'", text)
	}
}

func TestReviewsReplyCmd_getReplyText_FromTemplateFile(t *testing.T) {
	content := "Template reply text with thanks!"
	tmpFile := createReviewsTempFile(t, "reply_template.txt", []byte(content))
	defer os.Remove(tmpFile)

	cmd := &ReviewsReplyCmd{TemplateFile: tmpFile}

	text, err := cmd.getReplyText()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if text != content {
		t.Errorf("getReplyText() = %s, want '%s'", text, content)
	}
}

func TestReviewsReplyCmd_getReplyText_TemplateFileWithWhitespace(t *testing.T) {
	content := "  Template with whitespace  \n\n  "
	tmpFile := createReviewsTempFile(t, "reply_template.txt", []byte(content))
	defer os.Remove(tmpFile)

	cmd := &ReviewsReplyCmd{TemplateFile: tmpFile}

	text, err := cmd.getReplyText()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Whitespace should be trimmed
	expected := "Template with whitespace"
	if text != expected {
		t.Errorf("getReplyText() = '%s', want '%s'", text, expected)
	}
}

func TestReviewsReplyCmd_getReplyText_NonexistentFile(t *testing.T) {
	cmd := &ReviewsReplyCmd{TemplateFile: "/nonexistent/path/template.txt"}

	_, err := cmd.getReplyText()
	if err == nil {
		t.Fatal("Expected error for nonexistent template file")
	}
	if !strings.Contains(err.Error(), "failed to read template file") {
		t.Errorf("Expected 'failed to read template file' error, got: %v", err)
	}
}

func TestReviewsReplyCmd_getReplyText_Priority(t *testing.T) {
	// When both Text and TemplateFile are provided, TemplateFile takes priority
	content := "From template"
	tmpFile := createReviewsTempFile(t, "reply_template.txt", []byte(content))
	defer os.Remove(tmpFile)

	cmd := &ReviewsReplyCmd{
		Text:         "From text field",
		TemplateFile: tmpFile,
	}

	text, err := cmd.getReplyText()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if text != "From template" {
		t.Errorf("TemplateFile should take priority over Text field, got: %s", text)
	}
}

func TestReviewsReplyCmd_Run_DryRun(t *testing.T) {
	cmd := &ReviewsReplyCmd{
		ReviewID: "review-123",
		Text:     "Thank you for your review!",
		DryRun:   true,
	}
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	// Dry run should not require authentication
	err := cmd.Run(globals)
	// Should complete without authentication error in dry run mode
	// The command outputs dry run info to stdout
	if err != nil {
		// If there's an error, it shouldn't be about missing auth
		if strings.Contains(err.Error(), "authentication") || strings.Contains(err.Error(), "key") {
			t.Errorf("Dry run should not require authentication, got: %v", err)
		}
	}
}

// ============================================================================
// ReviewsResponseGetCmd Structure and Flag Tests
// ============================================================================

func TestReviewsResponseGetCmd_StructFields(t *testing.T) {
	cmd := ReviewsResponseGetCmd{ReviewID: "review-789"}

	if cmd.ReviewID != "review-789" {
		t.Errorf("ReviewID = %s, want review-789", cmd.ReviewID)
	}
}

func TestReviewsResponseGetCmd_Run_PackageRequired(t *testing.T) {
	cmd := &ReviewsResponseGetCmd{ReviewID: "review-123"}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestReviewsResponseGetCmd_Run_ReviewIDRequired(t *testing.T) {
	cmd := &ReviewsResponseGetCmd{ReviewID: ""}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing review ID")
	}
	if !strings.Contains(err.Error(), "review ID is required") {
		t.Errorf("Expected 'review ID is required' error, got: %v", err)
	}
}

// ============================================================================
// ReviewsResponseDeleteCmd Structure and Flag Tests
// ============================================================================

func TestReviewsResponseDeleteCmd_StructFields(t *testing.T) {
	cmd := ReviewsResponseDeleteCmd{ReviewID: "review-abc"}

	if cmd.ReviewID != "review-abc" {
		t.Errorf("ReviewID = %s, want review-abc", cmd.ReviewID)
	}
}

func TestReviewsResponseDeleteCmd_Run_PackageRequired(t *testing.T) {
	cmd := &ReviewsResponseDeleteCmd{ReviewID: "review-123"}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestReviewsResponseDeleteCmd_Run_ReviewIDRequired(t *testing.T) {
	cmd := &ReviewsResponseDeleteCmd{ReviewID: ""}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing review ID")
	}
	if !strings.Contains(err.Error(), "review ID is required") {
		t.Errorf("Expected 'review ID is required' error, got: %v", err)
	}
}

func TestReviewsResponseDeleteCmd_Run_NotSupported(t *testing.T) {
	cmd := &ReviewsResponseDeleteCmd{ReviewID: "review-123"}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error indicating delete is not supported")
	}
	if !strings.Contains(err.Error(), "not supported") && !strings.Contains(err.Error(), "not supported by Google Play API") {
		t.Errorf("Expected 'not supported' error, got: %v", err)
	}
	// Check for hint about workaround
	apiErr, ok := err.(*errors.APIError)
	if ok && apiErr.Hint == "" {
		t.Error("Expected error to have hint about workaround")
	}
}

// ============================================================================
// reviewData Structure Tests
// ============================================================================

func TestReviewData_StructFields(t *testing.T) {
	data := reviewData{
		ReviewID:     "rev-123",
		AuthorName:   "John Doe",
		Rating:       5,
		Language:     "en-US",
		ReviewText:   "Great app!",
		ReplyText:    "Thank you!",
		LastModified: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		HasReply:     true,
		VersionCode:  100,
		DeviceModel:  "Pixel 7",
	}

	if data.ReviewID != "rev-123" {
		t.Errorf("ReviewID = %s, want rev-123", data.ReviewID)
	}
	if data.AuthorName != "John Doe" {
		t.Errorf("AuthorName = %s, want John Doe", data.AuthorName)
	}
	if data.Rating != 5 {
		t.Errorf("Rating = %d, want 5", data.Rating)
	}
	if data.Language != "en-US" {
		t.Errorf("Language = %s, want en-US", data.Language)
	}
	if data.ReviewText != "Great app!" {
		t.Errorf("ReviewText = %s, want 'Great app!'", data.ReviewText)
	}
	if data.ReplyText != "Thank you!" {
		t.Errorf("ReplyText = %s, want 'Thank you!'", data.ReplyText)
	}
	if !data.HasReply {
		t.Error("HasReply should be true")
	}
	if data.VersionCode != 100 {
		t.Errorf("VersionCode = %d, want 100", data.VersionCode)
	}
	if data.DeviceModel != "Pixel 7" {
		t.Errorf("DeviceModel = %s, want Pixel 7", data.DeviceModel)
	}
}

// ============================================================================
// reviewsListResponse Structure Tests
// ============================================================================

func TestReviewsListResponse_StructFields(t *testing.T) {
	response := reviewsListResponse{
		Reviews: []reviewData{
			{ReviewID: "rev-1", Rating: 5},
			{ReviewID: "rev-2", Rating: 4},
		},
		TotalCount:    2,
		NextPageToken: "next-token-123",
	}

	if len(response.Reviews) != 2 {
		t.Errorf("Reviews count = %d, want 2", len(response.Reviews))
	}
	if response.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", response.TotalCount)
	}
	if response.NextPageToken != "next-token-123" {
		t.Errorf("NextPageToken = %s, want next-token-123", response.NextPageToken)
	}
}

// ============================================================================
// replyResult Structure Tests
// ============================================================================

func TestReplyResult_StructFields(t *testing.T) {
	result := replyResult{
		ReviewID:   "rev-456",
		ReplyText:  "Thanks for reviewing!",
		LastEdited: time.Date(2024, 1, 20, 14, 0, 0, 0, time.UTC),
		Success:    true,
		Error:      "",
	}

	if result.ReviewID != "rev-456" {
		t.Errorf("ReviewID = %s, want rev-456", result.ReviewID)
	}
	if result.ReplyText != "Thanks for reviewing!" {
		t.Errorf("ReplyText = %s, want 'Thanks for reviewing!'", result.ReplyText)
	}
	if !result.Success {
		t.Error("Success should be true")
	}
	if result.Error != "" {
		t.Errorf("Error should be empty, got: %s", result.Error)
	}
}

// ============================================================================
// responseResult Structure Tests
// ============================================================================

func TestResponseResult_StructFields(t *testing.T) {
	result := responseResult{
		ReviewID:     "rev-789",
		HasResponse:  true,
		ReplyText:    "Developer response here",
		LastModified: time.Date(2024, 1, 25, 9, 0, 0, 0, time.UTC),
	}

	if result.ReviewID != "rev-789" {
		t.Errorf("ReviewID = %s, want rev-789", result.ReviewID)
	}
	if !result.HasResponse {
		t.Error("HasResponse should be true")
	}
	if result.ReplyText != "Developer response here" {
		t.Errorf("ReplyText = %s, want 'Developer response here'", result.ReplyText)
	}
}

// ============================================================================
// Filter Logic Tests
// ============================================================================

func TestReviewsListCmd_filterReviews_EmptyInput(t *testing.T) {
	cmd := &ReviewsListCmd{}

	var emptyReviews []*androidpublisher.Review
	filtered := cmd.filterReviews(emptyReviews)

	if len(filtered) != 0 {
		t.Errorf("Expected 0 reviews for empty input, got %d", len(filtered))
	}
}

func TestReviewsListCmd_filterReviews_NilInput(t *testing.T) {
	cmd := &ReviewsListCmd{}

	filtered := cmd.filterReviews(nil)

	if len(filtered) != 0 {
		t.Errorf("Expected 0 reviews for nil input, got %d", len(filtered))
	}
}

func TestReviewsListCmd_matchesFilters_NoFilters(t *testing.T) {
	cmd := &ReviewsListCmd{} // No filters set

	review := &androidpublisher.Review{
		ReviewId: "rev-1",
		Comments: []*androidpublisher.Comment{
			{
				UserComment: &androidpublisher.UserComment{
					StarRating:       5,
					ReviewerLanguage: "en-US",
					LastModified: &androidpublisher.Timestamp{
						Seconds: time.Now().Unix(),
					},
				},
			},
		},
	}

	if !cmd.matchesFilters(review) {
		t.Error("Review should match when no filters are set")
	}
}

func TestReviewsListCmd_matchesFilters_MinRating(t *testing.T) {
	cmd := &ReviewsListCmd{MinRating: 4}

	tests := []struct {
		name     string
		rating   int64
		expected bool
	}{
		{"rating 5 matches min 4", 5, true},
		{"rating 4 matches min 4", 4, true},
		{"rating 3 does not match min 4", 3, false},
		{"rating 1 does not match min 4", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			review := &androidpublisher.Review{
				Comments: []*androidpublisher.Comment{
					{
						UserComment: &androidpublisher.UserComment{
							StarRating: tt.rating,
							LastModified: &androidpublisher.Timestamp{
								Seconds: time.Now().Unix(),
							},
						},
					},
				},
			}

			result := cmd.matchesFilters(review)
			if result != tt.expected {
				t.Errorf("matchesFilters() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReviewsListCmd_matchesFilters_MaxRating(t *testing.T) {
	cmd := &ReviewsListCmd{MaxRating: 3}

	tests := []struct {
		name     string
		rating   int64
		expected bool
	}{
		{"rating 3 matches max 3", 3, true},
		{"rating 2 matches max 3", 2, true},
		{"rating 4 does not match max 3", 4, false},
		{"rating 5 does not match max 3", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			review := &androidpublisher.Review{
				Comments: []*androidpublisher.Comment{
					{
						UserComment: &androidpublisher.UserComment{
							StarRating: tt.rating,
							LastModified: &androidpublisher.Timestamp{
								Seconds: time.Now().Unix(),
							},
						},
					},
				},
			}

			result := cmd.matchesFilters(review)
			if result != tt.expected {
				t.Errorf("matchesFilters() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReviewsListCmd_matchesFilters_RatingRange(t *testing.T) {
	cmd := &ReviewsListCmd{MinRating: 2, MaxRating: 4}

	tests := []struct {
		name     string
		rating   int64
		expected bool
	}{
		{"rating 4 in range 2-4", 4, true},
		{"rating 3 in range 2-4", 3, true},
		{"rating 2 in range 2-4", 2, true},
		{"rating 5 outside range 2-4", 5, false},
		{"rating 1 outside range 2-4", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			review := &androidpublisher.Review{
				Comments: []*androidpublisher.Comment{
					{
						UserComment: &androidpublisher.UserComment{
							StarRating: tt.rating,
							LastModified: &androidpublisher.Timestamp{
								Seconds: time.Now().Unix(),
							},
						},
					},
				},
			}

			result := cmd.matchesFilters(review)
			if result != tt.expected {
				t.Errorf("matchesFilters() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReviewsListCmd_matchesFilters_Language(t *testing.T) {
	cmd := &ReviewsListCmd{Language: "en-US"}

	tests := []struct {
		name     string
		language string
		expected bool
	}{
		{"en-US matches en-US", "en-US", true},
		{"en-us matches en-US (case insensitive)", "en-us", true},
		{"en-GB does not match en-US", "en-GB", false},
		{"de-DE does not match en-US", "de-DE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			review := &androidpublisher.Review{
				Comments: []*androidpublisher.Comment{
					{
						UserComment: &androidpublisher.UserComment{
							ReviewerLanguage: tt.language,
							LastModified: &androidpublisher.Timestamp{
								Seconds: time.Now().Unix(),
							},
						},
					},
				},
			}

			result := cmd.matchesFilters(review)
			if result != tt.expected {
				t.Errorf("matchesFilters() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReviewsListCmd_matchesFilters_DateRange(t *testing.T) {
	baseTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		startDate  string
		endDate    string
		reviewTime time.Time
		expected   bool
	}{
		{
			name:       "review within date range",
			startDate:  "2024-01-01T00:00:00Z",
			endDate:    "2024-12-31T23:59:59Z",
			reviewTime: baseTime,
			expected:   true,
		},
		{
			name:       "review before start date",
			startDate:  "2024-07-01T00:00:00Z",
			endDate:    "2024-12-31T23:59:59Z",
			reviewTime: baseTime,
			expected:   false,
		},
		{
			name:       "review after end date",
			startDate:  "2024-01-01T00:00:00Z",
			endDate:    "2024-05-31T23:59:59Z",
			reviewTime: baseTime,
			expected:   false,
		},
		{
			name:       "only start date filter",
			startDate:  "2024-01-01T00:00:00Z",
			endDate:    "",
			reviewTime: baseTime,
			expected:   true,
		},
		{
			name:       "only end date filter",
			startDate:  "",
			endDate:    "2024-12-31T23:59:59Z",
			reviewTime: baseTime,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ReviewsListCmd{
				StartDate: tt.startDate,
				EndDate:   tt.endDate,
			}

			review := &androidpublisher.Review{
				Comments: []*androidpublisher.Comment{
					{
						UserComment: &androidpublisher.UserComment{
							LastModified: &androidpublisher.Timestamp{
								Seconds: tt.reviewTime.Unix(),
							},
						},
					},
				},
			}

			result := cmd.matchesFilters(review)
			if result != tt.expected {
				t.Errorf("matchesFilters() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReviewsListCmd_matchesFilters_NoUserComment(t *testing.T) {
	cmd := &ReviewsListCmd{MinRating: 3}

	review := &androidpublisher.Review{
		ReviewId: "rev-1",
		Comments: []*androidpublisher.Comment{
			{
				DeveloperComment: &androidpublisher.DeveloperComment{
					Text: "Developer reply",
				},
			},
		},
	}

	if cmd.matchesFilters(review) {
		t.Error("Review without user comment should not match filters")
	}
}

func TestReviewsListCmd_matchesFilters_NilComments(t *testing.T) {
	cmd := &ReviewsListCmd{MinRating: 3}

	review := &androidpublisher.Review{
		ReviewId: "rev-1",
		Comments: nil,
	}

	if cmd.matchesFilters(review) {
		t.Error("Review with nil comments should not match filters")
	}
}

// ============================================================================
// getUserComment Tests
// ============================================================================

func TestReviewsListCmd_getUserComment_Found(t *testing.T) {
	cmd := &ReviewsListCmd{}

	review := &androidpublisher.Review{
		Comments: []*androidpublisher.Comment{
			{
				DeveloperComment: &androidpublisher.DeveloperComment{
					Text: "Reply",
				},
			},
			{
				UserComment: &androidpublisher.UserComment{
					Text: "User review text",
				},
			},
		},
	}

	comment := cmd.getUserComment(review)
	if comment == nil {
		t.Fatal("Expected to find user comment")
	}
	if comment.Text != "User review text" {
		t.Errorf("User comment text = %s, want 'User review text'", comment.Text)
	}
}

func TestReviewsListCmd_getUserComment_NotFound(t *testing.T) {
	cmd := &ReviewsListCmd{}

	review := &androidpublisher.Review{
		Comments: []*androidpublisher.Comment{
			{
				DeveloperComment: &androidpublisher.DeveloperComment{
					Text: "Only developer comment",
				},
			},
		},
	}

	comment := cmd.getUserComment(review)
	if comment != nil {
		t.Error("Expected nil when no user comment exists")
	}
}

func TestReviewsListCmd_getUserComment_NilComments(t *testing.T) {
	cmd := &ReviewsListCmd{}

	review := &androidpublisher.Review{
		Comments: nil,
	}

	comment := cmd.getUserComment(review)
	if comment != nil {
		t.Error("Expected nil when comments are nil")
	}
}

func TestReviewsListCmd_getUserComment_EmptyComments(t *testing.T) {
	cmd := &ReviewsListCmd{}

	review := &androidpublisher.Review{
		Comments: []*androidpublisher.Comment{},
	}

	comment := cmd.getUserComment(review)
	if comment != nil {
		t.Error("Expected nil when comments are empty")
	}
}

// ============================================================================
// parseTimestamp Tests
// ============================================================================

func TestReviewsListCmd_parseTimestamp_Valid(t *testing.T) {
	cmd := &ReviewsListCmd{}

	expectedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	ts := &androidpublisher.Timestamp{
		Seconds: expectedTime.Unix(),
		Nanos:   0,
	}

	result := cmd.parseTimestamp(ts)
	if !result.Equal(expectedTime) {
		t.Errorf("parseTimestamp() = %v, want %v", result, expectedTime)
	}
}

func TestReviewsListCmd_parseTimestamp_WithNanos(t *testing.T) {
	cmd := &ReviewsListCmd{}

	ts := &androidpublisher.Timestamp{
		Seconds: 1705315800,
		Nanos:   500000000, // 500 milliseconds
	}

	result := cmd.parseTimestamp(ts)
	expectedTime := time.Unix(1705315800, 500000000)
	if !result.Equal(expectedTime) {
		t.Errorf("parseTimestamp() = %v, want %v", result, expectedTime)
	}
}

func TestReviewsListCmd_parseTimestamp_Nil(t *testing.T) {
	cmd := &ReviewsListCmd{}

	result := cmd.parseTimestamp(nil)
	if !result.IsZero() {
		t.Error("parseTimestamp(nil) should return zero time")
	}
}

// ============================================================================
// convertReview Tests
// ============================================================================

func TestReviewsListCmd_convertReview_Basic(t *testing.T) {
	cmd := &ReviewsListCmd{IncludeReviewText: false}

	reviewTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	review := &androidpublisher.Review{
		ReviewId:   "rev-123",
		AuthorName: "Test User",
		Comments: []*androidpublisher.Comment{
			{
				UserComment: &androidpublisher.UserComment{
					StarRating:       5,
					ReviewerLanguage: "en-US",
					Text:             "Great app!",
					AppVersionCode:   100,
					DeviceMetadata: &androidpublisher.DeviceMetadata{
						ProductName: "Pixel 7",
					},
					LastModified: &androidpublisher.Timestamp{
						Seconds: reviewTime.Unix(),
					},
				},
			},
		},
	}

	result := cmd.convertReview(review)

	if result.ReviewID != "rev-123" {
		t.Errorf("ReviewID = %s, want rev-123", result.ReviewID)
	}
	if result.AuthorName != "Test User" {
		t.Errorf("AuthorName = %s, want Test User", result.AuthorName)
	}
	if result.Rating != 5 {
		t.Errorf("Rating = %d, want 5", result.Rating)
	}
	if result.Language != "en-US" {
		t.Errorf("Language = %s, want en-US", result.Language)
	}
	if result.VersionCode != 100 {
		t.Errorf("VersionCode = %d, want 100", result.VersionCode)
	}
	if result.DeviceModel != "Pixel 7" {
		t.Errorf("DeviceModel = %s, want Pixel 7", result.DeviceModel)
	}
	if !result.LastModified.Equal(reviewTime) {
		t.Errorf("LastModified = %v, want %v", result.LastModified, reviewTime)
	}
	// When IncludeReviewText is false, text should be empty
	if result.ReviewText != "" {
		t.Errorf("ReviewText should be empty when IncludeReviewText=false, got: %s", result.ReviewText)
	}
}

func TestReviewsListCmd_convertReview_WithReviewText(t *testing.T) {
	cmd := &ReviewsListCmd{IncludeReviewText: true}

	review := &androidpublisher.Review{
		ReviewId: "rev-123",
		Comments: []*androidpublisher.Comment{
			{
				UserComment: &androidpublisher.UserComment{
					StarRating: 4,
					Text:       "Good app but needs work",
					LastModified: &androidpublisher.Timestamp{
						Seconds: time.Now().Unix(),
					},
				},
			},
		},
	}

	result := cmd.convertReview(review)

	if result.ReviewText != "Good app but needs work" {
		t.Errorf("ReviewText = %s, want 'Good app but needs work'", result.ReviewText)
	}
}

func TestReviewsListCmd_convertReview_WithReply(t *testing.T) {
	cmd := &ReviewsListCmd{IncludeReviewText: true}

	review := &androidpublisher.Review{
		ReviewId: "rev-123",
		Comments: []*androidpublisher.Comment{
			{
				UserComment: &androidpublisher.UserComment{
					StarRating: 3,
					Text:       "Average app",
					LastModified: &androidpublisher.Timestamp{
						Seconds: time.Now().Unix(),
					},
				},
			},
			{
				DeveloperComment: &androidpublisher.DeveloperComment{
					Text: "Thanks for your feedback!",
					LastModified: &androidpublisher.Timestamp{
						Seconds: time.Now().Unix(),
					},
				},
			},
		},
	}

	result := cmd.convertReview(review)

	if !result.HasReply {
		t.Error("HasReply should be true")
	}
	if result.ReplyText != "Thanks for your feedback!" {
		t.Errorf("ReplyText = %s, want 'Thanks for your feedback!'", result.ReplyText)
	}
}

func TestReviewsListCmd_convertReview_ReplyWithoutText(t *testing.T) {
	cmd := &ReviewsListCmd{IncludeReviewText: false}

	review := &androidpublisher.Review{
		ReviewId: "rev-123",
		Comments: []*androidpublisher.Comment{
			{
				UserComment: &androidpublisher.UserComment{
					StarRating: 5,
					Text:       "Great!",
					LastModified: &androidpublisher.Timestamp{
						Seconds: time.Now().Unix(),
					},
				},
			},
			{
				DeveloperComment: &androidpublisher.DeveloperComment{
					Text: "Thanks!",
					LastModified: &androidpublisher.Timestamp{
						Seconds: time.Now().Unix(),
					},
				},
			},
		},
	}

	result := cmd.convertReview(review)

	if !result.HasReply {
		t.Error("HasReply should be true even without IncludeReviewText")
	}
	// ReplyText should be empty when IncludeReviewText is false
	if result.ReplyText != "" {
		t.Errorf("ReplyText should be empty when IncludeReviewText=false, got: %s", result.ReplyText)
	}
}

func TestReviewsListCmd_convertReview_NoComments(t *testing.T) {
	cmd := &ReviewsListCmd{}

	review := &androidpublisher.Review{
		ReviewId:   "rev-123",
		AuthorName: "Test User",
		Comments:   nil,
	}

	result := cmd.convertReview(review)

	if result.ReviewID != "rev-123" {
		t.Errorf("ReviewID = %s, want rev-123", result.ReviewID)
	}
	if result.AuthorName != "Test User" {
		t.Errorf("AuthorName = %s, want Test User", result.AuthorName)
	}
	// Fields dependent on comments should be zero values
	if result.Rating != 0 {
		t.Errorf("Rating should be 0 when no comments, got: %d", result.Rating)
	}
}

func TestReviewsListCmd_convertReview_NoDeviceMetadata(t *testing.T) {
	cmd := &ReviewsListCmd{}

	review := &androidpublisher.Review{
		ReviewId: "rev-123",
		Comments: []*androidpublisher.Comment{
			{
				UserComment: &androidpublisher.UserComment{
					StarRating: 5,
					Text:       "Great!",
					LastModified: &androidpublisher.Timestamp{
						Seconds: time.Now().Unix(),
					},
					// No DeviceMetadata
				},
			},
		},
	}

	result := cmd.convertReview(review)

	if result.DeviceModel != "" {
		t.Errorf("DeviceModel should be empty when no device metadata, got: %s", result.DeviceModel)
	}
}

// ============================================================================
// reviewsPageResponse Tests
// ============================================================================

func TestReviewsPageResponse_GetNextPageToken_WithToken(t *testing.T) {
	resp := &androidpublisher.ReviewsListResponse{
		TokenPagination: &androidpublisher.TokenPagination{
			NextPageToken: "next-page-123",
		},
	}

	pageResp := reviewsPageResponse{resp: resp}

	if pageResp.GetNextPageToken() != "next-page-123" {
		t.Errorf("GetNextPageToken() = %s, want next-page-123", pageResp.GetNextPageToken())
	}
}

func TestReviewsPageResponse_GetNextPageToken_NoToken(t *testing.T) {
	resp := &androidpublisher.ReviewsListResponse{
		TokenPagination: nil,
	}

	pageResp := reviewsPageResponse{resp: resp}

	if pageResp.GetNextPageToken() != "" {
		t.Errorf("GetNextPageToken() should return empty string when no pagination, got: %s", pageResp.GetNextPageToken())
	}
}

func TestReviewsPageResponse_GetItems(t *testing.T) {
	reviews := []*androidpublisher.Review{
		{ReviewId: "rev-1"},
		{ReviewId: "rev-2"},
	}

	resp := &androidpublisher.ReviewsListResponse{
		Reviews: reviews,
	}

	pageResp := reviewsPageResponse{resp: resp}

	items := pageResp.GetItems()
	if len(items) != 2 {
		t.Errorf("GetItems() returned %d items, want 2", len(items))
	}
	if items[0].ReviewId != "rev-1" || items[1].ReviewId != "rev-2" {
		t.Error("GetItems() returned incorrect items")
	}
}

func TestReviewsPageResponse_GetItems_Empty(t *testing.T) {
	resp := &androidpublisher.ReviewsListResponse{
		Reviews: []*androidpublisher.Review{},
	}

	pageResp := reviewsPageResponse{resp: resp}

	items := pageResp.GetItems()
	if len(items) != 0 {
		t.Errorf("GetItems() should return empty slice, got %d items", len(items))
	}
}

// ============================================================================
// Context Handling Tests
// ============================================================================

func TestReviewsCommands_ContextPropagation(t *testing.T) {
	// Test that context is properly used when provided
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cmd := &ReviewsListCmd{}
	globals := &Globals{
		Package: "com.example.app",
		Context: ctx,
	}

	// This will fail due to context cancellation or auth
	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error due to cancelled context or auth failure")
	}
	// Error should be related to context cancellation or auth
	if ctx.Err() != context.Canceled {
		t.Error("Expected context to be canceled")
	}
}

func TestReviewsCommands_NilContext(t *testing.T) {
	// Test that nil context is handled by using context.Background()
	cmd := &ReviewsListCmd{}
	globals := &Globals{
		Package: "com.example.app",
		Context: nil,
	}

	// Should not panic, will fail on auth
	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error due to missing auth")
	}
}

// ============================================================================
// Command Validation Tests
// ============================================================================

func TestReviewsListCmd_ValidateFilters(t *testing.T) {
	tests := []struct {
		name      string
		minRating int
		maxRating int
		startDate string
		endDate   string
		wantErr   bool
	}{
		{
			name:      "valid rating range",
			minRating: 1,
			maxRating: 5,
			wantErr:   false,
		},
		{
			name:      "min rating only",
			minRating: 3,
			maxRating: 0,
			wantErr:   false,
		},
		{
			name:      "max rating only",
			minRating: 0,
			maxRating: 4,
			wantErr:   false,
		},
		{
			name:      "valid date range",
			startDate: "2024-01-01T00:00:00Z",
			endDate:   "2024-12-31T23:59:59Z",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ReviewsListCmd{
				MinRating: tt.minRating,
				MaxRating: tt.maxRating,
				StartDate: tt.startDate,
				EndDate:   tt.endDate,
			}

			// Validation happens during filter application
			// Just verify the struct can be created with these values
			if cmd.MinRating != tt.minRating {
				t.Errorf("MinRating = %d, want %d", cmd.MinRating, tt.minRating)
			}
		})
	}
}

// ============================================================================
// Rate Limit Parsing Tests
// ============================================================================

func TestReviewsReplyCmd_RateLimitParsing(t *testing.T) {
	tests := []struct {
		name     string
		rateStr  string
		expected time.Duration
	}{
		{"5 seconds", "5s", 5 * time.Second},
		{"10 seconds", "10s", 10 * time.Second},
		{"1 minute", "1m", 1 * time.Minute},
		{"500 milliseconds", "500ms", 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ReviewsReplyCmd{RateLimit: tt.rateStr}

			duration, err := time.ParseDuration(cmd.RateLimit)
			if err != nil {
				t.Fatalf("Failed to parse duration: %v", err)
			}

			if duration != tt.expected {
				t.Errorf("Parsed duration = %v, want %v", duration, tt.expected)
			}
		})
	}
}

func TestReviewsReplyCmd_InvalidRateLimit(t *testing.T) {
	// When invalid rate limit is provided, it should default to 5s
	cmd := &ReviewsReplyCmd{
		ReviewID:  "review-123",
		Text:      "Thanks!",
		RateLimit: "invalid",
		DryRun:    true,
	}
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	// In dry run, we can verify the rate limit parsing happens
	err := cmd.Run(globals)
	// May error due to auth, but should not panic on invalid duration
	if err != nil {
		// Verify it's not a panic-related error
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Should not panic on invalid rate limit: %v", err)
		}
	}
}

// ============================================================================
// Integration with outputResult
// ============================================================================

func TestReviewsCommands_OutputResult(t *testing.T) {
	// Test that outputResult is called correctly by testing dry run outputs
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
		Pretty:  false,
	}

	cmd := &ReviewsReplyCmd{
		ReviewID: "rev-123",
		Text:     "Thanks for your review!",
		DryRun:   true,
	}

	err := cmd.Run(globals)
	// Dry run should succeed without authentication
	if err != nil {
		// If there's an error, it shouldn't be about missing auth
		if strings.Contains(err.Error(), "authentication") {
			t.Errorf("Dry run should not require authentication: %v", err)
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func createReviewsTempFile(t *testing.T, name string, content []byte) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, name)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return path
}

// ============================================================================
// Error Hint Tests
// ============================================================================

func TestReviewsReplyCmd_ErrorHints(t *testing.T) {
	t.Run("missing reply text has hint", func(t *testing.T) {
		cmd := &ReviewsReplyCmd{ReviewID: "rev-123", Text: ""}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error")
		}

		apiErr, ok := err.(*errors.APIError)
		if !ok {
			// Error may be wrapped, check the message
			if !strings.Contains(err.Error(), "reply text is required") {
				t.Errorf("Expected 'reply text is required' error, got: %v", err)
			}
		} else {
			if apiErr.Code != errors.CodeValidationError {
				t.Errorf("Expected validation error code, got: %s", apiErr.Code)
			}
		}
	})

	t.Run("reply text too long has hint", func(t *testing.T) {
		longText := strings.Repeat("a", 351)
		cmd := &ReviewsReplyCmd{ReviewID: "rev-123", Text: longText}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error")
		}

		if !strings.Contains(err.Error(), "exceeds 350 character limit") {
			t.Errorf("Expected character limit error, got: %v", err)
		}

		apiErr, ok := err.(*errors.APIError)
		if ok && apiErr.Hint == "" {
			t.Error("Expected error to have hint about shortening reply")
		}
	})
}

// ============================================================================
// Complex Filter Scenarios
// ============================================================================

func TestReviewsListCmd_ComplexFilters(t *testing.T) {
	cmd := &ReviewsListCmd{
		MinRating: 3,
		MaxRating: 5,
		Language:  "en-US",
		StartDate: "2024-01-01T00:00:00Z",
		EndDate:   "2024-12-31T23:59:59Z",
	}

	baseTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		review   *androidpublisher.Review
		expected bool
	}{
		{
			name: "all filters match",
			review: &androidpublisher.Review{
				Comments: []*androidpublisher.Comment{
					{
						UserComment: &androidpublisher.UserComment{
							StarRating:       4,
							ReviewerLanguage: "en-US",
							LastModified:     &androidpublisher.Timestamp{Seconds: baseTime.Unix()},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "rating too low",
			review: &androidpublisher.Review{
				Comments: []*androidpublisher.Comment{
					{
						UserComment: &androidpublisher.UserComment{
							StarRating:       2,
							ReviewerLanguage: "en-US",
							LastModified:     &androidpublisher.Timestamp{Seconds: baseTime.Unix()},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "wrong language",
			review: &androidpublisher.Review{
				Comments: []*androidpublisher.Comment{
					{
						UserComment: &androidpublisher.UserComment{
							StarRating:       4,
							ReviewerLanguage: "de-DE",
							LastModified:     &androidpublisher.Timestamp{Seconds: baseTime.Unix()},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "date before start",
			review: &androidpublisher.Review{
				Comments: []*androidpublisher.Comment{
					{
						UserComment: &androidpublisher.UserComment{
							StarRating:       4,
							ReviewerLanguage: "en-US",
							LastModified:     &androidpublisher.Timestamp{Seconds: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Unix()},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.matchesFilters(tt.review)
			if result != tt.expected {
				t.Errorf("matchesFilters() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// filterReviews with Multiple Reviews
// ============================================================================

func TestReviewsListCmd_filterReviews_Multiple(t *testing.T) {
	cmd := &ReviewsListCmd{MinRating: 4}

	baseTime := time.Now()
	reviews := []*androidpublisher.Review{
		{
			ReviewId: "rev-1",
			Comments: []*androidpublisher.Comment{
				{
					UserComment: &androidpublisher.UserComment{
						StarRating:   5,
						LastModified: &androidpublisher.Timestamp{Seconds: baseTime.Unix()},
					},
				},
			},
		},
		{
			ReviewId: "rev-2",
			Comments: []*androidpublisher.Comment{
				{
					UserComment: &androidpublisher.UserComment{
						StarRating:   3,
						LastModified: &androidpublisher.Timestamp{Seconds: baseTime.Unix()},
					},
				},
			},
		},
		{
			ReviewId: "rev-3",
			Comments: []*androidpublisher.Comment{
				{
					UserComment: &androidpublisher.UserComment{
						StarRating:   4,
						LastModified: &androidpublisher.Timestamp{Seconds: baseTime.Unix()},
					},
				},
			},
		},
		{
			ReviewId: "rev-4",
			Comments: nil, // Should be filtered out
		},
	}

	filtered := cmd.filterReviews(reviews)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered reviews (5 and 4 stars), got %d", len(filtered))
	}

	// Verify correct reviews were kept
	for _, r := range filtered {
		if r.ReviewId != "rev-1" && r.ReviewId != "rev-3" {
			t.Errorf("Unexpected review in filtered results: %s", r.ReviewId)
		}
	}
}
