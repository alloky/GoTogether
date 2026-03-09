package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gotogether/backend/internal/domain"
	"github.com/gotogether/backend/internal/testutil"
)

func TestCreateMeeting_Success(t *testing.T) {
	organizerID := uuid.New()
	meetingRepo := &testutil.MockMeetingRepo{}
	userRepo := &testutil.MockUserRepo{
		FindByEmailsFn: func(ctx context.Context, emails []string) ([]domain.User, error) {
			return []domain.User{
				{ID: uuid.New(), Email: "bob@example.com", DisplayName: "Bob"},
			}, nil
		},
	}

	svc := NewMeetingService(meetingRepo, userRepo)

	meeting, err := svc.Create(context.Background(), organizerID, CreateMeetingInput{
		Title:       "Team Lunch",
		Description: "Let's eat together",
		TimeSlots: []TimeSlotInput{
			{StartTime: time.Now().Add(24 * time.Hour), EndTime: time.Now().Add(25 * time.Hour)},
		},
		ParticipantEmails: []string{"bob@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meeting.Title != "Team Lunch" {
		t.Fatalf("expected title 'Team Lunch', got %s", meeting.Title)
	}
	if meeting.Status != domain.MeetingStatusPending {
		t.Fatalf("expected status pending, got %s", meeting.Status)
	}
}

func TestCreateMeeting_EmptyTitle(t *testing.T) {
	meetingRepo := &testutil.MockMeetingRepo{}
	userRepo := &testutil.MockUserRepo{}
	svc := NewMeetingService(meetingRepo, userRepo)

	_, err := svc.Create(context.Background(), uuid.New(), CreateMeetingInput{
		Title: "",
		TimeSlots: []TimeSlotInput{
			{StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)},
		},
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestCreateMeeting_NoTimeSlots(t *testing.T) {
	meetingRepo := &testutil.MockMeetingRepo{}
	userRepo := &testutil.MockUserRepo{}
	svc := NewMeetingService(meetingRepo, userRepo)

	_, err := svc.Create(context.Background(), uuid.New(), CreateMeetingInput{
		Title:     "Test",
		TimeSlots: []TimeSlotInput{},
	})
	if err == nil {
		t.Fatal("expected error for no time slots")
	}
}

func TestConfirmMeeting_AutoPick(t *testing.T) {
	organizerID := uuid.New()
	meetingID := uuid.New()
	slot1ID := uuid.New()
	slot2ID := uuid.New()

	meetingRepo := &testutil.MockMeetingRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Meeting, error) {
			return &domain.Meeting{
				ID:          meetingID,
				OrganizerID: organizerID,
				Status:      domain.MeetingStatusPending,
			}, nil
		},
		GetVoteSummaryFn: func(ctx context.Context, meetingID uuid.UUID) ([]domain.TimeSlot, error) {
			return []domain.TimeSlot{
				{ID: slot1ID, VoteCount: 3},
				{ID: slot2ID, VoteCount: 1},
			}, nil
		},
	}
	userRepo := &testutil.MockUserRepo{}
	svc := NewMeetingService(meetingRepo, userRepo)

	meeting, err := svc.Confirm(context.Background(), meetingID, organizerID, ConfirmInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meeting.Status != domain.MeetingStatusConfirmed {
		t.Fatalf("expected confirmed, got %s", meeting.Status)
	}
	if *meeting.ConfirmedSlotID != slot1ID {
		t.Fatalf("expected slot %s, got %s", slot1ID, *meeting.ConfirmedSlotID)
	}
}

func TestConfirmMeeting_SpecificSlot(t *testing.T) {
	organizerID := uuid.New()
	meetingID := uuid.New()
	slotID := uuid.New()

	meetingRepo := &testutil.MockMeetingRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Meeting, error) {
			return &domain.Meeting{
				ID:          meetingID,
				OrganizerID: organizerID,
				Status:      domain.MeetingStatusPending,
			}, nil
		},
	}
	userRepo := &testutil.MockUserRepo{}
	svc := NewMeetingService(meetingRepo, userRepo)

	meeting, err := svc.Confirm(context.Background(), meetingID, organizerID, ConfirmInput{
		TimeSlotID: &slotID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *meeting.ConfirmedSlotID != slotID {
		t.Fatalf("expected slot %s, got %s", slotID, *meeting.ConfirmedSlotID)
	}
}

func TestConfirmMeeting_NotOrganizer(t *testing.T) {
	organizerID := uuid.New()
	otherUserID := uuid.New()
	meetingID := uuid.New()

	meetingRepo := &testutil.MockMeetingRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Meeting, error) {
			return &domain.Meeting{
				ID:          meetingID,
				OrganizerID: organizerID,
				Status:      domain.MeetingStatusPending,
			}, nil
		},
	}
	userRepo := &testutil.MockUserRepo{}
	svc := NewMeetingService(meetingRepo, userRepo)

	_, err := svc.Confirm(context.Background(), meetingID, otherUserID, ConfirmInput{})
	if err == nil {
		t.Fatal("expected forbidden error")
	}
}

func TestConfirmMeeting_AlreadyConfirmed(t *testing.T) {
	organizerID := uuid.New()
	meetingID := uuid.New()

	meetingRepo := &testutil.MockMeetingRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Meeting, error) {
			return &domain.Meeting{
				ID:          meetingID,
				OrganizerID: organizerID,
				Status:      domain.MeetingStatusConfirmed,
			}, nil
		},
	}
	userRepo := &testutil.MockUserRepo{}
	svc := NewMeetingService(meetingRepo, userRepo)

	_, err := svc.Confirm(context.Background(), meetingID, organizerID, ConfirmInput{})
	if err == nil {
		t.Fatal("expected error for already confirmed meeting")
	}
}

func TestVote_Success(t *testing.T) {
	organizerID := uuid.New()
	voterID := uuid.New()
	meetingID := uuid.New()
	slotID := uuid.New()

	meetingRepo := &testutil.MockMeetingRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Meeting, error) {
			return &domain.Meeting{
				ID:          meetingID,
				OrganizerID: organizerID,
				Status:      domain.MeetingStatusPending,
			}, nil
		},
		GetParticipantsFn: func(ctx context.Context, mid uuid.UUID) ([]domain.Participant, error) {
			return []domain.Participant{
				{UserID: voterID, MeetingID: meetingID},
			}, nil
		},
		GetTimeSlotsFn: func(ctx context.Context, mid uuid.UUID) ([]domain.TimeSlot, error) {
			return []domain.TimeSlot{
				{ID: slotID, MeetingID: meetingID},
			}, nil
		},
	}
	userRepo := &testutil.MockUserRepo{}
	svc := NewMeetingService(meetingRepo, userRepo)

	err := svc.Vote(context.Background(), meetingID, voterID, VoteInput{
		TimeSlotIDs: []uuid.UUID{slotID},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVote_NotParticipant(t *testing.T) {
	organizerID := uuid.New()
	randomUserID := uuid.New()
	meetingID := uuid.New()
	slotID := uuid.New()

	meetingRepo := &testutil.MockMeetingRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Meeting, error) {
			return &domain.Meeting{
				ID:          meetingID,
				OrganizerID: organizerID,
				Status:      domain.MeetingStatusPending,
			}, nil
		},
		GetParticipantsFn: func(ctx context.Context, mid uuid.UUID) ([]domain.Participant, error) {
			return []domain.Participant{}, nil
		},
	}
	userRepo := &testutil.MockUserRepo{}
	svc := NewMeetingService(meetingRepo, userRepo)

	err := svc.Vote(context.Background(), meetingID, randomUserID, VoteInput{
		TimeSlotIDs: []uuid.UUID{slotID},
	})
	if err == nil {
		t.Fatal("expected forbidden error")
	}
}

func TestDeleteMeeting_Success(t *testing.T) {
	organizerID := uuid.New()
	meetingID := uuid.New()
	deleted := false

	meetingRepo := &testutil.MockMeetingRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Meeting, error) {
			return &domain.Meeting{
				ID:          meetingID,
				OrganizerID: organizerID,
			}, nil
		},
		DeleteFn: func(ctx context.Context, id uuid.UUID) error {
			deleted = true
			return nil
		},
	}
	userRepo := &testutil.MockUserRepo{}
	svc := NewMeetingService(meetingRepo, userRepo)

	err := svc.Delete(context.Background(), meetingID, organizerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatal("expected meeting to be deleted")
	}
}

func TestDeleteMeeting_NotOrganizer(t *testing.T) {
	meetingRepo := &testutil.MockMeetingRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Meeting, error) {
			return &domain.Meeting{
				ID:          uuid.New(),
				OrganizerID: uuid.New(),
			}, nil
		},
	}
	userRepo := &testutil.MockUserRepo{}
	svc := NewMeetingService(meetingRepo, userRepo)

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected forbidden error")
	}
}
