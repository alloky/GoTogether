package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type LinkCode struct {
	ID        uuid.UUID
	Email     string
	Code      string
	ExpiresAt time.Time
	Used      bool
}

type LinkCodeRepository interface {
	Create(ctx context.Context, lc *LinkCode) error
	FindValid(ctx context.Context, email, code string) (*LinkCode, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}
