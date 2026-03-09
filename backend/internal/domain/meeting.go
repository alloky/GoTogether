package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MeetingStatus string

const (
	MeetingStatusPending   MeetingStatus = "pending"
	MeetingStatusConfirmed MeetingStatus = "confirmed"
	MeetingStatusCancelled MeetingStatus = "cancelled"
)

type Meeting struct {
	ID              uuid.UUID     `json:"id"`
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	OrganizerID     uuid.UUID     `json:"organizerId"`
	Status          MeetingStatus `json:"status"`
	IsPublic        bool          `json:"isPublic"`
	ConfirmedSlotID *uuid.UUID    `json:"confirmedSlotId,omitempty"`
	CreatedAt       time.Time     `json:"createdAt"`
	Tags []string `json:"tags,omitempty"`
	// Populated on detail queries
	TimeSlots    []TimeSlot    `json:"timeSlots,omitempty"`
	Participants []Participant `json:"participants,omitempty"`
	Organizer    *User         `json:"organizer,omitempty"`
}

type TimeSlot struct {
	ID        uuid.UUID `json:"id"`
	MeetingID uuid.UUID `json:"meetingId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	VoteCount int       `json:"voteCount"`
	Voters    []User    `json:"voters,omitempty"`
}

type RSVPStatus string

const (
	RSVPInvited  RSVPStatus = "invited"
	RSVPAccepted RSVPStatus = "accepted"
	RSVPDeclined RSVPStatus = "declined"
)

type Participant struct {
	ID         uuid.UUID  `json:"id"`
	MeetingID  uuid.UUID  `json:"meetingId"`
	UserID     uuid.UUID  `json:"userId"`
	RSVPStatus RSVPStatus `json:"rsvpStatus"`
	User       *User      `json:"user,omitempty"`
}

type Vote struct {
	ID         uuid.UUID `json:"id"`
	TimeSlotID uuid.UUID `json:"timeSlotId"`
	UserID     uuid.UUID `json:"userId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type MeetingRepository interface {
	Create(ctx context.Context, meeting *Meeting) error
	FindByID(ctx context.Context, id uuid.UUID) (*Meeting, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Meeting, error)
	ListAllVisible(ctx context.Context, userID uuid.UUID) ([]Meeting, error)
	Update(ctx context.Context, meeting *Meeting) error
	Delete(ctx context.Context, id uuid.UUID) error

	CreateTimeSlots(ctx context.Context, slots []TimeSlot) error
	GetTimeSlots(ctx context.Context, meetingID uuid.UUID) ([]TimeSlot, error)

	AddParticipants(ctx context.Context, participants []Participant) error
	GetParticipants(ctx context.Context, meetingID uuid.UUID) ([]Participant, error)
	UpdateRSVP(ctx context.Context, meetingID, userID uuid.UUID, status RSVPStatus) error

	ReplaceVotes(ctx context.Context, meetingID, userID uuid.UUID, slotIDs []uuid.UUID) error
	GetVoteSummary(ctx context.Context, meetingID uuid.UUID) ([]TimeSlot, error)

	SetTags(ctx context.Context, meetingID uuid.UUID, tags []string) error
	GetTags(ctx context.Context, meetingID uuid.UUID) ([]string, error)
	GetAllTags(ctx context.Context) ([]string, error)
	GetTagsForMeetings(ctx context.Context, meetingIDs []uuid.UUID) (map[uuid.UUID][]string, error)
}
