package main

import (
	"encoding/json"
	"net/http"
)

type Message struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Message   string `json:"message"`
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
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		input.Name = "Anonymous"
	}

	if input.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
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
