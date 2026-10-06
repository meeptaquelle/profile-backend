package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type Message struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Message   string `json:"message"`
	GifURL    string `json:"gif_url"`
	CreatedAt string `json:"created_at"`
}

func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	messages, err := getMessages()
	if err != nil {
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func createMessageHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name    string `json:"name"`
		Message string `json:"message"`
		GifURL  string `json:"gif_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Message = strings.TrimSpace(input.Message)
	input.GifURL = strings.TrimSpace(input.GifURL)

	if input.Name == "" {
		input.Name = "Anonymous"
	}
	if len(input.Name) > 50 {
		input.Name = input.Name[:50]
	}
	if len(input.Message) > 500 {
		http.Error(w, "Message is too long", http.StatusBadRequest)
		return
	}

	// A message is valid if it has text, a GIF, or both.
	if input.Message == "" && input.GifURL == "" {
		http.Error(w, "Message or GIF is required", http.StatusBadRequest)
		return
	}

	if input.GifURL != "" && !isAllowedGifURL(input.GifURL) {
		http.Error(w, "GIF URL must be a direct link from Tenor or Giphy", http.StatusBadRequest)
		return
	}

	message, err := addMessage(input.Name, input.Message)
	if err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

/*
 * Allowlist of CDN hosts that serve direct GIF/WebP media files.
 *
 * We deliberately match on the *exact* hostname, not a suffix — a suffix
 * check would let "evil.com.media.tenor.com" through if DNS ever resolved
 * a wildcard there, and Tenor/Giphy don't use sub-subdomains anyway.
 */
var allowedGifHosts = map[string]struct{}{
	"media.tenor.com":  {},
	"c.tenor.com":      {},
	"media.giphy.com":  {},
	"i.giphy.com":      {},
	"media0.giphy.com": {},
	"media1.giphy.com": {},
	"media2.giphy.com": {},
	"media3.giphy.com": {},
	"media4.giphy.com": {},
}

func isAllowedGifURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "https" {
		return false
	}
	_, ok := allowedGifHosts[strings.ToLower(u.Hostname())]
	return ok
}
