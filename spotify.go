package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type SpotifyTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}
type spotifyTrack struct {
	Name    string `json:"name"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
	Album struct {
		Name   string `json:"name"`
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
	} `json:"album"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

type SpotifyTrackDTO struct {
	Name   string `json:"name"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Image  string `json:"image"`
	URL    string `json:"url"`
}

type SpotifyTopTracksResponse struct {
	Items []spotifyTrack `json:"items"`
}

func getSpotifyAccessToken() (string, error) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	refreshToken := os.Getenv("SPOTIFY_REFRESH_TOKEN")

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest(
		"POST",
		"https://accounts.spotify.com/api/token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", err
	}

	auth := base64.StdEncoding.EncodeToString(
		[]byte(clientID + ":" + clientSecret),
	)

	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"spotify token request failed: %s",
			string(body),
		)
	}

	var token SpotifyTokenResponse

	err = json.Unmarshal(body, &token)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func getTopTracks() ([]SpotifyTrackDTO, error) {
	accessToken, err := getSpotifyAccessToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"GET",
		"https://api.spotify.com/v1/me/top/tracks?time_range=short_term&limit=10",
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+accessToken,
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"spotify top tracks request failed: %s",
			string(body),
		)
	}

	var result SpotifyTopTracksResponse

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	tracks := make([]SpotifyTrackDTO, 0, len(result.Items))

	for _, track := range result.Items {
		artistNames := make([]string, 0, len(track.Artists))

		for _, artist := range track.Artists {
			artistNames = append(artistNames, artist.Name)
		}

		image := ""

		if len(track.Album.Images) > 0 {
			image = track.Album.Images[0].URL
		}

		tracks = append(tracks, SpotifyTrackDTO{
			Name:   track.Name,
			Artist: strings.Join(artistNames, ", "),
			Album:  track.Album.Name,
			Image:  image,
			URL:    track.ExternalURLs.Spotify,
		})
	}

	return tracks, nil
}
