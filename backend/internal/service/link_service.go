package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gotogether/backend/internal/domain"
)

type LinkService struct {
	userRepo    domain.UserRepository
	lcRepo      domain.LinkCodeRepository
	authService *AuthService
	emailSender EmailSender
}

func NewLinkService(userRepo domain.UserRepository, lcRepo domain.LinkCodeRepository, authService *AuthService, emailSender EmailSender) *LinkService {
	return &LinkService{
		userRepo:    userRepo,
		lcRepo:      lcRepo,
		authService: authService,
		emailSender: emailSender,
	}
}

// InitiateLinkFromBot generates an OTP and sends it to the web user's email.
func (s *LinkService) InitiateLinkFromBot(ctx context.Context, telegramID int64, email string) error {
	// Check if this telegram ID is already linked
	existing, err := s.userRepo.FindByTelegramID(ctx, telegramID)
	if err == nil && existing != nil {
		return fmt.Errorf("%w: this Telegram account is already linked to %s", domain.ErrAlreadyExists, existing.Email)
	}

	// Find the web user by email
	webUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("%w: no account found with email %s", domain.ErrNotFound, email)
		}
		return err
	}

	// Check if the web user already has a different telegram linked
	if webUser.TelegramID != nil {
		return fmt.Errorf("%w: this web account is already linked to a Telegram account", domain.ErrAlreadyExists)
	}

	// Generate 6-digit code
	code, err := generateOTP()
	if err != nil {
		return fmt.Errorf("generating OTP: %w", err)
	}

	lc := &domain.LinkCode{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := s.lcRepo.Create(ctx, lc); err != nil {
		return err
	}

	if err := s.emailSender.SendLinkCode(email, code); err != nil {
		return fmt.Errorf("sending link code email: %w", err)
	}

	return nil
}

// ConfirmLinkFromBot validates the OTP and merges the bot shadow account into the web account.
func (s *LinkService) ConfirmLinkFromBot(ctx context.Context, telegramID int64, email, code string) (string, error) {
	// Validate the OTP
	lc, err := s.lcRepo.FindValid(ctx, email, code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", fmt.Errorf("%w: invalid or expired code", domain.ErrBadRequest)
		}
		return "", err
	}

	if err := s.lcRepo.MarkUsed(ctx, lc.ID); err != nil {
		return "", err
	}

	// Find the web user
	webUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	// Find the bot's shadow account
	shadowEmail := fmt.Sprintf("tg_%d@telegram.local", telegramID)
	shadowUser, err := s.userRepo.FindByEmail(ctx, shadowEmail)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return "", err
	}

	// Merge shadow -> web user if shadow exists
	if shadowUser != nil {
		if err := s.userRepo.MergeUsers(ctx, shadowUser.ID, webUser.ID); err != nil {
			return "", fmt.Errorf("merging accounts: %w", err)
		}
	}

	// Set telegram_id on the web user
	if err := s.userRepo.UpdateTelegramID(ctx, webUser.ID, telegramID); err != nil {
		return "", err
	}

	// Generate a new JWT for the web user
	token, err := s.authService.generateToken(webUser.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// LinkTelegramUsername sets the telegram username on a web user's profile.
func (s *LinkService) LinkTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error {
	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	if username == "" {
		return fmt.Errorf("%w: telegram username cannot be empty", domain.ErrBadRequest)
	}
	return s.userRepo.UpdateTelegramUsername(ctx, userID, username)
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
