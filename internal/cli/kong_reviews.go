package cli

import (
	"github.com/dl-alexandre/gpd/internal/errors"
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

// Run executes the reviews list command.
func (cmd *ReviewsListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "reviews list not yet implemented")
}

// ReviewsGetCmd gets a single review by ID.
type ReviewsGetCmd struct {
	ReviewID            string `arg:"" optional:"" help:"Review ID to fetch"`
	IncludeReviewText   bool   `help:"Include review text in output"`
	TranslationLanguage string `help:"Language for translated review"`
}

// Run executes the reviews get command.
func (cmd *ReviewsGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "reviews get not yet implemented")
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

// Run executes the reviews reply command.
func (cmd *ReviewsReplyCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "reviews reply not yet implemented")
}

// ReviewsResponseGetCmd gets a review response.
type ReviewsResponseGetCmd struct {
	ReviewID string `arg:"" optional:"" help:"Review ID to fetch response for"`
}

// Run executes the reviews response get command.
func (cmd *ReviewsResponseGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "reviews response get not yet implemented")
}

// ReviewsResponseDeleteCmd deletes a review response.
type ReviewsResponseDeleteCmd struct {
	ReviewID string `arg:"" optional:"" help:"Review ID to delete response for"`
}

// Run executes the reviews response delete command.
func (cmd *ReviewsResponseDeleteCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "reviews response delete not yet implemented")
}
