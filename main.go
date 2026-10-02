package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: failed to load .env: %v", err)
	}

	if err := initSheets(); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/api/health", healthHandler)

	http.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getMessagesHandler(w, r)

		case http.MethodPost:
			createMessageHandler(w, r)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/spotify/login", spotifyLoginHandler)
	http.HandleFunc("/api/spotify/callback", spotifyCallbackHandler)
	http.HandleFunc("/api/spotify/top-tracks", topTracksHandler)
	http.HandleFunc("/api/spotify/top-artists", getSpotifyTopArtists)
	http.HandleFunc("/api/github", getGitHubProfile)
	http.HandleFunc("/api/github/contributions", getGitHubContributions)
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	handler := enableCORS(http.DefaultServeMux)

	fmt.Printf("Backend running on port %s\n", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}

}
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigin := os.Getenv("FRONTEND_URL")

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func spotifyLoginHandler(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")

	url := "https://accounts.spotify.com/authorize" +
		"?client_id=" + clientID +
		"&response_type=code" +
		"&redirect_uri=" + redirectURI +
		"&scope=user-top-read"

	http.Redirect(w, r, url, http.StatusFound)
}
func spotifyCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest(
		"POST",
		"https://accounts.spotify.com/api/token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		http.Error(w, "Failed to create token request", http.StatusInternalServerError)
		return
	}

	auth := base64.StdEncoding.EncodeToString(
		[]byte(clientID + ":" + clientSecret),
	)

	req.Header.Set(
		"Authorization",
		"Basic "+auth,
	)

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to contact Spotify", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read Spotify response", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, string(body), resp.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Spotify authorization successful. Check your terminal."))
}

func topTracksHandler(w http.ResponseWriter, r *http.Request) {
	tracks, err := getTopTracks()

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(tracks)
}
