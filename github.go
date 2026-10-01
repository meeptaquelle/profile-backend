package main

import (
	"encoding/json"
	"net/http"
)

type GitHubProfile struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	HTMLURL     string `json:"html_url"`
}

type GitHubContribution struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
	Level int    `json:"level"`
}

type GitHubRepo struct {
	Name      string `json:"name"`
	FullName  string `json:"full_name"`
	HTMLURL   string `json:"html_url"`
	UpdatedAt string `json:"updated_at"`
}

type GitHubResponse struct {
	Profile GitHubProfile `json:"profile"`
	Repos   []GitHubRepo  `json:"repos"`
}

func getGitHubProfile(w http.ResponseWriter, r *http.Request) {
	profileResponse, err := http.Get(
		"https://api.github.com/users/meeptaquelle",
	)
	if err != nil {
		http.Error(w, "Failed to fetch GitHub profile", http.StatusInternalServerError)
		return
	}
	defer profileResponse.Body.Close()

	if profileResponse.StatusCode != http.StatusOK {
		http.Error(w, "GitHub API returned an error", profileResponse.StatusCode)
		return
	}

	var profile GitHubProfile

	if err := json.NewDecoder(profileResponse.Body).Decode(&profile); err != nil {
		http.Error(w, "Failed to decode GitHub profile", http.StatusInternalServerError)
		return
	}

	reposResponse, err := http.Get(
		"https://api.github.com/users/meeptaquelle/repos?sort=updated&per_page=5",
	)
	if err != nil {
		http.Error(w, "Failed to fetch GitHub repositories", http.StatusInternalServerError)
		return
	}
	defer reposResponse.Body.Close()

	if reposResponse.StatusCode != http.StatusOK {
		http.Error(w, "GitHub API returned an error", reposResponse.StatusCode)
		return
	}

	var repos []GitHubRepo

	if err := json.NewDecoder(reposResponse.Body).Decode(&repos); err != nil {
		http.Error(w, "Failed to decode GitHub repositories", http.StatusInternalServerError)
		return
	}

	response := GitHubResponse{
		Profile: profile,
		Repos:   repos,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

func getGitHubContributions(w http.ResponseWriter, r *http.Request) {
	response, err := http.Get(
		"https://github-contributions-api.jogruber.de/v4/meeptaquelle?y=last",
	)
	if err != nil {
		http.Error(w, "Failed to fetch GitHub contributions", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		http.Error(w, "GitHub contributions API returned an error", response.StatusCode)
		return
	}

	var data struct {
		Total         map[string]int       `json:"total"`
		Contributions []GitHubContribution `json:"contributions"`
	}

	if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
		http.Error(w, "Failed to decode GitHub contributions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(data)
}
