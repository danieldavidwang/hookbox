package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RequestSummary struct {
	ID         int       `json:"id"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	ReceivedAt time.Time `json:"received_at"`
}

func main() {
	http.HandleFunc("/requests", handleListRequests)
	http.HandleFunc("/", handleRequest)

	fmt.Println("Hookbox listening on http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("server error:", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request!")
	fmt.Println("Method:", r.Method)
	fmt.Println("Path:", r.URL.Path)
	fmt.Println("Query:", r.URL.Query())
	fmt.Println("Headers:", r.Header)

	// Limit request bodies to 1 MB.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read request body", http.StatusBadRequest)
		return
	}

	fmt.Println("Body:", string(body))

	captured := CapturedRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		Query:      r.URL.Query(),
		Headers:    r.Header.Clone(),
		Body:       body,
		ReceivedAt: time.Now(),
	}

	captured = storeRequest(captured)

	fmt.Println("Captured request ID:", captured.ID)

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func handleListRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requests := getRequests()
	summaries := make([]RequestSummary, 0, len(requests))

	for _, request := range requests {
		summary := RequestSummary{
			ID:         request.ID,
			Method:     request.Method,
			Path:       request.Path,
			ReceivedAt: request.ReceivedAt,
		}

		summaries = append(summaries, summary)
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(summaries)
	if err != nil {
		fmt.Println("Error encoding requests:", err)
	}
}
