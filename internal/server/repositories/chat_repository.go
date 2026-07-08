package repositories

import (
	"context"
	"database/sql"
	"time"
)

type ChatSession struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatMessage struct {
	ID        int       `json:"id"`
	SessionID int       `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// ListSessions returns all chat sessions for a user, newest first
func (r *ChatRepository) ListSessions(ctx context.Context, userID int) ([]ChatSession, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, title, created_at, updated_at FROM chat_sessions WHERE user_id = ? ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []ChatSession
	for rows.Next() {
		var s ChatSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.Title, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// CreateSession creates a new chat session
func (r *ChatRepository) CreateSession(ctx context.Context, userID int, title string) (int, error) {
	res, err := r.db.ExecContext(ctx, `INSERT INTO chat_sessions (user_id, title) VALUES (?, ?)`, userID, title)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

// DeleteSession deletes a chat session and its messages (CASCADE)
func (r *ChatRepository) DeleteSession(ctx context.Context, sessionID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM chat_sessions WHERE id = ?`, sessionID)
	return err
}

// UpdateSessionTitle updates the title of a session
func (r *ChatRepository) UpdateSessionTitle(ctx context.Context, sessionID int, title string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE chat_sessions SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, title, sessionID)
	return err
}

// TouchSession updates the updated_at timestamp
func (r *ChatRepository) TouchSession(ctx context.Context, sessionID int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE chat_sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, sessionID)
	return err
}

// GetMessages returns all messages for a session, ordered chronologically
func (r *ChatRepository) GetMessages(ctx context.Context, sessionID int) ([]ChatMessage, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, session_id, role, content, created_at FROM chat_messages WHERE session_id = ? ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// AddMessage adds a message to a session
func (r *ChatRepository) AddMessage(ctx context.Context, sessionID int, role, content string) (int, error) {
	res, err := r.db.ExecContext(ctx, `INSERT INTO chat_messages (session_id, role, content) VALUES (?, ?, ?)`, sessionID, role, content)
	if err != nil {
		return 0, err
	}
	_ = r.TouchSession(ctx, sessionID)
	id, err := res.LastInsertId()
	return int(id), err
}
