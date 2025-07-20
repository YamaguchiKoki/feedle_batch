package reddit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/YamaguchiKoki/feedle_batch/internal/fetcher"
	"github.com/YamaguchiKoki/feedle_batch/internal/models"
	"github.com/samber/lo"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RedditPostData struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Selftext    string  `json:"selftext"`
	Author      string  `json:"author"`
	CreatedUTC  float64 `json:"created_utc"`
	Score       int     `json:"score"`
	NumComments int     `json:"num_comments"`
	Subreddit   string  `json:"subreddit"`
	Permalink   string  `json:"permalink"`
	URL         string  `json:"url"`
	IsSelf      bool    `json:"is_self"`
}

type RedditPost struct {
	Data RedditPostData `json:"data"`
}

type RedditResponse struct {
	Data struct {
		Children []RedditPost `json:"children"`
		After    string       `json:"after"`
	} `json:"data"`
}

type RedditFetcher struct {
	client    HTTPClient
	auth      *RedditAuth
	baseURL   string
	userAgent string
}

func NewRedditFetcher(client HTTPClient, clientID, clientSecret, username string) *RedditFetcher {
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	var auth *RedditAuth
	if clientID != "" && clientSecret != "" {
		auth = NewRedditAuth(clientID, clientSecret, username)
	}

	userAgent := fmt.Sprintf("golang:feedle-batch:v1.0.0 (by /u/%s)", username)
	if username == "" {
		userAgent = "golang:feedle-batch:v1.0.0"
	}

	return &RedditFetcher{
		client:    client,
		auth:      auth,
		baseURL:   "https://oauth.reddit.com",
		userAgent: userAgent,
	}
}

func (rf *RedditFetcher) Name() string {
	return "reddit"
}

func (rf *RedditFetcher) Fetch(config fetcher.FetchConfig) ([]*models.FetchedData, error) {
	if len(config.Reddit.Subreddits) == 0 && len(config.Keywords) == 0 {
		return nil, fmt.Errorf("no subreddits or keywords provided")
	}

	var allResults []*models.FetchedData

	// Fetch from subreddits
	for _, subreddit := range config.Reddit.Subreddits {
		results, err := rf.fetchSubreddit(subreddit, config.Limit)
		if err != nil {
			fmt.Printf("Error fetching r/%s: %v\n", subreddit, err)
			continue
		}
		allResults = append(allResults, results...)
	}

	// Search by keywords
	if len(config.Keywords) > 0 {
		query := strings.Join(config.Keywords, " OR ")
		results, err := rf.search(query, config.Limit)
		if err != nil {
			fmt.Printf("Error searching: %v\n", err)
		} else {
			allResults = append(allResults, results...)
		}
	}

	return rf.deduplicateResults(allResults), nil
}

func (rf *RedditFetcher) fetchSubreddit(subreddit string, limit int) ([]*models.FetchedData, error) {
	url := fmt.Sprintf("%s/r/%s.json?limit=%d&raw_json=1", rf.baseURL, subreddit, limit)
	return rf.fetchFromURL(url)
}

func (rf *RedditFetcher) search(query string, limit int) ([]*models.FetchedData, error) {
	url := fmt.Sprintf("%s/search.json?q=%s&limit=%d&raw_json=1&sort=new", rf.baseURL, query, limit)
	return rf.fetchFromURL(url)
}

func (rf *RedditFetcher) fetchFromURL(url string) ([]*models.FetchedData, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", rf.userAgent)

	// 認証が設定されている場合
	if rf.auth != nil {
		token, err := rf.auth.GetAccessToken()
		if err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		// 認証時はURLをoauth.reddit.comに変更
		if !strings.Contains(url, "oauth.reddit.com") {
			url = strings.Replace(url, "www.reddit.com", "oauth.reddit.com", 1)
			req.URL, _ = req.URL.Parse(url)
		}
	}

	resp, err := rf.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Reddit: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reddit API returned status %d", resp.StatusCode)
	}

	var redditResp RedditResponse
	if err := json.NewDecoder(resp.Body).Decode(&redditResp); err != nil {
		return nil, fmt.Errorf("failed to decode Reddit response: %w", err)
	}

	results := lo.Map(redditResp.Data.Children, func(post RedditPost, _ int) *models.FetchedData {
		return rf.transformPost(post)
	})

	return lo.Filter(results, func(item *models.FetchedData, _ int) bool {
		return item != nil
	}), nil
}

func (rf *RedditFetcher) transformPost(post RedditPost) *models.FetchedData {
	data := post.Data
	createdAt := time.Unix(int64(data.CreatedUTC), 0)

	fetchedData := &models.FetchedData{
		Title:       data.Title,
		Content:     data.Selftext,
		URL:         fmt.Sprintf("https://reddit.com%s", data.Permalink),
		AuthorName:  data.Author,
		AuthorID:    data.Author,
		PublishedAt: &createdAt,
		Engagement: map[string]interface{}{
			"score":    data.Score,
			"comments": data.NumComments,
		},
		Tags:      []string{data.Subreddit},
		FetchedAt: time.Now(),
		RawData: map[string]interface{}{
			"reddit_id": data.ID,
			"is_self":   data.IsSelf,
		},
	}

	if !data.IsSelf && data.URL != "" && !strings.Contains(data.URL, "reddit.com") {
		fetchedData.Media = []string{data.URL}
	}

	return fetchedData
}

func (rf *RedditFetcher) deduplicateResults(results []*models.FetchedData) []*models.FetchedData {
	seen := make(map[string]bool)

	return lo.Filter(results, func(item *models.FetchedData, _ int) bool {
		redditID, ok := item.RawData["reddit_id"].(string)
		if !ok || redditID == "" {
			return true
		}

		if seen[redditID] {
			return false
		}
		seen[redditID] = true
		return true
	})
}
