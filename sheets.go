package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetMessage struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
	GifURL    string `json:"gif_url"`
}

var sheetsService *sheets.Service

func initSheets() error {
	ctx := context.Background()

	credentialsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")

	service, err := sheets.NewService(
		ctx,
		option.WithCredentialsJSON([]byte(credentialsJSON)),
	)
	if err != nil {
		return err
	}

	sheetsService = service

	fmt.Println("Google Sheets connected")

	return nil
}

func getMessages() ([]SheetMessage, error) {
	spreadsheetID := os.Getenv("GOOGLE_SHEETS_ID")
	tab := os.Getenv("GOOGLE_SHEETS_TAB")

	readRange := fmt.Sprintf("%s!A2:E", tab)

	resp, err := sheetsService.Spreadsheets.Values.Get(
		spreadsheetID,
		readRange,
	).Do()
	if err != nil {
		return nil, err
	}

	messages := make([]SheetMessage, 0)

	for _, row := range resp.Values {
		if len(row) < 4 {
			continue
		}
		gifURL := ""
		if len(row) >= 5 {
			gifURL = fmt.Sprint(row[4])
		}
		id, err := strconv.ParseInt(fmt.Sprint(row[0]), 10, 64)
		if err != nil {
			continue
		}

		messages = append(messages, SheetMessage{
			ID:        id,
			Name:      fmt.Sprint(row[1]),
			Message:   fmt.Sprint(row[2]),
			CreatedAt: fmt.Sprint(row[3]),
			GifURL:    gifURL,
		})
	}

	return messages, nil
}

func addMessage(name string, message string, gifURL string) (SheetMessage, error) {
	spreadsheetID := os.Getenv("GOOGLE_SHEETS_ID")
	tab := os.Getenv("GOOGLE_SHEETS_TAB")

	// Get existing rows so we can generate the next ID.
	messages, err := getMessages()
	if err != nil {
		return SheetMessage{}, err
	}

	var nextID int64 = 1

	if len(messages) > 0 {
		nextID = messages[len(messages)-1].ID + 1
	}

	createdAt := time.Now().Format(time.RFC3339)

	values := [][]interface{}{
		{
			nextID,
			name,
			message,
			createdAt,
			gifURL,
		},
	}

	_, err = sheetsService.Spreadsheets.Values.Append(
		spreadsheetID,
		fmt.Sprintf("%s!A:E", tab),
		&sheets.ValueRange{
			Values: values,
		},
	).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()

	if err != nil {
		return SheetMessage{}, err
	}

	return SheetMessage{
		ID:        nextID,
		Name:      name,
		Message:   message,
		CreatedAt: createdAt,
		GifURL:    gifURL,
	}, nil
}
