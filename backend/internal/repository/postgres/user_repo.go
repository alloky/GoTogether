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

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

const userColumns = `id, email, display_name, password_hash, created_at, telegram_id, telegram_username`

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.CreatedAt, &u.TelegramID, &u.TelegramUsername)
	return &u, err
}

func scanUsers(rows pgx.Rows) ([]domain.User, error) {
	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.CreatedAt, &u.TelegramID, &u.TelegramUsername); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	user.ID = uuid.New()
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (id, email, display_name, password_hash)
		 VALUES ($1, $2, $3, $4)
		 RETURNING created_at`,
		user.ID, user.Email, user.DisplayName, user.PasswordHash,
	).Scan(&user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: email already registered", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = $1`, email,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	return u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("finding user by id: %w", err)
	}
	return u, nil
}

func (r *UserRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.db.Query(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = ANY($1)`, ids,
	)
	if err != nil {
		return nil, fmt.Errorf("finding users by ids: %w", err)
	}
	defer rows.Close()
	return scanUsers(rows)
}

func (r *UserRepo) FindByEmails(ctx context.Context, emails []string) ([]domain.User, error) {
	if len(emails) == 0 {
		return nil, nil
	}
	rows, err := r.db.Query(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = ANY($1)`, emails,
	)
	if err != nil {
		return nil, fmt.Errorf("finding users by emails: %w", err)
	}
	defer rows.Close()
	return scanUsers(rows)
}

func (r *UserRepo) SearchByName(ctx context.Context, query string, limit int) ([]domain.User, error) {
	if query == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	rows, err := r.db.Query(ctx,
		`SELECT `+userColumns+`
		 FROM users
		 WHERE display_name ILIKE '%' || $1 || '%'
		 ORDER BY display_name
		 LIMIT $2`,
		query, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("searching users by name: %w", err)
	}
	defer rows.Close()
	return scanUsers(rows)
}

func (r *UserRepo) FindByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE telegram_id = $1`, telegramID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("finding user by telegram_id: %w", err)
	}
	return u, nil
}

func (r *UserRepo) UpdateTelegramID(ctx context.Context, userID uuid.UUID, telegramID int64) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE users SET telegram_id = $1 WHERE id = $2`, telegramID, userID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: telegram account already linked to another user", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("updating telegram_id: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepo) UpdateTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE users SET telegram_username = $1 WHERE id = $2`, username, userID,
	)
	if err != nil {
		return fmt.Errorf("updating telegram_username: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepo) MergeUsers(ctx context.Context, fromID, toID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin merge tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Transfer meeting ownership
	if _, err := tx.Exec(ctx,
		`UPDATE meetings SET organizer_id = $1 WHERE organizer_id = $2`, toID, fromID,
	); err != nil {
		return fmt.Errorf("transferring meetings: %w", err)
	}

	// Transfer participants (skip duplicates)
	if _, err := tx.Exec(ctx,
		`UPDATE participants SET user_id = $1
		 WHERE user_id = $2
		   AND meeting_id NOT IN (SELECT meeting_id FROM participants WHERE user_id = $1)`,
		toID, fromID,
	); err != nil {
		return fmt.Errorf("transferring participants: %w", err)
	}
	// Remove remaining duplicates
	if _, err := tx.Exec(ctx,
		`DELETE FROM participants WHERE user_id = $1`, fromID,
	); err != nil {
		return fmt.Errorf("cleaning up participants: %w", err)
	}

	// Transfer votes (skip duplicates)
	if _, err := tx.Exec(ctx,
		`UPDATE votes SET user_id = $1
		 WHERE user_id = $2
		   AND time_slot_id NOT IN (SELECT time_slot_id FROM votes WHERE user_id = $1)`,
		toID, fromID,
	); err != nil {
		return fmt.Errorf("transferring votes: %w", err)
	}
	// Remove remaining duplicates
	if _, err := tx.Exec(ctx,
		`DELETE FROM votes WHERE user_id = $1`, fromID,
	); err != nil {
		return fmt.Errorf("cleaning up votes: %w", err)
	}

	// Delete the shadow user
	if _, err := tx.Exec(ctx,
		`DELETE FROM users WHERE id = $1`, fromID,
	); err != nil {
		return fmt.Errorf("deleting shadow user: %w", err)
	}

	return tx.Commit(ctx)
}
