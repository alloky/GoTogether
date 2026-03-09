package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gotogether/tgbot/internal/apiclient"
)

type Manager struct {
	api       *apiclient.Client
	jwtSecret string
	tokens    sync.Map // int64 (telegramID) -> string (JWT)
}

func NewManager(api *apiclient.Client, jwtSecret string) *Manager {
	return &Manager{
		api:       api,
		jwtSecret: jwtSecret,
	}
}

// EnsureAuth returns a valid JWT token for the given Telegram user.
// Auto-registers or logs in as needed.
func (m *Manager) EnsureAuth(ctx context.Context, telegramID int64, firstName, lastName, username string) (string, error) {
	// Check cache first
	if tok, ok := m.tokens.Load(telegramID); ok {
		return tok.(string), nil
	}

	email := fmt.Sprintf("tg_%d@telegram.local", telegramID)
	password := m.derivePassword(telegramID)
	displayName := buildDisplayName(firstName, lastName, username)

	// Try login first (user may already exist)
	resp, loginErr := m.api.Login(ctx, email, password)
	if loginErr == nil {
		m.tokens.Store(telegramID, resp.Token)
		log.Printf("Telegram user %d logged in as %s", telegramID, email)
		return resp.Token, nil
	}

	// If login failed due to a network/connection error, don't bother trying register
	if isConnectionError(loginErr) {
		return "", fmt.Errorf("backend unavailable: %w", loginErr)
	}

	// Login failed (likely 401 — user doesn't exist yet) — try to register
	resp, err := m.api.Register(ctx, email, displayName, password)
	if err != nil {
		return "", fmt.Errorf("auth failed (login: %v, register: %v)", loginErr, err)
	}

	m.tokens.Store(telegramID, resp.Token)
	log.Printf("Telegram user %d registered as %s (%s)", telegramID, email, displayName)
	return resp.Token, nil
}

// GetToken returns cached token or empty string.
func (m *Manager) GetToken(telegramID int64) (string, bool) {
	tok, ok := m.tokens.Load(telegramID)
	if !ok {
		return "", false
	}
	return tok.(string), true
}

// InvalidateToken removes cached token so next call re-authenticates.
func (m *Manager) InvalidateToken(telegramID int64) {
	m.tokens.Delete(telegramID)
}

// RefreshToken forces a new login to get a fresh JWT.
func (m *Manager) RefreshToken(ctx context.Context, telegramID int64) (string, error) {
	m.tokens.Delete(telegramID)
	email := fmt.Sprintf("tg_%d@telegram.local", telegramID)
	password := m.derivePassword(telegramID)

	resp, err := m.api.Login(ctx, email, password)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	m.tokens.Store(telegramID, resp.Token)
	return resp.Token, nil
}

// isConnectionError checks if the error is a network-level failure (not an HTTP error response).
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "dial tcp") ||
		strings.Contains(msg, "do request:")
}

func (m *Manager) derivePassword(telegramID int64) string {
	h := hmac.New(sha256.New, []byte(m.jwtSecret))
	h.Write([]byte(fmt.Sprintf("tgbot:%d", telegramID)))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func buildDisplayName(firstName, lastName, username string) string {
	parts := []string{}
	if firstName != "" {
		parts = append(parts, firstName)
	}
	if lastName != "" {
		parts = append(parts, lastName)
	}
	if len(parts) > 0 {
		return strings.Join(parts, " ")
	}
	if username != "" {
		return username
	}
	return "Telegram User"
}
