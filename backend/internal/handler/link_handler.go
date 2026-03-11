package handler

import (
	"net/http"

	"github.com/gotogether/backend/internal/service"
)

type LinkHandler struct {
	linkService *service.LinkService
}

func NewLinkHandler(linkService *service.LinkService) *LinkHandler {
	return &LinkHandler{linkService: linkService}
}

// BotSecretMiddleware validates the X-Bot-Secret header for bot-to-backend calls.
func BotSecretMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Bot-Secret") != secret {
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid bot secret"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type initiateRequest struct {
	TelegramID int64  `json:"telegramId"`
	Email      string `json:"email"`
}

// InitiateFromBot handles POST /api/link/bot/initiate
func (h *LinkHandler) InitiateFromBot(w http.ResponseWriter, r *http.Request) {
	var req initiateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.TelegramID == 0 || req.Email == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "telegramId and email are required"})
		return
	}

	if err := h.linkService.InitiateLinkFromBot(r.Context(), req.TelegramID, req.Email); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "code_sent"})
}

type confirmRequest struct {
	TelegramID int64  `json:"telegramId"`
	Email      string `json:"email"`
	Code       string `json:"code"`
}

// ConfirmFromBot handles POST /api/link/bot/confirm
func (h *LinkHandler) ConfirmFromBot(w http.ResponseWriter, r *http.Request) {
	var req confirmRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.TelegramID == 0 || req.Email == "" || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "telegramId, email and code are required"})
		return
	}

	token, err := h.linkService.ConfirmLinkFromBot(r.Context(), req.TelegramID, req.Email, req.Code)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

type linkTelegramRequest struct {
	TelegramUsername string `json:"telegramUsername"`
}

// LinkTelegram handles POST /api/auth/link/telegram (JWT-protected, web-initiated)
func (h *LinkHandler) LinkTelegram(w http.ResponseWriter, r *http.Request) {
	var req linkTelegramRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	userID := getUserID(r)
	if err := h.linkService.LinkTelegramUsername(r.Context(), userID, req.TelegramUsername); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "linked"})
}
