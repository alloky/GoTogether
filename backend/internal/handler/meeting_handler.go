package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gotogether/backend/internal/domain"
	"github.com/gotogether/backend/internal/service"
)

type MeetingHandler struct {
	meetingService *service.MeetingService
}

func NewMeetingHandler(meetingService *service.MeetingService) *MeetingHandler {
	return &MeetingHandler{meetingService: meetingService}
}

func (h *MeetingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.CreateMeetingInput
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	meeting, err := h.meetingService.Create(r.Context(), getUserID(r), input)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, meeting)
}

func (h *MeetingHandler) List(w http.ResponseWriter, r *http.Request) {
	meetings, err := h.meetingService.ListByUser(r.Context(), getUserID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	if meetings == nil {
		meetings = []domain.Meeting{}
	}
	writeJSON(w, http.StatusOK, meetings)
}

func (h *MeetingHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	meetings, err := h.meetingService.ListAllVisible(r.Context(), getUserID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	if meetings == nil {
		meetings = []domain.Meeting{}
	}
	writeJSON(w, http.StatusOK, meetings)
}

func (h *MeetingHandler) Get(w http.ResponseWriter, r *http.Request) {
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid meeting id"})
		return
	}

	meeting, err := h.meetingService.GetByID(r.Context(), meetingID, getUserID(r))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, meeting)
}

func (h *MeetingHandler) Update(w http.ResponseWriter, r *http.Request) {
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid meeting id"})
		return
	}

	var input service.UpdateMeetingInput
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	meeting, err := h.meetingService.Update(r.Context(), meetingID, getUserID(r), input)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, meeting)
}

func (h *MeetingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid meeting id"})
		return
	}

	if err := h.meetingService.Delete(r.Context(), meetingID, getUserID(r)); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MeetingHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid meeting id"})
		return
	}

	var input service.ConfirmInput
	// Body is optional — empty means auto-pick
	decodeJSON(r, &input)

	meeting, err := h.meetingService.Confirm(r.Context(), meetingID, getUserID(r), input)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, meeting)
}

func (h *MeetingHandler) AddParticipants(w http.ResponseWriter, r *http.Request) {
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid meeting id"})
		return
	}

	var input struct {
		Emails []string `json:"emails"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.meetingService.AddParticipants(r.Context(), meetingID, getUserID(r), input.Emails); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MeetingHandler) UpdateRSVP(w http.ResponseWriter, r *http.Request) {
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid meeting id"})
		return
	}

	var input struct {
		Status domain.RSVPStatus `json:"status"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.meetingService.UpdateRSVP(r.Context(), meetingID, getUserID(r), input.Status); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MeetingHandler) Vote(w http.ResponseWriter, r *http.Request) {
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid meeting id"})
		return
	}

	var input service.VoteInput
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.meetingService.Vote(r.Context(), meetingID, getUserID(r), input); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MeetingHandler) GetVotes(w http.ResponseWriter, r *http.Request) {
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid meeting id"})
		return
	}

	slots, err := h.meetingService.GetVotes(r.Context(), meetingID, getUserID(r))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, slots)
}

func (h *MeetingHandler) SetTags(w http.ResponseWriter, r *http.Request) {
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid meeting id"})
		return
	}

	var input service.SetTagsInput
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.meetingService.SetTags(r.Context(), meetingID, getUserID(r), input); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MeetingHandler) GetAllTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.meetingService.GetAllTags(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	if tags == nil {
		tags = []string{}
	}
	writeJSON(w, http.StatusOK, tags)
}
