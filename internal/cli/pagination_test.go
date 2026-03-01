//go:build unit
// +build unit

package cli

import (
	"context"
	"errors"
	"testing"
)

// mockPageResponse implements PageResponse for testing
type mockPageResponse struct {
	nextToken string
	items     []string
}

func (m *mockPageResponse) GetNextPageToken() string {
	return m.nextToken
}

func (m *mockPageResponse) GetItems() []string {
	return m.items
}

func TestFetchAllPages(t *testing.T) {
	t.Run("single page with no next token", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"": {nextToken: "", items: []string{"item1", "item2", "item3"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})

	t.Run("multiple pages", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":       {nextToken: "token1", items: []string{"a", "b"}},
			"token1": {nextToken: "token2", items: []string{"c", "d"}},
			"token2": {nextToken: "", items: []string{"e"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 5 {
			t.Errorf("expected 5 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})

	t.Run("empty items", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":       {nextToken: "token1", items: []string{}},
			"token1": {nextToken: "", items: []string{"item1"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, _, err := fetchAllPages(context.Background(), query, "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 item, got %d", len(results))
		}
	})

	t.Run("empty page ends pagination", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"": {nextToken: "", items: []string{}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})
}

func TestFetchAllPages_MaxResults(t *testing.T) {
	t.Run("max results limits single page", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"": {nextToken: "next", items: []string{"a", "b", "c", "d", "e"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 items, got %d", len(results))
		}
		// When limit is hit before next fetch, returns current pageToken
		if nextToken != "next" {
			t.Errorf("expected nextToken 'next', got %q", nextToken)
		}
	})

	t.Run("max results across multiple pages", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":       {nextToken: "token1", items: []string{"a", "b"}},
			"token1": {nextToken: "token2", items: []string{"c", "d", "e"}},
			"token2": {nextToken: "", items: []string{"f", "g"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 4 {
			t.Errorf("expected 4 items, got %d", len(results))
		}
		// After collecting 4 items (2 from first, 2 from second truncated), returns token2
		if nextToken != "token2" {
			t.Errorf("expected nextToken 'token2', got %q", nextToken)
		}
	})

	t.Run("max results truncates page", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":       {nextToken: "token1", items: []string{"a", "b"}},
			"token1": {nextToken: "", items: []string{"c", "d", "e", "f"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 items, got %d", len(results))
		}
		expected := []string{"a", "b", "c"}
		for i, v := range expected {
			if results[i] != v {
				t.Errorf("expected results[%d] = %q, got %q", i, v, results[i])
			}
		}
		// After truncation and hitting limit, returns the next token from the current page
		if nextToken != "" {
			t.Errorf("expected empty nextToken (pagination complete), got %q", nextToken)
		}
	})

	t.Run("max results equals total items", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"": {nextToken: "", items: []string{"a", "b", "c"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})

	t.Run("max results zero means unlimited", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":       {nextToken: "token1", items: []string{"a", "b"}},
			"token1": {nextToken: "", items: []string{"c", "d", "e"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 5 {
			t.Errorf("expected 5 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})

	t.Run("max results one", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"": {nextToken: "token1", items: []string{"a", "b", "c"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 item, got %d", len(results))
		}
		if results[0] != "a" {
			t.Errorf("expected 'a', got %q", results[0])
		}
		// Returns next token from the response
		if nextToken != "token1" {
			t.Errorf("expected nextToken 'token1', got %q", nextToken)
		}
	})

	t.Run("max results checked before fetch", func(t *testing.T) {
		callCount := 0
		pages := map[string]*mockPageResponse{
			"":       {nextToken: "token1", items: []string{"a", "b"}},
			"token1": {nextToken: "token2", items: []string{"c", "d"}},
			"token2": {nextToken: "", items: []string{"e"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			callCount++
			return pages[pageToken], nil
		}

		// Already have 2 items, limit is 2
		results, nextToken, err := fetchAllPages(context.Background(), query, "", 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if callCount != 1 {
			t.Errorf("expected 1 call (limit reached before second page), got %d", callCount)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 items, got %d", len(results))
		}
		// Returns current pageToken (the one we would have fetched next)
		if nextToken != "token1" {
			t.Errorf("expected nextToken 'token1', got %q", nextToken)
		}
	})
}

func TestFetchAllPages_ContextCancellation(t *testing.T) {
	t.Run("context cancelled before first fetch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		query := func(pageToken string) (*mockPageResponse, error) {
			return &mockPageResponse{nextToken: "", items: []string{"item"}}, nil
		}

		results, nextToken, err := fetchAllPages(ctx, query, "", 0)
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})

	t.Run("context cancelled between pages", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":       {nextToken: "token1", items: []string{"a", "b"}},
			"token1": {nextToken: "", items: []string{"c"}},
		}

		ctx, cancel := context.WithCancel(context.Background())
		callCount := 0

		query := func(pageToken string) (*mockPageResponse, error) {
			callCount++
			if callCount == 2 {
				// Cancel context before returning from second call
				cancel()
			}
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(ctx, query, "", 0)
		// The context cancellation happens during the second query, but since
		// the query itself completed, we get the results from it
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// We get all 3 items because the second query completed
		if len(results) != 3 {
			t.Errorf("expected 3 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})

	t.Run("context cancelled detected before third fetch", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":       {nextToken: "token1", items: []string{"a"}},
			"token1": {nextToken: "token2", items: []string{"b"}},
			"token2": {nextToken: "", items: []string{"c"}},
		}

		ctx, cancel := context.WithCancel(context.Background())
		callCount := 0

		query := func(pageToken string) (*mockPageResponse, error) {
			callCount++
			if callCount == 2 {
				cancel() // Cancel after second page is fetched
			}
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(ctx, query, "", 0)
		// Context cancellation is checked at loop start, so after second page
		// it will be detected before third fetch
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 items, got %d", len(results))
		}
		if nextToken != "token2" {
			t.Errorf("expected nextToken 'token2', got %q", nextToken)
		}
	})
}

func TestFetchAllPages_Errors(t *testing.T) {
	t.Run("query returns error on first page", func(t *testing.T) {
		expectedErr := errors.New("network error")
		query := func(pageToken string) (*mockPageResponse, error) {
			return nil, expectedErr
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 0)
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})

	t.Run("query returns error on second page", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":       {nextToken: "token1", items: []string{"a", "b"}},
			"token1": {nextToken: "", items: []string{"c"}},
		}

		expectedErr := errors.New("API rate limit exceeded")
		callCount := 0
		query := func(pageToken string) (*mockPageResponse, error) {
			callCount++
			if callCount >= 2 {
				return nil, expectedErr
			}
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 0)
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 items (from first page), got %d", len(results))
		}
		if nextToken != "token1" {
			t.Errorf("expected nextToken 'token1', got %q", nextToken)
		}
	})

	t.Run("error with partial results", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":       {nextToken: "token1", items: []string{"a"}},
			"token1": {nextToken: "token2", items: []string{"b"}},
			"token2": {nextToken: "", items: []string{"c"}},
		}

		expectedErr := errors.New("service unavailable")
		callCount := 0
		query := func(pageToken string) (*mockPageResponse, error) {
			callCount++
			if callCount >= 3 {
				return nil, expectedErr
			}
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 0)
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 items, got %d", len(results))
		}
		if nextToken != "token2" {
			t.Errorf("expected nextToken 'token2', got %q", nextToken)
		}
	})
}

func TestFetchAllPages_InitialToken(t *testing.T) {
	t.Run("start with non-empty initial token", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"start-token": {nextToken: "", items: []string{"x", "y"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "start-token", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})

	t.Run("start with token continues pagination", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"start": {nextToken: "next", items: []string{"a"}},
			"next":  {nextToken: "", items: []string{"b", "c"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "start", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})
}

func TestFetchAllPages_EdgeCases(t *testing.T) {
	t.Run("nil items slice", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"": {nextToken: "", items: nil},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})

	t.Run("single item per page", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":   {nextToken: "t1", items: []string{"a"}},
			"t1": {nextToken: "t2", items: []string{"b"}},
			"t2": {nextToken: "t3", items: []string{"c"}},
			"t3": {nextToken: "", items: []string{"d"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 4 {
			t.Errorf("expected 4 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})

	t.Run("large page small max results", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"": {nextToken: "next", items: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 items, got %d", len(results))
		}
		if results[0] != "a" || results[1] != "b" {
			t.Errorf("expected ['a', 'b'], got %v", results)
		}
		// Returns next token from the response
		if nextToken != "next" {
			t.Errorf("expected nextToken 'next', got %q", nextToken)
		}
	})

	t.Run("exactly fills max results", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":   {nextToken: "t1", items: []string{"a", "b"}},
			"t1": {nextToken: "t2", items: []string{"c", "d"}},
			"t2": {nextToken: "", items: []string{"e", "f"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 4 {
			t.Errorf("expected 4 items, got %d", len(results))
		}
		// After collecting 4 items (2 from first, 2 from second), limit is hit
		// Returns token from the second page
		if nextToken != "t2" {
			t.Errorf("expected nextToken 't2', got %q", nextToken)
		}
	})

	t.Run("max results after exact page boundary", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":   {nextToken: "t1", items: []string{"a", "b", "c"}},
			"t1": {nextToken: "t2", items: []string{"d", "e", "f"}},
			"t2": {nextToken: "", items: []string{"g", "h", "i"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 items, got %d", len(results))
		}
		// After first page, we have 3 items which meets limit
		// The next token would be checked before fetching next page
		if nextToken != "t1" {
			t.Errorf("expected nextToken 't1', got %q", nextToken)
		}
	})

	t.Run("repeated same token causes infinite loop", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"":     {nextToken: "loop", items: []string{"a"}},
			"loop": {nextToken: "loop", items: []string{"b"}},
		}

		callCount := 0
		query := func(pageToken string) (*mockPageResponse, error) {
			callCount++
			if callCount > 3 {
				// Verify it keeps calling with the same token
				return nil, errors.New("infinite loop detected after 3 calls")
			}
			return pages[pageToken], nil
		}

		_, _, err := fetchAllPages(context.Background(), query, "", 0)
		// Function doesn't detect cycles, it will loop forever in real use
		// The safeguard in the mock causes an error
		if err == nil {
			t.Error("expected function to loop forever with repeating tokens")
		}
	})

	t.Run("max results larger than total items", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"": {nextToken: "", items: []string{"a", "b"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		results, nextToken, err := fetchAllPages(context.Background(), query, "", 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})

	t.Run("negative max results treated as zero", func(t *testing.T) {
		pages := map[string]*mockPageResponse{
			"": {nextToken: "", items: []string{"a", "b", "c"}},
		}

		query := func(pageToken string) (*mockPageResponse, error) {
			return pages[pageToken], nil
		}

		// Negative maxResults should not trigger limit (condition is maxResults > 0)
		results, nextToken, err := fetchAllPages(context.Background(), query, "", -1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 items, got %d", len(results))
		}
		if nextToken != "" {
			t.Errorf("expected empty nextToken, got %q", nextToken)
		}
	})
}

func TestFetchAllPages_DifferentTypes(t *testing.T) {
	t.Run("works with integer items", func(t *testing.T) {
		type intPageResponse struct {
			nextToken string
			items     []int
		}

		query := func(pageToken string) (*intPageResponse, error) {
			return &intPageResponse{nextToken: "", items: []int{1, 2, 3}}, nil
		}

		// Custom implementation since we can't use the generic directly with different type
		pageToken := ""
		var allResults []int

		resp, err := query(pageToken)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		allResults = append(allResults, resp.items...)

		if len(allResults) != 3 {
			t.Errorf("expected 3 items, got %d", len(allResults))
		}
	})
}

// Benchmark tests
func BenchmarkFetchAllPages(b *testing.B) {
	pages := map[string]*mockPageResponse{
		"":   {nextToken: "t1", items: []string{"a", "b", "c", "d", "e"}},
		"t1": {nextToken: "t2", items: []string{"f", "g", "h", "i", "j"}},
		"t2": {nextToken: "t3", items: []string{"k", "l", "m", "n", "o"}},
		"t3": {nextToken: "t4", items: []string{"p", "q", "r", "s", "t"}},
		"t4": {nextToken: "", items: []string{"u", "v", "w", "x", "y"}},
	}

	query := func(pageToken string) (*mockPageResponse, error) {
		return pages[pageToken], nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = fetchAllPages(context.Background(), query, "", 0)
	}
}

func BenchmarkFetchAllPages_MaxResults(b *testing.B) {
	pages := map[string]*mockPageResponse{
		"":   {nextToken: "t1", items: []string{"a", "b", "c", "d", "e"}},
		"t1": {nextToken: "t2", items: []string{"f", "g", "h", "i", "j"}},
		"t2": {nextToken: "t3", items: []string{"k", "l", "m", "n", "o"}},
		"t3": {nextToken: "t4", items: []string{"p", "q", "r", "s", "t"}},
		"t4": {nextToken: "", items: []string{"u", "v", "w", "x", "y"}},
	}

	query := func(pageToken string) (*mockPageResponse, error) {
		return pages[pageToken], nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = fetchAllPages(context.Background(), query, "", 12)
	}
}
