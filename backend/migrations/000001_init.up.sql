CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE meetings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    organizer_id UUID NOT NULL REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'pending',
    confirmed_slot_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE time_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meeting_id UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL
);

ALTER TABLE meetings
    ADD CONSTRAINT fk_confirmed_slot
    FOREIGN KEY (confirmed_slot_id) REFERENCES time_slots(id);

CREATE TABLE participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meeting_id UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    rsvp_status TEXT NOT NULL DEFAULT 'invited',
    UNIQUE(meeting_id, user_id)
);

CREATE TABLE votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    time_slot_id UUID NOT NULL REFERENCES time_slots(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(time_slot_id, user_id)
);

CREATE INDEX idx_meetings_organizer ON meetings(organizer_id);
CREATE INDEX idx_time_slots_meeting ON time_slots(meeting_id);
CREATE INDEX idx_participants_meeting ON participants(meeting_id);
CREATE INDEX idx_participants_user ON participants(user_id);
CREATE INDEX idx_votes_slot ON votes(time_slot_id);
CREATE INDEX idx_votes_user ON votes(user_id);
