-- name: CreateMessage :one
INSERT INTO chat.messages (chat_id, user_id, sender_username, reply_to_id, type, body)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, chat_id, user_id, sender_username, reply_to_id, type, body, is_deleted, created_at, edited_at;

-- name: GetMessagesByChatID :many
SELECT id, chat_id, user_id, sender_username, reply_to_id, type, body, is_deleted, created_at, edited_at
FROM chat.messages
WHERE chat_id = $1
  AND is_deleted = FALSE
  AND ($2::timestamptz IS NULL OR created_at < $2)
ORDER BY created_at DESC
LIMIT $3;

-- name: SoftDeleteMessage :exec
UPDATE chat.messages
SET is_deleted = TRUE
WHERE id = $1 AND user_id = $2;

-- name: GetMessageByID :one
SELECT id, chat_id, user_id, sender_username, reply_to_id, type, body, is_deleted, created_at, edited_at
FROM chat.messages
WHERE id = $1
LIMIT 1;
