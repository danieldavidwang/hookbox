package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// RequestSummary is the lightweight representation returned by GET /requests.
// Listing many requests should not send every stored header, query value, and body.
type RequestSummary struct {
	ID         int       `json:"id"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	ReceivedAt time.Time `json:"received_at"`
}

// RequestDetail is the full representation returned by GET /requests/{id}.
// Body is exposed as a string so captured text/JSON is readable in the API response.
type RequestDetail struct {
	ID         int         `json:"id"`
	Method     string      `json:"method"`
	Path       string      `json:"path"`
	Query      url.Values  `json:"query"`
	Headers    http.Header `json:"headers"`
	Body       string      `json:"body"`
	ReceivedAt time.Time   `json:"received_at"`
}

func main() {
	// The specific /requests routes expose Hookbox's inspection API.
	// "/" remains the catch-all endpoint for incoming webhook traffic.
	http.HandleFunc("/requests", handleListRequests)
	http.HandleFunc("/requests/", handleGetRequest)
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

	// Step 1: bound the live request body before reading it into memory.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read request body", http.StatusBadRequest)
		return
	}

	fmt.Println("Body:", string(body))

	// Step 2: turn the live http.Request into Hookbox's own durable snapshot.
	// Clone the headers so the stored request does not share the live header map.
	captured := CapturedRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		Query:      r.URL.Query(),
		Headers:    r.Header.Clone(),
		Body:       body,
		ReceivedAt: time.Now(),
	}

	// Step 3: store the snapshot and receive the version with its assigned ID.
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

	// Read the stored requests, then project each one into a lightweight summary.
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

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Step 1: extract the identifier from "/requests/7" -> "7".
	idString := strings.TrimPrefix(r.URL.Path, "/requests/")

	// Step 2: distinguish a malformed ID (400) from a valid-but-missing ID (404).
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	request, found := getRequest(id)
	if !found {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	// Step 3: convert the internal stored request into the API representation.
	detail := RequestDetail{
		ID:         request.ID,
		Method:     request.Method,
		Path:       request.Path,
		Query:      request.Query,
		Headers:    request.Headers,
		Body:       string(request.Body),
		ReceivedAt: request.ReceivedAt,
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(detail)
	if err != nil {
		http.Error(w, "Could not encode response", http.StatusInternalServerError)
		return
	}
}
