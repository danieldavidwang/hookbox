package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

func storeRequest(captured CapturedRequest) (CapturedRequest, error) {
	query, err := json.Marshal(captured.Query)
	if err != nil {
		return CapturedRequest{}, err
	}
	headers, err := json.Marshal(captured.Headers)
	if err != nil {
		return CapturedRequest{}, err
	}
	if captured.Body == nil {
		captured.Body = []byte{}
	}
	result, err := db.Exec(`INSERT INTO requests (method, path, query, headers, body, received_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		captured.Method, captured.Path, string(query), string(headers), captured.Body,
		captured.ReceivedAt.Format(time.RFC3339Nano))
	if err != nil {
		return CapturedRequest{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return CapturedRequest{}, err
	}
	if int64(int(id)) != id {
		return CapturedRequest{}, fmt.Errorf("request ID %d overflows int", id)
	}
	captured.ID = int(id)
	return captured, nil
}

func getRequests() ([]CapturedRequest, error) {
	rows, err := db.Query("SELECT id, method, path, query, headers, body, received_at FROM requests ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]CapturedRequest, 0)
	for rows.Next() {
		var request CapturedRequest
		var query, headers, receivedAt string
		if err := rows.Scan(&request.ID, &request.Method, &request.Path, &query, &headers, &request.Body, &receivedAt); err != nil {
			return nil, err
		}
		if err := decodeRequest(&request, query, headers, receivedAt); err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return requests, nil
}

func getRequest(id int) (CapturedRequest, bool, error) {
	var request CapturedRequest
	var query, headers, receivedAt string
	err := db.QueryRow("SELECT id, method, path, query, headers, body, received_at FROM requests WHERE id = ?", id).
		Scan(&request.ID, &request.Method, &request.Path, &query, &headers, &request.Body, &receivedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return CapturedRequest{}, false, nil
	}
	if err != nil {
		return CapturedRequest{}, false, err
	}
	if err := decodeRequest(&request, query, headers, receivedAt); err != nil {
		return CapturedRequest{}, false, err
	}
	return request, true, nil
}

func decodeRequest(request *CapturedRequest, query, headers, receivedAt string) error {
	if err := json.Unmarshal([]byte(query), &request.Query); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(headers), &request.Headers); err != nil {
		return err
	}
	var err error
	request.ReceivedAt, err = time.Parse(time.RFC3339Nano, receivedAt)
	return err
}
