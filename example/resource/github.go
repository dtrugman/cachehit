package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dtrugman/cachehit/internal"
)

type GithubUser struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Company   string `json:"company"`
	Location  string `json:"location"`
	Followers int    `json:"followers"`
}

type GithubUserRepository struct {
	client *http.Client
}

func NewGithubUserRepository() *GithubUserRepository {
	return &GithubUserRepository{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *GithubUserRepository) Get(ctx context.Context, username string) (GithubUser, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s", username)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return GithubUser{}, fmt.Errorf("new request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return GithubUser{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return GithubUser{}, internal.ErrNotFound
	} else if resp.StatusCode != http.StatusOK {
		return GithubUser{}, fmt.Errorf("status: %d", resp.StatusCode)
	}

	var user GithubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return GithubUser{}, fmt.Errorf("decode: %w", err)
	}

	return user, nil
}
