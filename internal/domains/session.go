package domains

import "time"

type Session struct {
	UserID       int       `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SessionRepository interface {
	CreateOrUpdate(session *Session) error
	Delete(userID int) error
}
