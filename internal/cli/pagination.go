package cli

import (
	"context"
)

// PageResponse is the interface that any paginated API response must implement.
type PageResponse[T any] interface {
	GetNextPageToken() string
	GetItems() []T
}

// PageQuery is a function that fetches a single page given a page token.
// It returns the response which must implement PageResponse, or an error.
type PageQuery[T any, R PageResponse[T]] func(pageToken string) (R, error)

// fetchAllPages fetches all pages of results using the provided query function.
// It handles pagination automatically and respects context cancellation.
// If maxResults > 0, fetching stops when that many items have been collected.
//
// Type parameters:
//   - T: The type of individual items in the result (e.g., *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow)
//   - R: The response type that implements PageResponse[T]
//
// Parameters:
//   - ctx: Context for cancellation
//   - query: Function that performs a single page query
//   - initialToken: Starting page token (empty string for first page)
//   - maxResults: Maximum number of items to fetch (0 = unlimited)
//
// Returns:
//   - allResults: Slice of all collected items
//   - nextToken: The next page token if there are more results (for partial fetches)
//   - err: Any error encountered during fetching
func fetchAllPages[T any, R PageResponse[T]](
	ctx context.Context,
	query PageQuery[T, R],
	initialToken string,
	maxResults int,
) (allResults []T, nextToken string, err error) {
	pageToken := initialToken

	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return allResults, pageToken, ctx.Err()
		default:
		}

		// Check max results limit before fetching
		if maxResults > 0 && len(allResults) >= maxResults {
			return allResults, pageToken, nil
		}

		// Fetch the page
		resp, err := query(pageToken)
		if err != nil {
			return allResults, pageToken, err
		}

		// Collect items from this page
		items := resp.GetItems()
		if len(items) > 0 {
			// If we have a max results limit, only take what we need
			if maxResults > 0 && len(allResults)+len(items) > maxResults {
				remaining := maxResults - len(allResults)
				items = items[:remaining]
			}
			allResults = append(allResults, items...)
		}

		// Get next page token
		pageToken = resp.GetNextPageToken()

		// Stop if no more pages
		if pageToken == "" {
			return allResults, "", nil
		}

		// Check max results limit after this page
		if maxResults > 0 && len(allResults) >= maxResults {
			return allResults, pageToken, nil
		}
	}
}
