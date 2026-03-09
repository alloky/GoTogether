package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gotogether/backend/internal/domain"
	"github.com/gotogether/backend/internal/service"
	"github.com/gotogether/backend/internal/testutil"
)

func setupMeetingTest() (*service.AuthService, *service.MeetingService, *testutil.MockMeetingRepo, string) {
	userID := uuid.New()

	userRepo := &testutil.MockUserRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return &domain.User{ID: userID, Email: "test@test.com", DisplayName: "Test"}, nil
		},
	}
	meetingRepo := &testutil.MockMeetingRepo{}
	authService := service.NewAuthService(userRepo, "test-secret")
	meetingService := service.NewMeetingService(meetingRepo, userRepo)

	// Generate a valid token
	token := generateTestToken(authService, userID)

	return authService, meetingService, meetingRepo, token
}

func generateTestToken(authService *service.AuthService, userID uuid.UUID) string {
	resp, err := authService.Register(context.Background(), service.RegisterInput{
		Email:       "tokenuser@test.com",
		DisplayName: "Token User",
		Password:    "password123",
	})
	if err == nil {
		return resp.Token
	}
	return ""
}

func TestListMeetings_Empty(t *testing.T) {
	authService, meetingService, meetingRepo, _ := setupMeetingTest()
	meetingRepo.ListByUserFn = func(ctx context.Context, userID uuid.UUID) ([]domain.Meeting, error) {
		return nil, nil
	}

	router := NewRouter(authService, meetingService, "*", nil)

	// Register a user to get a valid token
	regBody, _ := json.Marshal(map[string]string{
		"email":       "listuser@test.com",
		"displayName": "List User",
		"password":    "password123",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)

	if regW.Code != http.StatusCreated {
		t.Fatalf("register failed: %d %s", regW.Code, regW.Body.String())
	}

	var authResp service.AuthResponse
	json.Unmarshal(regW.Body.Bytes(), &authResp)

	// Now list meetings
	req := httptest.NewRequest(http.MethodGet, "/api/meetings/", nil)
	req.Header.Set("Authorization", "Bearer "+authResp.Token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var meetings []domain.Meeting
	json.Unmarshal(w.Body.Bytes(), &meetings)
	if len(meetings) != 0 {
		t.Fatalf("expected 0 meetings, got %d", len(meetings))
	}
}

func TestListMeetings_Unauthorized(t *testing.T) {
	authService, meetingService, _, _ := setupMeetingTest()
	router := NewRouter(authService, meetingService, "*", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/meetings/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestCreateMeeting_Unauthorized(t *testing.T) {
	authService, meetingService, _, _ := setupMeetingTest()
	router := NewRouter(authService, meetingService, "*", nil)

	body, _ := json.Marshal(map[string]interface{}{
		"title": "Test Meeting",
		"timeSlots": []map[string]string{
			{"startTime": "2026-03-01T12:00:00Z", "endTime": "2026-03-01T13:00:00Z"},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/meetings/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}
