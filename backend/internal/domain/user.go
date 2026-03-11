package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID `json:"id"`
	Email            string    `json:"email"`
	DisplayName      string    `json:"displayName"`
	PasswordHash     string    `json:"-"`
	CreatedAt        time.Time `json:"createdAt"`
	TelegramID       *int64    `json:"telegramId,omitempty"`
	TelegramUsername *string   `json:"telegramUsername,omitempty"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]User, error)
	FindByEmails(ctx context.Context, emails []string) ([]User, error)
	SearchByName(ctx context.Context, query string, limit int) ([]User, error)
	FindByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	UpdateTelegramID(ctx context.Context, userID uuid.UUID, telegramID int64) error
	UpdateTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error
	MergeUsers(ctx context.Context, fromID, toID uuid.UUID) error
}
