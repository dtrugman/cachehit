package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

func (r *GithubUserRepository) Get(ctx context.Context, username string) (GithubUser, bool) {
	url := fmt.Sprintf("https://api.github.com/users/%s", username)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return GithubUser{}, false
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return GithubUser{}, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GithubUser{}, false
	}

	var user GithubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return GithubUser{}, false
	}

	return user, true
}
