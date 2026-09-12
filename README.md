# HookBox

HookBox is a local-first webhook debugging tool for capturing, inspecting, and eventually replaying HTTP webhook requests.

## Current Features

- Starts a local HTTP server
- Accepts incoming HTTP requests
- Displays the request method and path

## Getting Started

### Requirements

- Go 1.25+

### Run HookBox

```bash
go run .
```

By default, HookBox listens on:

```text
http://localhost:8080
```

Send a test request:

```bash
curl -X POST http://localhost:8080/test
```

HookBox will output:

```text
Received request!
Method: POST
Path: /test
```

## Roadmap

Planned features include:

- Request body and header inspection
- Webhook persistence
- CLI commands for listing and inspecting requests
- Webhook replay
- Retry and exponential backoff simulation
- HMAC signature verification
- Failure and latency simulation

## Status

HookBox is currently under active development.
