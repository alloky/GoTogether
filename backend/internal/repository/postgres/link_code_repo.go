package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gotogether/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinkCodeRepo struct {
	db *pgxpool.Pool
}

func NewLinkCodeRepo(db *pgxpool.Pool) *LinkCodeRepo {
	return &LinkCodeRepo{db: db}
}

func (r *LinkCodeRepo) Create(ctx context.Context, lc *domain.LinkCode) error {
	lc.ID = uuid.New()
	_, err := r.db.Exec(ctx,
		`INSERT INTO link_codes (id, email, code, expires_at) VALUES ($1, $2, $3, $4)`,
		lc.ID, lc.Email, lc.Code, lc.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("inserting link_code: %w", err)
	}
	return nil
}

func (r *LinkCodeRepo) FindValid(ctx context.Context, email, code string) (*domain.LinkCode, error) {
	var lc domain.LinkCode
	err := r.db.QueryRow(ctx,
		`SELECT id, email, code, expires_at, used
		 FROM link_codes
		 WHERE email = $1 AND code = $2 AND used = false AND expires_at > now()
		 ORDER BY created_at DESC
		 LIMIT 1`,
		email, code,
	).Scan(&lc.ID, &lc.Email, &lc.Code, &lc.ExpiresAt, &lc.Used)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("finding link_code: %w", err)
	}
	return &lc, nil
}

func (r *LinkCodeRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE link_codes SET used = true WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("marking link_code used: %w", err)
	}
	return nil
}
