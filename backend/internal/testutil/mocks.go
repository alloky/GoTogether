package testutil

import (
	"context"

	"github.com/google/uuid"
	"github.com/gotogether/backend/internal/domain"
)

// MockUserRepo implements domain.UserRepository for testing
type MockUserRepo struct {
	CreateFn                 func(ctx context.Context, user *domain.User) error
	FindByEmailFn            func(ctx context.Context, email string) (*domain.User, error)
	FindByIDFn               func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByIDsFn              func(ctx context.Context, ids []uuid.UUID) ([]domain.User, error)
	FindByEmailsFn           func(ctx context.Context, emails []string) ([]domain.User, error)
	SearchByNameFn           func(ctx context.Context, query string, limit int) ([]domain.User, error)
	FindByTelegramIDFn       func(ctx context.Context, telegramID int64) (*domain.User, error)
	UpdateTelegramIDFn       func(ctx context.Context, userID uuid.UUID, telegramID int64) error
	UpdateTelegramUsernameFn func(ctx context.Context, userID uuid.UUID, username string) error
	MergeUsersFn             func(ctx context.Context, fromID, toID uuid.UUID) error
}

func (m *MockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	user.ID = uuid.New()
	return nil
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.FindByEmailFn != nil {
		return m.FindByEmailFn(ctx, email)
	}
	return nil, domain.ErrNotFound
}

func (m *MockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *MockUserRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.User, error) {
	if m.FindByIDsFn != nil {
		return m.FindByIDsFn(ctx, ids)
	}
	return nil, nil
}

func (m *MockUserRepo) FindByEmails(ctx context.Context, emails []string) ([]domain.User, error) {
	if m.FindByEmailsFn != nil {
		return m.FindByEmailsFn(ctx, emails)
	}
	return nil, nil
}

func (m *MockUserRepo) SearchByName(ctx context.Context, query string, limit int) ([]domain.User, error) {
	if m.SearchByNameFn != nil {
		return m.SearchByNameFn(ctx, query, limit)
	}
	return nil, nil
}

func (m *MockUserRepo) FindByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	if m.FindByTelegramIDFn != nil {
		return m.FindByTelegramIDFn(ctx, telegramID)
	}
	return nil, domain.ErrNotFound
}

func (m *MockUserRepo) UpdateTelegramID(ctx context.Context, userID uuid.UUID, telegramID int64) error {
	if m.UpdateTelegramIDFn != nil {
		return m.UpdateTelegramIDFn(ctx, userID, telegramID)
	}
	return nil
}

func (m *MockUserRepo) UpdateTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error {
	if m.UpdateTelegramUsernameFn != nil {
		return m.UpdateTelegramUsernameFn(ctx, userID, username)
	}
	return nil
}

func (m *MockUserRepo) MergeUsers(ctx context.Context, fromID, toID uuid.UUID) error {
	if m.MergeUsersFn != nil {
		return m.MergeUsersFn(ctx, fromID, toID)
	}
	return nil
}

// MockMeetingRepo implements domain.MeetingRepository for testing
type MockMeetingRepo struct {
	CreateFn             func(ctx context.Context, meeting *domain.Meeting) error
	FindByIDFn           func(ctx context.Context, id uuid.UUID) (*domain.Meeting, error)
	ListByUserFn         func(ctx context.Context, userID uuid.UUID) ([]domain.Meeting, error)
	ListAllVisibleFn     func(ctx context.Context, userID uuid.UUID) ([]domain.Meeting, error)
	UpdateFn             func(ctx context.Context, meeting *domain.Meeting) error
	DeleteFn             func(ctx context.Context, id uuid.UUID) error
	CreateTimeSlotsFn    func(ctx context.Context, slots []domain.TimeSlot) error
	GetTimeSlotsFn       func(ctx context.Context, meetingID uuid.UUID) ([]domain.TimeSlot, error)
	AddParticipantsFn    func(ctx context.Context, participants []domain.Participant) error
	GetParticipantsFn    func(ctx context.Context, meetingID uuid.UUID) ([]domain.Participant, error)
	UpdateRSVPFn         func(ctx context.Context, meetingID, userID uuid.UUID, status domain.RSVPStatus) error
	ReplaceVotesFn       func(ctx context.Context, meetingID, userID uuid.UUID, slotIDs []uuid.UUID) error
	GetVoteSummaryFn     func(ctx context.Context, meetingID uuid.UUID) ([]domain.TimeSlot, error)
	SetTagsFn            func(ctx context.Context, meetingID uuid.UUID, tags []string) error
	GetTagsFn            func(ctx context.Context, meetingID uuid.UUID) ([]string, error)
	GetAllTagsFn         func(ctx context.Context) ([]string, error)
	GetTagsForMeetingsFn func(ctx context.Context, meetingIDs []uuid.UUID) (map[uuid.UUID][]string, error)
}

func (m *MockMeetingRepo) Create(ctx context.Context, meeting *domain.Meeting) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, meeting)
	}
	meeting.ID = uuid.New()
	return nil
}

func (m *MockMeetingRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Meeting, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *MockMeetingRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Meeting, error) {
	if m.ListByUserFn != nil {
		return m.ListByUserFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockMeetingRepo) ListAllVisible(ctx context.Context, userID uuid.UUID) ([]domain.Meeting, error) {
	if m.ListAllVisibleFn != nil {
		return m.ListAllVisibleFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockMeetingRepo) Update(ctx context.Context, meeting *domain.Meeting) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, meeting)
	}
	return nil
}

func (m *MockMeetingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func (m *MockMeetingRepo) CreateTimeSlots(ctx context.Context, slots []domain.TimeSlot) error {
	if m.CreateTimeSlotsFn != nil {
		return m.CreateTimeSlotsFn(ctx, slots)
	}
	for i := range slots {
		slots[i].ID = uuid.New()
	}
	return nil
}

func (m *MockMeetingRepo) GetTimeSlots(ctx context.Context, meetingID uuid.UUID) ([]domain.TimeSlot, error) {
	if m.GetTimeSlotsFn != nil {
		return m.GetTimeSlotsFn(ctx, meetingID)
	}
	return nil, nil
}

func (m *MockMeetingRepo) AddParticipants(ctx context.Context, participants []domain.Participant) error {
	if m.AddParticipantsFn != nil {
		return m.AddParticipantsFn(ctx, participants)
	}
	return nil
}

func (m *MockMeetingRepo) GetParticipants(ctx context.Context, meetingID uuid.UUID) ([]domain.Participant, error) {
	if m.GetParticipantsFn != nil {
		return m.GetParticipantsFn(ctx, meetingID)
	}
	return nil, nil
}

func (m *MockMeetingRepo) UpdateRSVP(ctx context.Context, meetingID, userID uuid.UUID, status domain.RSVPStatus) error {
	if m.UpdateRSVPFn != nil {
		return m.UpdateRSVPFn(ctx, meetingID, userID, status)
	}
	return nil
}

func (m *MockMeetingRepo) ReplaceVotes(ctx context.Context, meetingID, userID uuid.UUID, slotIDs []uuid.UUID) error {
	if m.ReplaceVotesFn != nil {
		return m.ReplaceVotesFn(ctx, meetingID, userID, slotIDs)
	}
	return nil
}

func (m *MockMeetingRepo) GetVoteSummary(ctx context.Context, meetingID uuid.UUID) ([]domain.TimeSlot, error) {
	if m.GetVoteSummaryFn != nil {
		return m.GetVoteSummaryFn(ctx, meetingID)
	}
	return nil, nil
}

func (m *MockMeetingRepo) SetTags(ctx context.Context, meetingID uuid.UUID, tags []string) error {
	if m.SetTagsFn != nil {
		return m.SetTagsFn(ctx, meetingID, tags)
	}
	return nil
}

func (m *MockMeetingRepo) GetTags(ctx context.Context, meetingID uuid.UUID) ([]string, error) {
	if m.GetTagsFn != nil {
		return m.GetTagsFn(ctx, meetingID)
	}
	return nil, nil
}

func (m *MockMeetingRepo) GetAllTags(ctx context.Context) ([]string, error) {
	if m.GetAllTagsFn != nil {
		return m.GetAllTagsFn(ctx)
	}
	return nil, nil
}

func (m *MockMeetingRepo) GetTagsForMeetings(ctx context.Context, meetingIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	if m.GetTagsForMeetingsFn != nil {
		return m.GetTagsForMeetingsFn(ctx, meetingIDs)
	}
	return make(map[uuid.UUID][]string), nil
}
