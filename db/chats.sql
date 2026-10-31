-- name: GetChatByEventID :one
SELECT id, event_id, is_active, last_message_at, created_at, updated_at
FROM chat.event_chats
WHERE event_id = $1
LIMIT 1;

-- name: CreateChat :one
INSERT INTO chat.event_chats (event_id, is_active)
VALUES ($1, $2)
RETURNING id, event_id, is_active, last_message_at, created_at, updated_at;

-- name: UpdateChatLastMessageAt :exec
UPDATE chat.event_chats
SET last_message_at = $2, updated_at = NOW()
WHERE id = $1;
