package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gotogether/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MeetingRepo struct {
	db *pgxpool.Pool
}

func NewMeetingRepo(db *pgxpool.Pool) *MeetingRepo {
	return &MeetingRepo{db: db}
}

func (r *MeetingRepo) Create(ctx context.Context, meeting *domain.Meeting) error {
	meeting.ID = uuid.New()
	err := r.db.QueryRow(ctx,
		`INSERT INTO meetings (id, title, description, organizer_id, status, is_public)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING created_at`,
		meeting.ID, meeting.Title, meeting.Description, meeting.OrganizerID, meeting.Status, meeting.IsPublic,
	).Scan(&meeting.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting meeting: %w", err)
	}
	return nil
}

func (r *MeetingRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Meeting, error) {
	var m domain.Meeting
	err := r.db.QueryRow(ctx,
		`SELECT id, title, description, organizer_id, status, is_public, confirmed_slot_id, created_at
		 FROM meetings WHERE id = $1`,
		id,
	).Scan(&m.ID, &m.Title, &m.Description, &m.OrganizerID, &m.Status, &m.IsPublic, &m.ConfirmedSlotID, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("finding meeting: %w", err)
	}
	return &m, nil
}

func (r *MeetingRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Meeting, error) {
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT m.id, m.title, m.description, m.organizer_id, m.status, m.is_public, m.confirmed_slot_id, m.created_at
		 FROM meetings m
		 LEFT JOIN participants p ON p.meeting_id = m.id
		 WHERE m.organizer_id = $1 OR p.user_id = $1
		 ORDER BY m.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing meetings: %w", err)
	}
	defer rows.Close()

	var meetings []domain.Meeting
	for rows.Next() {
		var m domain.Meeting
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.OrganizerID, &m.Status, &m.IsPublic, &m.ConfirmedSlotID, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning meeting: %w", err)
		}
		meetings = append(meetings, m)
	}
	return meetings, nil
}

func (r *MeetingRepo) ListAllVisible(ctx context.Context, userID uuid.UUID) ([]domain.Meeting, error) {
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT m.id, m.title, m.description, m.organizer_id, m.status, m.is_public, m.confirmed_slot_id, m.created_at
		 FROM meetings m
		 LEFT JOIN participants p ON p.meeting_id = m.id
		 WHERE m.is_public = true OR m.organizer_id = $1 OR p.user_id = $1
		 ORDER BY m.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing all visible meetings: %w", err)
	}
	defer rows.Close()

	var meetings []domain.Meeting
	for rows.Next() {
		var m domain.Meeting
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.OrganizerID, &m.Status, &m.IsPublic, &m.ConfirmedSlotID, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning meeting: %w", err)
		}
		meetings = append(meetings, m)
	}
	return meetings, nil
}

func (r *MeetingRepo) Update(ctx context.Context, meeting *domain.Meeting) error {
	_, err := r.db.Exec(ctx,
		`UPDATE meetings SET title=$1, description=$2, status=$3, confirmed_slot_id=$4, is_public=$5 WHERE id=$6`,
		meeting.Title, meeting.Description, meeting.Status, meeting.ConfirmedSlotID, meeting.IsPublic, meeting.ID,
	)
	if err != nil {
		return fmt.Errorf("updating meeting: %w", err)
	}
	return nil
}

func (r *MeetingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM meetings WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting meeting: %w", err)
	}
	return nil
}

func (r *MeetingRepo) CreateTimeSlots(ctx context.Context, slots []domain.TimeSlot) error {
	for i := range slots {
		slots[i].ID = uuid.New()
		_, err := r.db.Exec(ctx,
			`INSERT INTO time_slots (id, meeting_id, start_time, end_time) VALUES ($1, $2, $3, $4)`,
			slots[i].ID, slots[i].MeetingID, slots[i].StartTime, slots[i].EndTime,
		)
		if err != nil {
			return fmt.Errorf("inserting time slot: %w", err)
		}
	}
	return nil
}

func (r *MeetingRepo) GetTimeSlots(ctx context.Context, meetingID uuid.UUID) ([]domain.TimeSlot, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, meeting_id, start_time, end_time FROM time_slots WHERE meeting_id = $1 ORDER BY start_time`,
		meetingID,
	)
	if err != nil {
		return nil, fmt.Errorf("getting time slots: %w", err)
	}
	defer rows.Close()

	var slots []domain.TimeSlot
	for rows.Next() {
		var s domain.TimeSlot
		if err := rows.Scan(&s.ID, &s.MeetingID, &s.StartTime, &s.EndTime); err != nil {
			return nil, fmt.Errorf("scanning time slot: %w", err)
		}
		slots = append(slots, s)
	}
	return slots, nil
}

func (r *MeetingRepo) AddParticipants(ctx context.Context, participants []domain.Participant) error {
	for i := range participants {
		participants[i].ID = uuid.New()
		_, err := r.db.Exec(ctx,
			`INSERT INTO participants (id, meeting_id, user_id, rsvp_status) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (meeting_id, user_id) DO NOTHING`,
			participants[i].ID, participants[i].MeetingID, participants[i].UserID, participants[i].RSVPStatus,
		)
		if err != nil {
			return fmt.Errorf("inserting participant: %w", err)
		}
	}
	return nil
}

func (r *MeetingRepo) GetParticipants(ctx context.Context, meetingID uuid.UUID) ([]domain.Participant, error) {
	rows, err := r.db.Query(ctx,
		`SELECT p.id, p.meeting_id, p.user_id, p.rsvp_status,
		        u.id, u.email, u.display_name
		 FROM participants p
		 JOIN users u ON u.id = p.user_id
		 WHERE p.meeting_id = $1`,
		meetingID,
	)
	if err != nil {
		return nil, fmt.Errorf("getting participants: %w", err)
	}
	defer rows.Close()

	var participants []domain.Participant
	for rows.Next() {
		var p domain.Participant
		var u domain.User
		if err := rows.Scan(&p.ID, &p.MeetingID, &p.UserID, &p.RSVPStatus, &u.ID, &u.Email, &u.DisplayName); err != nil {
			return nil, fmt.Errorf("scanning participant: %w", err)
		}
		p.User = &u
		participants = append(participants, p)
	}
	return participants, nil
}

func (r *MeetingRepo) UpdateRSVP(ctx context.Context, meetingID, userID uuid.UUID, status domain.RSVPStatus) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE participants SET rsvp_status = $1 WHERE meeting_id = $2 AND user_id = $3`,
		status, meetingID, userID,
	)
	if err != nil {
		return fmt.Errorf("updating rsvp: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *MeetingRepo) ReplaceVotes(ctx context.Context, meetingID, userID uuid.UUID, slotIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing votes for this user on this meeting's slots
	_, err = tx.Exec(ctx,
		`DELETE FROM votes WHERE user_id = $1 AND time_slot_id IN (
			SELECT id FROM time_slots WHERE meeting_id = $2
		)`,
		userID, meetingID,
	)
	if err != nil {
		return fmt.Errorf("deleting old votes: %w", err)
	}

	// Insert new votes
	for _, slotID := range slotIDs {
		_, err = tx.Exec(ctx,
			`INSERT INTO votes (id, time_slot_id, user_id) VALUES ($1, $2, $3)`,
			uuid.New(), slotID, userID,
		)
		if err != nil {
			return fmt.Errorf("inserting vote: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *MeetingRepo) GetVoteSummary(ctx context.Context, meetingID uuid.UUID) ([]domain.TimeSlot, error) {
	rows, err := r.db.Query(ctx,
		`SELECT ts.id, ts.meeting_id, ts.start_time, ts.end_time,
		        COUNT(v.id) as vote_count
		 FROM time_slots ts
		 LEFT JOIN votes v ON v.time_slot_id = ts.id
		 WHERE ts.meeting_id = $1
		 GROUP BY ts.id, ts.meeting_id, ts.start_time, ts.end_time
		 ORDER BY vote_count DESC, ts.start_time ASC`,
		meetingID,
	)
	if err != nil {
		return nil, fmt.Errorf("getting vote summary: %w", err)
	}
	defer rows.Close()

	var slots []domain.TimeSlot
	for rows.Next() {
		var s domain.TimeSlot
		if err := rows.Scan(&s.ID, &s.MeetingID, &s.StartTime, &s.EndTime, &s.VoteCount); err != nil {
			return nil, fmt.Errorf("scanning vote summary: %w", err)
		}

		// Get voters for each slot
		voterRows, err := r.db.Query(ctx,
			`SELECT u.id, u.email, u.display_name
			 FROM votes v JOIN users u ON u.id = v.user_id
			 WHERE v.time_slot_id = $1`,
			s.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("getting voters: %w", err)
		}
		for voterRows.Next() {
			var u domain.User
			if err := voterRows.Scan(&u.ID, &u.Email, &u.DisplayName); err != nil {
				voterRows.Close()
				return nil, fmt.Errorf("scanning voter: %w", err)
			}
			s.Voters = append(s.Voters, u)
		}
		voterRows.Close()

		slots = append(slots, s)
	}
	return slots, nil
}

func (r *MeetingRepo) SetTags(ctx context.Context, meetingID uuid.UUID, tags []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM meeting_tags WHERE meeting_id = $1`, meetingID)
	if err != nil {
		return fmt.Errorf("deleting old tags: %w", err)
	}

	for _, tag := range tags {
		_, err = tx.Exec(ctx,
			`INSERT INTO meeting_tags (meeting_id, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			meetingID, tag,
		)
		if err != nil {
			return fmt.Errorf("inserting tag: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *MeetingRepo) GetTags(ctx context.Context, meetingID uuid.UUID) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT tag FROM meeting_tags WHERE meeting_id = $1 ORDER BY tag`,
		meetingID,
	)
	if err != nil {
		return nil, fmt.Errorf("getting tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scanning tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *MeetingRepo) GetAllTags(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT DISTINCT tag FROM meeting_tags ORDER BY tag`)
	if err != nil {
		return nil, fmt.Errorf("getting all tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scanning tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *MeetingRepo) GetTagsForMeetings(ctx context.Context, meetingIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	if len(meetingIDs) == 0 {
		return make(map[uuid.UUID][]string), nil
	}

	rows, err := r.db.Query(ctx,
		`SELECT meeting_id, tag FROM meeting_tags WHERE meeting_id = ANY($1) ORDER BY tag`,
		meetingIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("getting tags for meetings: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]string)
	for rows.Next() {
		var meetingID uuid.UUID
		var tag string
		if err := rows.Scan(&meetingID, &tag); err != nil {
			return nil, fmt.Errorf("scanning tag: %w", err)
		}
		result[meetingID] = append(result[meetingID], tag)
	}
	return result, nil
}
