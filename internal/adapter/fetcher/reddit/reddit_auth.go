package reddit

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type RedditAuth struct {
	clientID     string
	clientSecret string
	userAgent    string
	accessToken  string
	expiresAt    time.Time
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func NewRedditAuth(clientID, clientSecret, username string) *RedditAuth {
	userAgent := fmt.Sprintf("golang:feedle-batch:v1.0.0 (by /u/%s)", username)
	if username == "" {
		userAgent = "golang:feedle-batch:v1.0.0"
	}

	return &RedditAuth{
		clientID:     clientID,
		clientSecret: clientSecret,
		userAgent:    userAgent,
	}
}

func (ra *RedditAuth) GetAccessToken() (string, error) {
	// トークンがまだ有効な場合は再利用
	if ra.accessToken != "" && time.Now().Before(ra.expiresAt) {
		return ra.accessToken, nil
	}

	// 新しいトークンを取得
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	// Basic認証
	auth := base64.StdEncoding.EncodeToString([]byte(ra.clientID + ":" + ra.clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("User-Agent", ra.userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	ra.accessToken = tokenResp.AccessToken
	ra.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second) // 1分前に期限切れとする

	return ra.accessToken, nil
}
