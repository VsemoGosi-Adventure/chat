package service

import (
	"context"
	"time"

	"github.com/VsemoGosi-Adventure/chat/internal/db"
	"github.com/VsemoGosi-Adventure/chat/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ChatService struct {
	queries db.Querier
}

func NewChatService(queries db.Querier) *ChatService {
	return &ChatService{
		queries: queries,
	}
}

func (s *ChatService) GetOrCreateChat(ctx context.Context, eventID uuid.UUID) (*domain.Chat, error) {
	pgEventID := pgtype.UUID{Bytes: eventID, Valid: true}

	chatRow, err := s.queries.GetChatByEventID(ctx, pgEventID)
	if err == nil {
		return s.toDomainChat(chatRow), nil
	}

	chatRow, err = s.queries.CreateChat(ctx, db.CreateChatParams{
		EventID:  pgEventID,
		IsActive: true,
	})
	if err != nil {
		return nil, err
	}

	return s.toDomainChat(chatRow), nil
}

func (s *ChatService) CreateMessage(ctx context.Context, chatID, userID uuid.UUID, username string, msg domain.IncomingMessage) (*domain.Message, error) {
	pgChatID := pgtype.UUID{Bytes: chatID, Valid: true}
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}

	var pgReplyToID pgtype.UUID
	if msg.ReplyToID != nil {
		pgReplyToID = pgtype.UUID{Bytes: *msg.ReplyToID, Valid: true}
	}

	messageRow, err := s.queries.CreateMessage(ctx, db.CreateMessageParams{
		ChatID:         pgChatID,
		UserID:         pgUserID,
		SenderUsername: username,
		ReplyToID:      pgReplyToID,
		Type:           string(msg.Type),
		Body:           msg.Body,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.queries.UpdateChatLastMessageAt(ctx, db.UpdateChatLastMessageAtParams{
		ID:            pgChatID,
		LastMessageAt: pgtype.Timestamptz{Time: now, Valid: true},
	}); err != nil {
		return nil, err
	}

	return s.toDomainMessage(messageRow), nil
}

func (s *ChatService) GetMessages(ctx context.Context, chatID uuid.UUID, cursor *time.Time, limit int) ([]domain.Message, error) {
	pgChatID := pgtype.UUID{Bytes: chatID, Valid: true}

	var pgCursor pgtype.Timestamptz
	if cursor != nil {
		pgCursor = pgtype.Timestamptz{Time: *cursor, Valid: true}
	}

	rows, err := s.queries.GetMessagesByChatID(ctx, db.GetMessagesByChatIDParams{
		ChatID:  pgChatID,
		Column2: pgCursor,
		Limit:   int32(limit),
	})
	if err != nil {
		return nil, err
	}

	messages := make([]domain.Message, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, *s.toDomainMessage(row))
	}

	return messages, nil
}

func (s *ChatService) DeleteMessage(ctx context.Context, messageID, userID uuid.UUID) error {
	pgMessageID := pgtype.UUID{Bytes: messageID, Valid: true}
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}

	return s.queries.SoftDeleteMessage(ctx, db.SoftDeleteMessageParams{
		ID:     pgMessageID,
		UserID: pgUserID,
	})
}

func (s *ChatService) toDomainChat(row db.ChatEventChat) *domain.Chat {
	chat := &domain.Chat{
		ID:       uuid.UUID(row.ID.Bytes),
		EventID:  uuid.UUID(row.EventID.Bytes),
		IsActive: row.IsActive,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}

	if row.LastMessageAt.Valid {
		chat.LastMessageAt = &row.LastMessageAt.Time
	}

	return chat
}

func (s *ChatService) toDomainMessage(row db.ChatMessage) *domain.Message {
	msg := &domain.Message{
		ID:             uuid.UUID(row.ID.Bytes),
		ChatID:         uuid.UUID(row.ChatID.Bytes),
		UserID:         uuid.UUID(row.UserID.Bytes),
		SenderUsername: row.SenderUsername,
		Type:           domain.MessageType(row.Type),
		Body:           row.Body,
		IsDeleted:      row.IsDeleted,
		CreatedAt:      row.CreatedAt.Time,
	}

	if row.ReplyToID.Valid {
		replyID := uuid.UUID(row.ReplyToID.Bytes)
		msg.ReplyToID = &replyID
	}

	if row.EditedAt.Valid {
		msg.EditedAt = &row.EditedAt.Time
	}

	return msg
}
