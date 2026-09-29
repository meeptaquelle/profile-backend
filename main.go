package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	fmt.Println("Backend running on http://127.0.0.1:8080")

	http.ListenAndServe(":8080", nil)
}