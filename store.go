package main

import (
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

func storeRequest(captured CapturedRequest) CapturedRequest {
	mu.Lock()
	defer mu.Unlock()

	captured.ID = nextID
	nextID++

	capturedRequests = append(capturedRequests, captured)

	return captured
}

func getRequests() []CapturedRequest {
	mu.Lock()
	defer mu.Unlock()

	requests := make([]CapturedRequest, len(capturedRequests))
	copy(requests, capturedRequests)

	return requests
}
