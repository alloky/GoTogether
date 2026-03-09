package apiclient

import "time"

type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	CreatedAt   time.Time `json:"createdAt"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type Meeting struct {
	ID              string        `json:"id"`
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	OrganizerID     string        `json:"organizerId"`
	Status          string        `json:"status"`
	IsPublic        bool          `json:"isPublic"`
	Tags            []string      `json:"tags"`
	ConfirmedSlotID *string       `json:"confirmedSlotId"`
	CreatedAt       time.Time     `json:"createdAt"`
	Organizer       *User         `json:"organizer"`
	TimeSlots       []TimeSlot    `json:"timeSlots"`
	Participants    []Participant `json:"participants"`
}

type TimeSlot struct {
	ID        string    `json:"id"`
	MeetingID string    `json:"meetingId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	VoteCount int       `json:"voteCount"`
	Voters    []User    `json:"voters"`
}

type Participant struct {
	ID         string `json:"id"`
	MeetingID  string `json:"meetingId"`
	UserID     string `json:"userId"`
	RSVPStatus string `json:"rsvpStatus"`
	User       *User  `json:"user"`
}

type CreateMeetingInput struct {
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	IsPublic         bool           `json:"isPublic"`
	Tags             []string       `json:"tags"`
	TimeSlots        []TimeSlotInput `json:"timeSlots"`
	ParticipantEmails []string      `json:"participantEmails"`
	ParticipantIDs   []string       `json:"participantIds"`
}

type TimeSlotInput struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
