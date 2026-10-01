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
	rows, err := db.Query(`
		SELECT id, name, message, created_at
		FROM messages
		ORDER BY created_at ASC
	`)

	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	messages := []Message{}

	for rows.Next() {
		var message Message

		err := rows.Scan(
			&message.ID,
			&message.Name,
			&message.Message,
			&message.CreatedAt,
		)

		if err != nil {
			http.Error(w, "Failed to read messages", http.StatusInternalServerError)
			return
		}

		messages = append(messages, message)
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(messages)
}
func createMessageHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name    string `json:"name"`
		Message string `json:"message"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
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

	result, err := db.Exec(`
		INSERT INTO messages (name, message)
		VALUES (?, ?)
	`, input.Name, input.Message)

	if err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get message ID", http.StatusInternalServerError)
		return
	}

	var message Message

	err = db.QueryRow(`
		SELECT id, name, message, created_at
		FROM messages
		WHERE id = ?
	`, id).Scan(
		&message.ID,
		&message.Name,
		&message.Message,
		&message.CreatedAt,
	)

	if err != nil {
		http.Error(w, "Failed to fetch created message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(message)
}
