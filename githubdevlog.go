package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type GitHubCommit struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
	Author  string `json:"author"`
	Date    string `json:"date"`
	HTMLURL string `json:"url"`
}

type GitHubDevlog struct {
	Frontend []GitHubCommit `json:"frontend"`
	Backend  []GitHubCommit `json:"backend"`
}

type githubCommitResponse struct {
	SHA    string `json:"sha"`
	Commit struct {
		Author struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"author"`
		Message string `json:"message"`
	} `json:"commit"`
	HTMLURL string `json:"html_url"`
}

func fetchGitHubCommits(repo string, branch string, limit int) ([]GitHubCommit, error) {
	apiURL := "https://api.github.com/repos/" +
		repo +
		"/commits?sha=" +
		branch +
		"&per_page=" +
		strconv.Itoa(limit)

	response, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"GitHub API returned status %d",
			response.StatusCode,
		)
	}

	var commits []githubCommitResponse

	if err := json.NewDecoder(response.Body).Decode(&commits); err != nil {
		return nil, err
	}

	result := make([]GitHubCommit, 0, len(commits))

	for _, commit := range commits {
		result = append(result, GitHubCommit{
			SHA:     commit.SHA,
			Message: commit.Commit.Message,
			Author:  commit.Commit.Author.Name,
			Date:    commit.Commit.Author.Date,
			HTMLURL: commit.HTMLURL,
		})
	}

	return result, nil
}

func getGitHubDevlog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	frontend, err := fetchGitHubCommits(
		"meeptaquelle/meeptaquelle.github.io",
		"gh-pages",
		10,
	)

	if err != nil {
		http.Error(
			w,
			"Failed to fetch frontend commits: "+err.Error(),
			http.StatusBadGateway,
		)
		return
	}

	backend, err := fetchGitHubCommits(
		"meeptaquelle/profile-backend",
		"main",
		10,
	)

	if err != nil {
		http.Error(
			w,
			"Failed to fetch backend commits: "+err.Error(),
			http.StatusBadGateway,
		)
		return
	}

	response := GitHubDevlog{
		Frontend: frontend,
		Backend:  backend,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(
			w,
			"Failed to encode devlog response",
			http.StatusInternalServerError,
		)
	}
}
func getGithubCommits(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get("repo")

	if repo == "" {
		http.Error(w, "repo is required", http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/commits?per_page=10",
		repo,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "failed to fetch GitHub", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "GitHub API error", resp.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}
