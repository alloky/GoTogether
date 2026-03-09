package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gotogether/backend/internal/domain"
)

type MeetingService struct {
	meetingRepo domain.MeetingRepository
	userRepo    domain.UserRepository
}

func NewMeetingService(meetingRepo domain.MeetingRepository, userRepo domain.UserRepository) *MeetingService {
	return &MeetingService{
		meetingRepo: meetingRepo,
		userRepo:    userRepo,
	}
}

type CreateMeetingInput struct {
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	IsPublic          *bool           `json:"isPublic"`
	Tags              []string        `json:"tags"`
	TimeSlots         []TimeSlotInput `json:"timeSlots"`
	ParticipantEmails []string        `json:"participantEmails"`
	ParticipantIDs    []uuid.UUID     `json:"participantIds"`
}

type TimeSlotInput struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type UpdateMeetingInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type VoteInput struct {
	TimeSlotIDs []uuid.UUID `json:"timeSlotIds"`
}

type ConfirmInput struct {
	TimeSlotID *uuid.UUID `json:"timeSlotId,omitempty"`
}

type SetTagsInput struct {
	Tags []string `json:"tags"`
}

func (s *MeetingService) Create(ctx context.Context, organizerID uuid.UUID, input CreateMeetingInput) (*domain.Meeting, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrBadRequest)
	}
	if len(input.TimeSlots) == 0 {
		return nil, fmt.Errorf("%w: at least one time slot is required", domain.ErrBadRequest)
	}

	isPublic := true
	if input.IsPublic != nil {
		isPublic = *input.IsPublic
	}

	meeting := &domain.Meeting{
		Title:       input.Title,
		Description: input.Description,
		OrganizerID: organizerID,
		Status:      domain.MeetingStatusPending,
		IsPublic:    isPublic,
	}

	if err := s.meetingRepo.Create(ctx, meeting); err != nil {
		return nil, err
	}

	// Set tags if provided
	if len(input.Tags) > 0 {
		if err := s.meetingRepo.SetTags(ctx, meeting.ID, input.Tags); err != nil {
			return nil, err
		}
		meeting.Tags = input.Tags
	}

	// Create time slots
	slots := make([]domain.TimeSlot, len(input.TimeSlots))
	for i, ts := range input.TimeSlots {
		slots[i] = domain.TimeSlot{
			MeetingID: meeting.ID,
			StartTime: ts.StartTime,
			EndTime:   ts.EndTime,
		}
	}
	if err := s.meetingRepo.CreateTimeSlots(ctx, slots); err != nil {
		return nil, err
	}
	meeting.TimeSlots = slots

	// Collect participant user IDs from both emails and direct IDs
	participantUserIDs := make(map[uuid.UUID]bool)

	if len(input.ParticipantEmails) > 0 {
		users, err := s.userRepo.FindByEmails(ctx, input.ParticipantEmails)
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			participantUserIDs[u.ID] = true
		}
	}

	for _, id := range input.ParticipantIDs {
		participantUserIDs[id] = true
	}

	// Don't add organizer as participant
	delete(participantUserIDs, organizerID)

	if len(participantUserIDs) > 0 {
		participants := make([]domain.Participant, 0, len(participantUserIDs))
		for uid := range participantUserIDs {
			participants = append(participants, domain.Participant{
				MeetingID:  meeting.ID,
				UserID:     uid,
				RSVPStatus: domain.RSVPInvited,
			})
		}
		if err := s.meetingRepo.AddParticipants(ctx, participants); err != nil {
			return nil, err
		}
		meeting.Participants = participants
	}

	return meeting, nil
}

func (s *MeetingService) GetByID(ctx context.Context, meetingID uuid.UUID, userID uuid.UUID) (*domain.Meeting, error) {
	meeting, err := s.meetingRepo.FindByID(ctx, meetingID)
	if err != nil {
		return nil, err
	}

	// Check access: must be organizer or participant
	if meeting.OrganizerID != userID {
		participants, err := s.meetingRepo.GetParticipants(ctx, meetingID)
		if err != nil {
			return nil, err
		}
		isParticipant := false
		for _, p := range participants {
			if p.UserID == userID {
				isParticipant = true
				break
			}
		}
		if !isParticipant {
			return nil, domain.ErrForbidden
		}
	}

	// Populate related data
	organizer, err := s.userRepo.FindByID(ctx, meeting.OrganizerID)
	if err != nil {
		return nil, err
	}
	meeting.Organizer = organizer

	slots, err := s.meetingRepo.GetVoteSummary(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	meeting.TimeSlots = slots

	participants, err := s.meetingRepo.GetParticipants(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	meeting.Participants = participants

	tags, err := s.meetingRepo.GetTags(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	meeting.Tags = tags

	return meeting, nil
}

func (s *MeetingService) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Meeting, error) {
	meetings, err := s.meetingRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Populate tags
	if len(meetings) > 0 {
		ids := make([]uuid.UUID, len(meetings))
		for i := range meetings {
			ids[i] = meetings[i].ID
		}
		tagsMap, err := s.meetingRepo.GetTagsForMeetings(ctx, ids)
		if err != nil {
			return nil, err
		}
		for i := range meetings {
			meetings[i].Tags = tagsMap[meetings[i].ID]
		}
	}

	return meetings, nil
}

func (s *MeetingService) ListAllVisible(ctx context.Context, userID uuid.UUID) ([]domain.Meeting, error) {
	meetings, err := s.meetingRepo.ListAllVisible(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Collect meeting IDs for bulk tag fetch
	ids := make([]uuid.UUID, len(meetings))
	for i := range meetings {
		ids[i] = meetings[i].ID
	}

	tagsMap, err := s.meetingRepo.GetTagsForMeetings(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Populate time slots, organizer, and tags for each meeting
	for i := range meetings {
		slots, err := s.meetingRepo.GetTimeSlots(ctx, meetings[i].ID)
		if err != nil {
			return nil, err
		}
		meetings[i].TimeSlots = slots

		organizer, err := s.userRepo.FindByID(ctx, meetings[i].OrganizerID)
		if err != nil {
			return nil, err
		}
		meetings[i].Organizer = organizer

		meetings[i].Tags = tagsMap[meetings[i].ID]
	}

	return meetings, nil
}

func (s *MeetingService) Update(ctx context.Context, meetingID, userID uuid.UUID, input UpdateMeetingInput) (*domain.Meeting, error) {
	meeting, err := s.meetingRepo.FindByID(ctx, meetingID)
	if err != nil {
		return nil, err
	}

	if meeting.OrganizerID != userID {
		return nil, domain.ErrForbidden
	}
	if meeting.Status != domain.MeetingStatusPending {
		return nil, fmt.Errorf("%w: can only update pending meetings", domain.ErrBadRequest)
	}

	if input.Title != "" {
		meeting.Title = input.Title
	}
	if input.Description != "" {
		meeting.Description = input.Description
	}

	if err := s.meetingRepo.Update(ctx, meeting); err != nil {
		return nil, err
	}
	return meeting, nil
}

func (s *MeetingService) Delete(ctx context.Context, meetingID, userID uuid.UUID) error {
	meeting, err := s.meetingRepo.FindByID(ctx, meetingID)
	if err != nil {
		return err
	}
	if meeting.OrganizerID != userID {
		return domain.ErrForbidden
	}
	return s.meetingRepo.Delete(ctx, meetingID)
}

func (s *MeetingService) AddParticipants(ctx context.Context, meetingID, userID uuid.UUID, emails []string) error {
	meeting, err := s.meetingRepo.FindByID(ctx, meetingID)
	if err != nil {
		return err
	}
	if meeting.OrganizerID != userID {
		return domain.ErrForbidden
	}

	users, err := s.userRepo.FindByEmails(ctx, emails)
	if err != nil {
		return err
	}

	participants := make([]domain.Participant, len(users))
	for i, u := range users {
		participants[i] = domain.Participant{
			MeetingID:  meetingID,
			UserID:     u.ID,
			RSVPStatus: domain.RSVPInvited,
		}
	}
	return s.meetingRepo.AddParticipants(ctx, participants)
}

func (s *MeetingService) UpdateRSVP(ctx context.Context, meetingID, userID uuid.UUID, status domain.RSVPStatus) error {
	if status != domain.RSVPAccepted && status != domain.RSVPDeclined {
		return fmt.Errorf("%w: status must be 'accepted' or 'declined'", domain.ErrBadRequest)
	}
	return s.meetingRepo.UpdateRSVP(ctx, meetingID, userID, status)
}

func (s *MeetingService) Vote(ctx context.Context, meetingID, userID uuid.UUID, input VoteInput) error {
	meeting, err := s.meetingRepo.FindByID(ctx, meetingID)
	if err != nil {
		return err
	}
	if meeting.Status != domain.MeetingStatusPending {
		return fmt.Errorf("%w: can only vote on pending meetings", domain.ErrBadRequest)
	}

	// Verify user is organizer or participant
	if meeting.OrganizerID != userID {
		participants, err := s.meetingRepo.GetParticipants(ctx, meetingID)
		if err != nil {
			return err
		}
		isParticipant := false
		for _, p := range participants {
			if p.UserID == userID {
				isParticipant = true
				break
			}
		}
		if !isParticipant {
			return domain.ErrForbidden
		}
	}

	// Verify all slot IDs belong to this meeting
	slots, err := s.meetingRepo.GetTimeSlots(ctx, meetingID)
	if err != nil {
		return err
	}
	slotMap := make(map[uuid.UUID]bool)
	for _, slot := range slots {
		slotMap[slot.ID] = true
	}
	for _, slotID := range input.TimeSlotIDs {
		if !slotMap[slotID] {
			return fmt.Errorf("%w: time slot %s does not belong to this meeting", domain.ErrBadRequest, slotID)
		}
	}

	return s.meetingRepo.ReplaceVotes(ctx, meetingID, userID, input.TimeSlotIDs)
}

func (s *MeetingService) GetVotes(ctx context.Context, meetingID, userID uuid.UUID) ([]domain.TimeSlot, error) {
	// Verify access
	_, err := s.GetByID(ctx, meetingID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return nil, err
		}
		// If meeting not found, return not found
		return nil, err
	}

	return s.meetingRepo.GetVoteSummary(ctx, meetingID)
}

func (s *MeetingService) Confirm(ctx context.Context, meetingID, userID uuid.UUID, input ConfirmInput) (*domain.Meeting, error) {
	meeting, err := s.meetingRepo.FindByID(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	if meeting.OrganizerID != userID {
		return nil, domain.ErrForbidden
	}
	if meeting.Status != domain.MeetingStatusPending {
		return nil, fmt.Errorf("%w: can only confirm pending meetings", domain.ErrBadRequest)
	}

	var chosenSlotID uuid.UUID
	if input.TimeSlotID != nil {
		chosenSlotID = *input.TimeSlotID
	} else {
		// Auto-pick: most votes, earliest start time for ties
		slots, err := s.meetingRepo.GetVoteSummary(ctx, meetingID)
		if err != nil {
			return nil, err
		}
		if len(slots) == 0 {
			return nil, fmt.Errorf("%w: no time slots to confirm", domain.ErrBadRequest)
		}
		chosenSlotID = slots[0].ID // Already sorted by vote_count DESC, start_time ASC
	}

	meeting.Status = domain.MeetingStatusConfirmed
	meeting.ConfirmedSlotID = &chosenSlotID

	if err := s.meetingRepo.Update(ctx, meeting); err != nil {
		return nil, err
	}

	return meeting, nil
}

func (s *MeetingService) SetTags(ctx context.Context, meetingID, userID uuid.UUID, input SetTagsInput) error {
	meeting, err := s.meetingRepo.FindByID(ctx, meetingID)
	if err != nil {
		return err
	}
	if meeting.OrganizerID != userID {
		return domain.ErrForbidden
	}
	return s.meetingRepo.SetTags(ctx, meetingID, input.Tags)
}

func (s *MeetingService) GetAllTags(ctx context.Context) ([]string, error) {
	return s.meetingRepo.GetAllTags(ctx)
}
