package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/gotogether/backend/internal/domain"
	"github.com/gotogether/backend/internal/testutil"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_Success(t *testing.T) {
	repo := &testutil.MockUserRepo{}
	svc := NewAuthService(repo, "test-secret")

	resp, err := svc.Register(context.Background(), RegisterInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
		Password:    "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.User.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", resp.User.Email)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := &testutil.MockUserRepo{
		CreateFn: func(ctx context.Context, user *domain.User) error {
			return domain.ErrAlreadyExists
		},
	}
	svc := NewAuthService(repo, "test-secret")

	_, err := svc.Register(context.Background(), RegisterInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
		Password:    "password123",
	})
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	repo := &testutil.MockUserRepo{}
	svc := NewAuthService(repo, "test-secret")

	tests := []struct {
		name  string
		input RegisterInput
	}{
		{"empty email", RegisterInput{Email: "", DisplayName: "Test", Password: "password123"}},
		{"empty password", RegisterInput{Email: "test@test.com", DisplayName: "Test", Password: ""}},
		{"empty name", RegisterInput{Email: "test@test.com", DisplayName: "", Password: "password123"}},
		{"short password", RegisterInput{Email: "test@test.com", DisplayName: "Test", Password: "12345"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Register(context.Background(), tc.input)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userID := uuid.New()

	repo := &testutil.MockUserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{
				ID:           userID,
				Email:        email,
				DisplayName:  "Test User",
				PasswordHash: string(hash),
			}, nil
		},
	}
	svc := NewAuthService(repo, "test-secret")

	resp, err := svc.Login(context.Background(), LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	repo := &testutil.MockUserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{
				ID:           uuid.New(),
				Email:        email,
				PasswordHash: string(hash),
			}, nil
		},
	}
	svc := NewAuthService(repo, "test-secret")

	_, err := svc.Login(context.Background(), LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &testutil.MockUserRepo{}
	svc := NewAuthService(repo, "test-secret")

	_, err := svc.Login(context.Background(), LoginInput{
		Email:    "nonexistent@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestValidateToken_Roundtrip(t *testing.T) {
	repo := &testutil.MockUserRepo{}
	svc := NewAuthService(repo, "test-secret")

	userID := uuid.New()
	token, err := svc.generateToken(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsedID, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsedID != userID {
		t.Fatalf("expected %s, got %s", userID, parsedID)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	repo := &testutil.MockUserRepo{}
	svc := NewAuthService(repo, "test-secret")

	_, err := svc.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	repo := &testutil.MockUserRepo{}
	svc1 := NewAuthService(repo, "secret-1")
	svc2 := NewAuthService(repo, "secret-2")

	token, _ := svc1.generateToken(uuid.New())

	_, err := svc2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for token signed with different secret")
	}
}
