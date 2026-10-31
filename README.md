# Chat Service

Real-time chat service for event participants with WebSocket support.

## Features

- **Event-based group chats**: Each event has its own isolated chat room
- **Real-time messaging**: WebSocket support for bidirectional communication
- **Message history**: Cursor-based pagination for efficient history retrieval
- **Soft delete**: Messages can be deleted without permanent removal
- **Reply support**: Messages can reference other messages
- **System messages**: Support for both user and system message types

## Tech Stack

- **Language**: Go 1.22+
- **Database**: PostgreSQL 16 (schema: `chat`)
- **Database Layer**: sqlc with pgx/v5 driver
- **Migrations**: Goose
- **Configuration**: caarlos0/env
- **WebSocket**: nhooyr.io/websocket
- **API Documentation**: Swagger/OpenAPI via swaggo/swag

## Architecture

Standard three-layer architecture:
- **Transport Layer**: HTTP handlers, middleware, WebSocket upgrade
- **Business Logic**: Service layer, WebSocket Hub/Client management
- **Data Layer**: PostgreSQL queries via sqlc-generated code

## API Documentation

### Swagger UI

Start the server and navigate to:
```
http://localhost:8080/swagger/
```

### Authentication

All endpoints require the `X-User-ID` header (UUID format). This header is injected by the API Gateway (Traefik) after authentication.

Example:
```bash
curl -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
     http://localhost:8080/api/v1/chats/{event_id}/messages
```

### Endpoints

#### REST API

**GET /api/v1/chats/{event_id}/messages**
- Get paginated message history for a chat
- Query parameters:
  - `cursor` (optional): RFC3339 timestamp for pagination
  - `limit` (optional): Number of messages (default: 50, max: 100)
- Returns: `{ messages: [...], has_more: bool }`

**GET /health**
- Health check endpoint
- Returns: `OK`

#### WebSocket

**GET /api/v1/chats/{event_id}/ws**
- Upgrade to WebSocket connection for real-time chat
- Incoming message format:
  ```json
  {
    "reply_to_id": "uuid (optional)",
    "type": "text",
    "body": "message content"
  }
  ```
- Outgoing message format:
  ```json
  {
    "event_type": "message.new",
    "id": "uuid",
    "chat_id": "uuid",
    "user_id": "uuid",
    "sender_username": "string",
    "reply_to_id": "uuid (optional)",
    "type": "text",
    "body": "message content",
    "is_deleted": false,
    "created_at": "2026-10-31T12:00:00Z",
    "edited_at": null
  }
  ```

## Configuration

Environment variables (see `internal/config/config.go`):

### Server
- `SERVER_PORT` (default: `8080`)
- `SERVER_READ_TIMEOUT` (default: `10s`)
- `SERVER_WRITE_TIMEOUT` (default: `10s`)
- `SERVER_SHUTDOWN_TIMEOUT` (default: `15s`)

### Database
- `DATABASE_DSN` (required): PostgreSQL connection string
- `DATABASE_MAX_CONNS` (default: `25`)
- `DATABASE_MIN_CONNS` (default: `5`)
- `DATABASE_MAX_CONN_LIFETIME` (default: `1h`)
- `DATABASE_MAX_CONN_IDLE_TIME` (default: `30m`)

### WebSocket
- `WS_PING_INTERVAL` (default: `30s`)
- `WS_PONG_WAIT` (default: `60s`)
- `WS_WRITE_WAIT` (default: `10s`)
- `WS_MAX_MESSAGE_SIZE` (default: `8192`)
- `WS_READ_BUFFER_SIZE` (default: `4096`)
- `WS_WRITE_BUFFER_SIZE` (default: `4096`)
- `WS_BROADCAST_BUFFER` (default: `256`)

## Development

### Prerequisites
- Go 1.22+
- PostgreSQL 16+
- sqlc CLI (for regenerating DB code)
- swag CLI (for regenerating API docs)

### Build
```bash
go build -o bin/chat ./cmd/chat
```

### Run
```bash
export DATABASE_DSN="postgres://user:pass@localhost:5432/adventure?sslmode=disable"
./bin/chat
```

### Regenerate Database Code
```bash
sqlc generate
```

### Regenerate API Documentation
```bash
swag init -g cmd/chat/main.go -o docs --parseDependency --parseInternal
```

## Database Schema

See `migrations/20261023194512_create_chat_schema.sql` for the complete schema.

### Tables
- `chat.event_chats`: One chat per event
- `chat.messages`: Messages with soft delete support

## Graceful Shutdown

The service handles `SIGTERM` and `SIGINT` gracefully:
1. Stops accepting new connections
2. Closes WebSocket Hub (disconnects all clients)
3. Drains active HTTP requests (with timeout)
4. Closes database connection pool

## License

MIT
