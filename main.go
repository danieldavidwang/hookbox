package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type CapturedRequest struct {
	ID         int
	Method     string
	Path       string
	Query      url.Values
	Headers    http.Header
	Body       []byte
	ReceivedAt time.Time
}

var (
	capturedRequests []CapturedRequest
	nextID           = 1
	mu               sync.Mutex
)

func main() {
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

	// CapturedRequest is HookBox's stored snapshot of an HTTP request
	captured := CapturedRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		Query:      r.URL.Query(),
		Headers:    r.Header.Clone(),
		Body:       body,
		ReceivedAt: time.Now(),
	}

	// Store the request and get back the version with its assigned ID.
	captured = storeRequest(captured)

	fmt.Println("Captured request ID:", captured.ID)

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func storeRequest(captured CapturedRequest) CapturedRequest {
	mu.Lock()
	defer mu.Unlock()

	captured.ID = nextID
	nextID++

	capturedRequests = append(capturedRequests, captured)
	return captured
}
