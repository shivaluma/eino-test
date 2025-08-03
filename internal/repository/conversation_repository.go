package repository

import (
	"context"

	"github.com/shivaluma/eino-agent/internal/database"
	"github.com/shivaluma/eino-agent/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ConversationRepository struct {
	db *database.DB
}

func NewConversationRepository(db *database.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(ctx context.Context, conversation *models.Conversation) error {
	query := `
		INSERT INTO conversations (user_id, title)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	return r.db.Pool.QueryRow(ctx, query, conversation.UserID, conversation.Title).
		Scan(&conversation.ID, &conversation.CreatedAt, &conversation.UpdatedAt)
}

func (r *ConversationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Conversation, error) {
	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM conversations
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []models.Conversation
	for rows.Next() {
		var conv models.Conversation
		err := rows.Scan(&conv.ID, &conv.UserID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}

	return conversations, rows.Err()
}

func (r *ConversationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM conversations
		WHERE id = $1`

	conversation := &models.Conversation{}
	err := r.db.Pool.QueryRow(ctx, query, id).
		Scan(&conversation.ID, &conversation.UserID, &conversation.Title, &conversation.CreatedAt, &conversation.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return conversation, nil
}

func (r *ConversationRepository) Update(ctx context.Context, conversation *models.Conversation) error {
	query := `
		UPDATE conversations
		SET title = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.Pool.QueryRow(ctx, query, conversation.ID, conversation.Title).
		Scan(&conversation.UpdatedAt)
}

func (r *ConversationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM conversations WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *ConversationRepository) CreateMessage(ctx context.Context, message *models.Message) error {
	query := `
		INSERT INTO messages (conversation_id, sender_id, sender_type, content, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	return r.db.Pool.QueryRow(ctx, query,
		message.ConversationID,
		message.SenderID,
		message.SenderType,
		message.Content,
		message.Metadata,
	).Scan(&message.ID, &message.CreatedAt)
}

func (r *ConversationRepository) GetMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]models.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, sender_type, content, metadata, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Pool.Query(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.SenderType,
			&msg.Content,
			&msg.Metadata,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

func (r *ConversationRepository) GetMessageCount(ctx context.Context, conversationID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE conversation_id = $1`

	var count int
	err := r.db.Pool.QueryRow(ctx, query, conversationID).Scan(&count)
	return count, err
}