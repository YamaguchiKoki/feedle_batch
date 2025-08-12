package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

const (
	defaultLimit     = 100
	maxLimit         = 100
	defaultSort      = "relevance"
	defaultTimeframe = "all"
)

type SearchParams struct {
	Query      string
	Subreddit  string
	Limit      int
	After      string // for pagination
	Sort       string // relevance, hot, top, new, comments
	Time       string // all, year, month, week, day, hour
	RestrictSR bool   // restrict search to subreddit
}

// RedditFetcher handles Reddit API interactions
type RedditFetcher struct {
	baseURL   string
	userAgent string
	client    *http.Client
	auth      *RedditAuth
}

func NewRedditFetcher(userAgent string, auth *RedditAuth) *RedditFetcher {
	if userAgent == "" {
		userAgent = "Go Reddit Fetcher/1.0"
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	baseURL := "https://www.reddit.com"
	if auth != nil {
		baseURL = "https://oauth.reddit.com"
	}

	return &RedditFetcher{
		baseURL:   baseURL,
		userAgent: userAgent,
		client:    client,
		auth:      auth,
	}
}

func NewRedditFetcherWithClient(userAgent string, auth *RedditAuth, client *http.Client) *RedditFetcher {
	if userAgent == "" {
		userAgent = "Go Reddit Fetcher/1.0"
	}

	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	baseURL := "https://www.reddit.com"
	if auth != nil {
		baseURL = "https://oauth.reddit.com"
	}

	return &RedditFetcher{
		baseURL:   baseURL,
		userAgent: userAgent,
		client:    client,
		auth:      auth,
	}
}

func (rf *RedditFetcher) Name() string {
	return "reddit"
}

func (rf *RedditFetcher) Fetch(ctx context.Context, config model.RedditFetchConfigDetail) ([]*model.FetchedData, error) {
	if config.Subreddit == "" && len(config.Keywords) == 0 {
		return nil, fmt.Errorf("no subreddit or keywords provided")
	}

	var allResults []*model.FetchedData

	if len(config.Keywords) == 0 {
		posts, err := rf.fetchSubredditPosts(config.Subreddit, config.LimitCount)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch subreddit posts: %w", err)
		}
		// Set ConfigID for all posts (use UserFetchConfigID, not the reddit config ID)
		for _, post := range posts {
			if post != nil {
				post.ConfigID = config.UserFetchConfigID
			}
		}
		return posts, nil
	}

	// Search with keywords
	for _, keyword := range config.Keywords {
		fmt.Println("iter", keyword)
		params := SearchParams{
			Query:      keyword,
			Subreddit:  config.Subreddit,
			Limit:      config.LimitCount,
			Sort:       config.SortBy,
			Time:       config.TimeFilter,
			RestrictSR: config.Subreddit != "", // restrict to subreddit if specified
		}

		posts, err := rf.searchPosts(ctx, params)
		if err != nil {
			// Log error but continue with other keywords
			fmt.Printf("failed to search for keyword %s: %v\n", keyword, err)
			continue
		}

		// Set ConfigID for all posts (use UserFetchConfigID, not the reddit config ID)
		for _, post := range posts {
			if post != nil {
				post.ConfigID = config.UserFetchConfigID
			}
		}

		allResults = append(allResults, posts...)
	}

	// Remove duplicates based on post ID
	return rf.deduplicatePosts(allResults), nil
}

func (rf *RedditFetcher) searchPosts(ctx context.Context, params SearchParams) ([]*model.FetchedData, error) {
	var allPosts []*model.FetchedData

	// Handle pagination
	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return allPosts, ctx.Err()
		default:
			// Continue processing
		}

		posts, nextAfter, err := rf.searchPage(params)
		if err != nil {
			return allPosts, err
		}

		allPosts = append(allPosts, posts...)

		// Check if we've reached the limit or no more pages
		if nextAfter == "" || (params.Limit > 0 && len(allPosts) >= params.Limit) {
			break
		}

		params.After = nextAfter
	}

	// Trim to requested limit
	if params.Limit > 0 && len(allPosts) > params.Limit {
		allPosts = allPosts[:params.Limit]
	}

	return allPosts, nil
}

// searchPage fetches a single page of search results
func (rf *RedditFetcher) searchPage(params SearchParams) ([]*model.FetchedData, string, error) {
	// Build search URL
	searchURL := rf.buildSearchURL(params)

	posts, nextAfter, err := rf.fetchFromURL(searchURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch search results: %w", err)
	}

	return posts, nextAfter, nil
}

// buildSearchURL constructs the search URL with parameters
func (rf *RedditFetcher) buildSearchURL(params SearchParams) string {
	var endpoint string
	if params.Subreddit != "" {
		endpoint = fmt.Sprintf("%s/r/%s/search.json", rf.baseURL, params.Subreddit)
	} else {
		endpoint = fmt.Sprintf("%s/search.json", rf.baseURL)
	}

	// Build query parameters
	query := url.Values{}
	query.Set("q", params.Query)

	// Set limit (max 100 per page)
	limit := params.Limit
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}
	query.Set("limit", fmt.Sprintf("%d", limit))

	// Pagination
	if params.After != "" {
		query.Set("after", params.After)
	}

	// Sort order
	if params.Sort != "" {
		query.Set("sort", params.Sort)
	} else {
		query.Set("sort", defaultSort)
	}

	// Time filter (for top/controversial sort)
	if params.Time != "" && (params.Sort == "top" || params.Sort == "controversial") {
		query.Set("t", params.Time)
	}

	// Restrict to subreddit
	if params.RestrictSR && params.Subreddit != "" {
		query.Set("restrict_sr", "true")
	}

	// Include NSFW content (optional)
	query.Set("include_over_18", "true")

	return fmt.Sprintf("%s?%s", endpoint, query.Encode())
}

// fetchSubredditPosts fetches posts from a specific subreddit
func (rf *RedditFetcher) fetchSubredditPosts(subreddit string, limit int) ([]*model.FetchedData, error) {
	if limit <= 0 {
		limit = defaultLimit
	}

	url := fmt.Sprintf("%s/r/%s.json?limit=%d", rf.baseURL, subreddit, limit)

	posts, _, err := rf.fetchFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subreddit posts: %w", err)
	}

	return posts, nil
}

// fetchFromURL fetches data from a Reddit URL and returns posts and pagination info
func (rf *RedditFetcher) fetchFromURL(url string) ([]*model.FetchedData, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("User-Agent", rf.userAgent)
	req.Header.Set("Accept", "application/json")

	// Handle authentication
	if rf.auth != nil {
		if err := rf.addAuthHeaders(req); err != nil {
			return nil, "", fmt.Errorf("failed to add auth headers: %w", err)
		}
	}

	// Respect rate limiting
	time.Sleep(time.Second) // Basic rate limiting - adjust as needed

	resp, err := rf.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch from Reddit: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("failed to close response body: %v\n", cerr)
		}
	}()

	// Handle HTTP errors
	if err := rf.handleHTTPError(resp); err != nil {
		return nil, "", err
	}

	// Decode response
	var redditResp RedditResponse
	if err := json.NewDecoder(resp.Body).Decode(&redditResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode Reddit response: %w", err)
	}

	// Transform posts
	results := lo.FilterMap(redditResp.Data.Children, func(post RedditPost, _ int) (*model.FetchedData, bool) {
		if post.Kind != "t3" { // t3 = link/post
			return nil, false
		}
		transformed := rf.transformPost(post)
		return transformed, transformed != nil
	})

	return results, redditResp.Data.After, nil
}

// addAuthHeaders adds authentication headers to the request
func (rf *RedditFetcher) addAuthHeaders(req *http.Request) error {
	token, err := rf.auth.GetAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Update URL to use OAuth endpoint
	if !strings.Contains(req.URL.String(), "oauth.reddit.com") {
		newURL := strings.Replace(req.URL.String(), "www.reddit.com", "oauth.reddit.com", 1)
		parsedURL, err := url.Parse(newURL)
		if err != nil {
			return fmt.Errorf("failed to parse OAuth URL: %w", err)
		}
		req.URL = parsedURL
	}

	return nil
}

// handleHTTPError checks and handles HTTP error responses
func (rf *RedditFetcher) handleHTTPError(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// Read error response
	var errorResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
		if message, ok := errorResp["message"].(string); ok {
			return fmt.Errorf("reddit API error (status %d): %s", resp.StatusCode, message)
		}
	}

	// Handle specific status codes
	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limit exceeded, retry after %s", resp.Header.Get("X-Ratelimit-Reset"))
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed")
	case http.StatusForbidden:
		return fmt.Errorf("access forbidden - check permissions")
	default:
		return fmt.Errorf("reddit API returned status %d", resp.StatusCode)
	}
}

// deduplicatePosts removes duplicate posts based on ID
func (rf *RedditFetcher) deduplicatePosts(posts []*model.FetchedData) []*model.FetchedData {
	seen := make(map[uuid.UUID]bool)
	var unique []*model.FetchedData

	for _, post := range posts {
		if post == nil || post.ID == uuid.Nil {
			continue
		}
		if !seen[post.ID] {
			seen[post.ID] = true
			unique = append(unique, post)
		}
	}

	return unique
}

// transformPost converts a Reddit post to FetchedData
func (rf *RedditFetcher) transformPost(post RedditPost) *model.FetchedData {
	if post.Data == nil {
		return nil
	}

	// Generate UUID from Reddit post ID
	postUUID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("reddit:%s", post.Data.ID)))

	// Convert Unix timestamp to time.Time
	publishedAt := time.Unix(int64(post.Data.CreatedUTC), 0)

	// Build metadata
	metadata := map[string]interface{}{
		"score":        post.Data.Score,
		"num_comments": post.Data.NumComments,
		"subreddit":    post.Data.Subreddit,
		"permalink":    post.Data.Permalink,
		"over_18":      post.Data.Over18,
	}

	// Extract media URLs if present
	var mediaURLs []string
	if post.Data.URL != "" && !strings.HasPrefix(post.Data.URL, "https://www.reddit.com") {
		// External URL might be media
		mediaURLs = append(mediaURLs, post.Data.URL)
	}

	// Generate tags
	tags := []string{
		fmt.Sprintf("subreddit:%s", post.Data.Subreddit),
		fmt.Sprintf("author:%s", post.Data.Author),
	}
	if post.Data.Over18 {
		tags = append(tags, "nsfw")
	}

	now := time.Now()

	return &model.FetchedData{
		ID:           postUUID,
		Source:       "reddit",
		Title:        post.Data.Title,
		Content:      post.Data.Selftext,
		URL:          fmt.Sprintf("https://reddit.com%s", post.Data.Permalink),
		AuthorName:   post.Data.Author,
		SourceItemID: post.Data.ID,
		PublishedAt:  &publishedAt,
		Tags:         tags,
		MediaURLs:    mediaURLs,
		Metadata:     metadata,
		FetchedAt:    now,
		CreatedAt:    now,
	}
}

// RedditResponse represents the top-level Reddit API response
type RedditResponse struct {
	Kind string `json:"kind"`
	Data struct {
		After    string       `json:"after"`
		Before   string       `json:"before"`
		Children []RedditPost `json:"children"`
	} `json:"data"`
}

// RedditPost represents a Reddit post in the API response
type RedditPost struct {
	Kind string          `json:"kind"`
	Data *RedditPostData `json:"data"`
}

// RedditPostData contains the actual post information
type RedditPostData struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Selftext    string  `json:"selftext"`
	URL         string  `json:"url"`
	Author      string  `json:"author"`
	Score       int     `json:"score"`
	NumComments int     `json:"num_comments"`
	CreatedUTC  float64 `json:"created_utc"`
	Subreddit   string  `json:"subreddit"`
	Permalink   string  `json:"permalink"`
	Over18      bool    `json:"over_18"`
	// Add other fields as needed
}
