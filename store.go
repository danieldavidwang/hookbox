package main

import (
	"net/http"
	"net/url"
	"sync"
	"time"
)

// CapturedRequest is Hookbox's internal snapshot of an incoming HTTP request.
// Unlike *http.Request, it contains only the data we want to keep after the handler returns.
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

	// The HTTP server may run multiple handlers concurrently, so access to
	// capturedRequests and nextID must be synchronized.
	mu sync.Mutex
)

func storeRequest(captured CapturedRequest) CapturedRequest {
	mu.Lock()
	defer mu.Unlock()

	// Assign the ID while holding the lock so two concurrent requests
	// cannot observe and claim the same nextID.
	captured.ID = nextID
	nextID++

	capturedRequests = append(capturedRequests, captured)

	return captured
}

func getRequests() []CapturedRequest {
	mu.Lock()
	defer mu.Unlock()

	// Return a separate slice so callers do not receive the store's backing slice.
	requests := make([]CapturedRequest, len(capturedRequests))
	copy(requests, capturedRequests)

	return requests
}

func getRequest(id int) (CapturedRequest, bool) {
	mu.Lock()
	defer mu.Unlock()

	for _, request := range capturedRequests {
		if request.ID == id {
			return request, true
		}
	}

	// The bool lets callers distinguish "not found" from a real zero-value request.
	return CapturedRequest{}, false
}
