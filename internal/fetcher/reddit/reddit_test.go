package reddit

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/YamaguchiKoki/feedle_batch/internal/fetcher"
	"github.com/YamaguchiKoki/feedle_batch/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient for testing
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestRedditFetcher_Fetch(t *testing.T) {
	tests := []struct {
		name      string
		config    fetcher.FetchConfig
		mockSetup func(*MockHTTPClient)
		wantCount int
		wantErr   bool
		errMsg    string
	}{
		// 既存のテストケースをそのまま使用
		{
			name: "successful subreddit fetch",
			config: fetcher.FetchConfig{
				Reddit: struct{ Subreddits []string }{
					Subreddits: []string{"golang"},
				},
				Limit: 10,
			},
			mockSetup: func(m *MockHTTPClient) {
				response := `{
                    "data": {
                        "children": [
                            {
                                "data": {
                                    "id": "1",
                                    "title": "Test Post",
                                    "selftext": "Content here",
                                    "author": "user1",
                                    "created_utc": 1704067200,
                                    "score": 10,
                                    "num_comments": 5,
                                    "subreddit": "golang",
                                    "permalink": "/r/golang/comments/1/test_post/",
                                    "url": "https://reddit.com/r/golang/comments/1/test_post/",
                                    "is_self": true
                                }
                            },
                            {
                                "data": {
                                    "id": "2",
                                    "title": "Another Post",
                                    "selftext": "",
                                    "author": "user2",
                                    "created_utc": 1704067200,
                                    "score": 20,
                                    "num_comments": 10,
                                    "subreddit": "golang",
                                    "permalink": "/r/golang/comments/2/another_post/",
                                    "url": "https://example.com/article",
                                    "is_self": false
                                }
                            }
                        ]
                    }
                }`
				m.On("Do", mock.Anything).Return(&http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(response)),
				}, nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "no subreddits or keywords provided",
			config: fetcher.FetchConfig{
				Limit: 10,
			},
			mockSetup: func(m *MockHTTPClient) {},
			wantCount: 0,
			wantErr:   true,
			errMsg:    "no subreddits or keywords provided",
		},
		{
			name: "handle API error gracefully",
			config: fetcher.FetchConfig{
				Reddit: struct{ Subreddits []string }{
					Subreddits: []string{"nonexistent"},
				},
				Limit: 10,
			},
			mockSetup: func(m *MockHTTPClient) {
				m.On("Do", mock.Anything).Return(&http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}, nil)
			},
			wantCount: 0,
			wantErr:   false, // エラーでも続行
		},
		{
			name: "deduplicate results from multiple sources",
			config: fetcher.FetchConfig{
				Keywords: []string{"golang"},
				Reddit: struct{ Subreddits []string }{
					Subreddits: []string{"golang"},
				},
				Limit: 10,
			},
			mockSetup: func(m *MockHTTPClient) {
				response := `{
					"data": {
						"children": [
							{
								"data": {
									"id": "1",
									"title": "Duplicate Post",
									"selftext": "Content",
									"author": "user1",
									"created_utc": 1704067200,
									"score": 10,
									"num_comments": 5,
									"subreddit": "golang",
									"permalink": "/r/golang/comments/1/duplicate_post/",
									"url": "https://reddit.com/r/golang/comments/1/duplicate_post/",
									"is_self": true
								}
							}
						]
					}
				}`
				// subreddit fetch用
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.Contains(req.URL.String(), "/r/golang.json")
				})).Return(&http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(response)),
				}, nil).Once()

				// search用
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.Contains(req.URL.String(), "/search.json")
				})).Return(&http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(response)),
				}, nil).Once()
			},
			wantCount: 1, // 重複除去されて1件
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockHTTPClient)
			tt.mockSetup(mockClient)

			// 引数を変更: clientID, clientSecret, username を追加
			rf := NewRedditFetcher(mockClient, "", "", "")
			results, err := rf.Fetch(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.wantCount)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// RedditAuth のテストを追加
func TestRedditAuth_GetAccessToken(t *testing.T) {
	t.Run("successful token retrieval", func(t *testing.T) {
		auth := NewRedditAuth("test-client-id", "test-client-secret", "test-user")

		// 実際のHTTPリクエストは行わないので、このテストは
		// 主にAuthオブジェクトの作成を確認
		assert.NotNil(t, auth)
		assert.Equal(t, "test-client-id", auth.clientID)
		assert.Equal(t, "test-client-secret", auth.clientSecret)
		assert.Contains(t, auth.userAgent, "test-user")
	})
}

func TestRedditFetcher_TransformPost(t *testing.T) {
	rf := &RedditFetcher{}

	tests := []struct {
		name string
		post RedditPost
		want *models.FetchedData
	}{
		{
			name: "self post",
			post: RedditPost{
				Data: RedditPostData{
					ID:          "abc123",
					Title:       "Test Post",
					Selftext:    "This is the content",
					Author:      "testuser",
					CreatedUTC:  1704067200,
					Score:       42,
					NumComments: 10,
					Subreddit:   "golang",
					Permalink:   "/r/golang/comments/abc123/",
					IsSelf:      true,
				},
			},
			want: &models.FetchedData{
				Title:      "Test Post",
				Content:    "This is the content",
				URL:        "https://reddit.com/r/golang/comments/abc123/",
				AuthorName: "testuser",
				AuthorID:   "testuser",
				Tags:       []string{"golang"},
				Engagement: map[string]interface{}{
					"score":    42,
					"comments": 10,
				},
				RawData: map[string]interface{}{
					"id": "abc123",
					"is_self":   true,
				},
			},
		},
		{
			name: "link post",
			post: RedditPost{
				Data: RedditPostData{
					ID:          "xyz789",
					Title:       "Check this out",
					Selftext:    "",
					Author:      "linkposter",
					CreatedUTC:  1704067200,
					Score:       100,
					NumComments: 25,
					Subreddit:   "programming",
					Permalink:   "/r/programming/comments/xyz789/",
					URL:         "https://example.com/article",
					IsSelf:      false,
				},
			},
			want: &models.FetchedData{
				Title:      "Check this out",
				Content:    "",
				URL:        "https://reddit.com/r/programming/comments/xyz789/",
				AuthorName: "linkposter",
				AuthorID:   "linkposter",
				Tags:       []string{"programming"},
				Media:      []string{"https://example.com/article"},
				Engagement: map[string]interface{}{
					"score":    100,
					"comments": 25,
				},
				RawData: map[string]interface{}{
					"id": "xyz789",
					"is_self":   false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rf.transformPost(tt.post)

			assert.Equal(t, tt.want.Title, got.Title)
			assert.Equal(t, tt.want.Content, got.Content)
			assert.Equal(t, tt.want.URL, got.URL)
			assert.Equal(t, tt.want.AuthorName, got.AuthorName)
			assert.Equal(t, tt.want.AuthorID, got.AuthorID)
			assert.Equal(t, tt.want.Tags, got.Tags)
			assert.Equal(t, tt.want.Media, got.Media)
			assert.Equal(t, tt.want.Engagement, got.Engagement)
			assert.Equal(t, tt.want.RawData, got.RawData)

			// Check PublishedAt
			expectedTime := time.Unix(1704067200, 0)
			assert.Equal(t, expectedTime.Unix(), got.PublishedAt.Unix())
		})
	}
}

func TestRedditFetcher_DeduplicateResults(t *testing.T) {
	rf := &RedditFetcher{}

	input := []*models.FetchedData{
		{
			Title:   "First",
			RawData: map[string]interface{}{"id": "1"},
		},
		{
			Title:   "Second",
			RawData: map[string]interface{}{"id": "2"},
		},
		{
			Title:   "Duplicate of First",
			RawData: map[string]interface{}{"id": "1"},
		},
		{
			Title:   "No Reddit ID",
			RawData: map[string]interface{}{},
		},
	}

	result := rf.deduplicateResults(input)

	assert.Len(t, result, 3)
	assert.Equal(t, "First", result[0].Title)
	assert.Equal(t, "Second", result[1].Title)
	assert.Equal(t, "No Reddit ID", result[2].Title)
}
