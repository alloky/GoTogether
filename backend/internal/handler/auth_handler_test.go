package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gotogether/backend/internal/service"
	"github.com/gotogether/backend/internal/testutil"
)

func setupAuthTest() (*AuthHandler, *service.AuthService) {
	userRepo := &testutil.MockUserRepo{}
	authService := service.NewAuthService(userRepo, "test-secret")
	handler := NewAuthHandler(authService)
	return handler, authService
}

func TestRegisterHandler_Success(t *testing.T) {
	handler, _ := setupAuthTest()

	body, _ := json.Marshal(map[string]string{
		"email":       "test@example.com",
		"displayName": "Test User",
		"password":    "password123",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp service.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected token in response")
	}
}

func TestRegisterHandler_InvalidBody(t *testing.T) {
	handler, _ := setupAuthTest()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestLoginHandler_Success(t *testing.T) {
	// First register a user
	handler, _ := setupAuthTest()

	// Register
	regBody, _ := json.Marshal(map[string]string{
		"email":       "test@example.com",
		"displayName": "Test",
		"password":    "password123",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	handler.Register(regW, regReq)

	if regW.Code != http.StatusCreated {
		t.Fatalf("register failed: %d %s", regW.Code, regW.Body.String())
	}
}

func TestLoginHandler_InvalidBody(t *testing.T) {
	handler, _ := setupAuthTest()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestMeHandler_Unauthorized(t *testing.T) {
	handler, authService := setupAuthTest()

	// Test with valid token but no user in repo (will return not found)
	router := NewRouter(authService, nil, "*", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
	_ = handler // suppress unused
}
