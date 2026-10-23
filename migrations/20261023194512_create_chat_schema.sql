-- +goose Up
CREATE SCHEMA IF NOT EXISTS chat;

CREATE TABLE IF NOT EXISTS chat.event_chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_message_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chat.messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id UUID NOT NULL REFERENCES chat.event_chats(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    sender_username VARCHAR(50) NOT NULL,
    reply_to_id UUID NULL REFERENCES chat.messages(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'text',
    body TEXT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    edited_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_chat_created ON chat.messages (chat_id, created_at DESC);

-- +goose Down
SELECT 'down SQL query';
