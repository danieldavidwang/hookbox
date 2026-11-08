package main

import (
	"bytes"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func setupTestDatabase(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hookbox.db")
	if err := initDatabase(path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		db = nil
	})
	return path
}

func sampleRequest() CapturedRequest {
	return CapturedRequest{
		Method:     http.MethodPost,
		Path:       "/payment",
		Query:      url.Values{"tag": {"first", "second"}, "empty": {""}},
		Headers:    http.Header{"Content-Type": {"application/octet-stream"}, "X-Event": {"one", "two"}},
		Body:       []byte{0, 1, 255, 'h', 'i'},
		ReceivedAt: time.Date(2026, time.October, 12, 19, 30, 0, 123456789, time.FixedZone("offset", -4*60*60)),
	}
}

func assertRequest(t *testing.T, got, want CapturedRequest) {
	t.Helper()
	if got.ID != want.ID || got.Method != want.Method || got.Path != want.Path ||
		!reflect.DeepEqual(got.Query, want.Query) || !reflect.DeepEqual(got.Headers, want.Headers) ||
		!bytes.Equal(got.Body, want.Body) || !got.ReceivedAt.Equal(want.ReceivedAt) {
		t.Fatalf("request = %#v, want %#v", got, want)
	}
}

func TestStoreRequest(t *testing.T) {
	setupTestDatabase(t)
	want := sampleRequest()
	want.ID = 999 // The caller's ID must not override SQLite's generated ID.
	got, err := storeRequest(want)
	if err != nil {
		t.Fatal(err)
	}
	want.ID = 1
	assertRequest(t, got, want)

	empty := sampleRequest()
	empty.Body = nil
	got, err = storeRequest(empty)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 2 || got.Body == nil || len(got.Body) != 0 {
		t.Fatalf("stored empty request = %#v", got)
	}
	var bodyType string
	var bodyLength int
	if err := db.QueryRow("SELECT typeof(body), length(body) FROM requests WHERE id = ?", got.ID).Scan(&bodyType, &bodyLength); err != nil {
		t.Fatal(err)
	}
	if bodyType != "blob" || bodyLength != 0 {
		t.Fatalf("empty body type/length = %s/%d, want blob/0", bodyType, bodyLength)
	}
}

func TestGetRequests(t *testing.T) {
	setupTestDatabase(t)
	requests, err := getRequests()
	if err != nil {
		t.Fatal(err)
	}
	if requests == nil || len(requests) != 0 {
		t.Fatalf("empty list = %#v, want nonnil empty slice", requests)
	}
	var want []CapturedRequest
	for _, path := range []string{"/second", "/first"} {
		request := sampleRequest()
		request.Path = path
		stored, err := storeRequest(request)
		if err != nil {
			t.Fatal(err)
		}
		want = append(want, stored)
	}
	requests, err = getRequests()
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != len(want) {
		t.Fatalf("list length = %d, want %d", len(requests), len(want))
	}
	for i := range want {
		assertRequest(t, requests[i], want[i])
	}
}

func TestGetRequest(t *testing.T) {
	setupTestDatabase(t)
	want, err := storeRequest(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	got, found, err := getRequest(want.ID)
	if err != nil || !found {
		t.Fatalf("getRequest: found=%v err=%v", found, err)
	}
	assertRequest(t, got, want)
}

func TestRequestsPersistAfterReopen(t *testing.T) {
	path := setupTestDatabase(t)
	want, err := storeRequest(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := initDatabase(path); err != nil {
		t.Fatal(err)
	}
	got, found, err := getRequest(want.ID)
	if err != nil || !found {
		t.Fatalf("getRequest after reopen: found=%v err=%v", found, err)
	}
	assertRequest(t, got, want)
	next, err := storeRequest(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	if next.ID != want.ID+1 {
		t.Fatalf("ID after reopen = %d, want %d", next.ID, want.ID+1)
	}
}

func TestGetRequestMissing(t *testing.T) {
	setupTestDatabase(t)
	got, found, err := getRequest(123)
	if err != nil || found || !reflect.DeepEqual(got, CapturedRequest{}) {
		t.Fatalf("missing request: got=%#v found=%v err=%v", got, found, err)
	}
}
