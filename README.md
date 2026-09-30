# HookBox

HookBox is a local-first webhook debugging tool for capturing, inspecting, and eventually replaying HTTP webhook requests.

The project is being built incrementally to explore practical backend concepts such as HTTP request handling, concurrency, persistence, serialization, retries, and webhook reliability.

## Current Features

- Starts a local HTTP server on port 8080
- Accepts arbitrary incoming HTTP requests
- Captures:
  - HTTP method
  - request path
  - query parameters
  - headers
  - request body
  - receive timestamp
- Limits request bodies to 1 MiB
- Assigns each captured request a unique in-memory ID
- Stores captured requests safely across concurrent handlers using a mutex
- Lists captured requests as lightweight JSON summaries
- Retrieves the full details of a captured request by ID
- Returns appropriate errors for malformed IDs, missing requests, and unsupported methods

## Getting Started

### Requirements

- Go 1.27+

### Run HookBox

```bash
go run .
```

HookBox listens on:

```text
http://localhost:8080
```

## Example Workflow

Capture a webhook-like request:

```bash
curl -X POST "http://localhost:8080/payment?test=true" \
  -H "Content-Type: application/json" \
  -H "X-Event-ID: abc123" \
  -d '{"event":"payment.succeeded","amount":2000}'
```

HookBox stores the request and assigns it an ID.

List captured requests:

```bash
curl http://localhost:8080/requests
```

Example response:

```json
[
  {
    "id": 1,
    "method": "POST",
    "path": "/payment",
    "received_at": "2026-09-29T21:00:00Z"
  }
]
```

Inspect the full request:

```bash
curl http://localhost:8080/requests/1
```

The detail response includes the request method, path, query parameters, headers, body, and receive timestamp.

## Current Architecture

```text
Incoming HTTP request
        |
        v
   HTTP handler
        |
        v
CapturedRequest snapshot
        |
        v
thread-safe in-memory store
        |
        +----> GET /requests
        |
        +----> GET /requests/{id}
```

The HTTP handlers work with HookBox's own `CapturedRequest` representation rather than keeping the live `http.Request`. This lets the request remain available after the original handler finishes.

## Current Limitation

Captured requests are currently stored only in memory.

If the HookBox process stops, the stored requests and ID counter are lost. Persistent storage is the next major milestone.

## Roadmap

Planned features include:

- SQLite persistence
- Automated HTTP and storage tests
- CLI commands for listing and inspecting requests
- Webhook replay
- Retry and exponential backoff support
- Timeout and failure simulation
- HMAC signature verification
- Public webhook ingestion through tunneling or deployment
- Docker support

## Status

HookBox is under active development.
