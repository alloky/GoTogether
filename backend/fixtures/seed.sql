-- Seed data for development
-- Passwords are bcrypt hash of "password123"
-- Generated with cost 10

INSERT INTO users (id, email, display_name, password_hash) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'alice@example.com', 'Alice Johnson', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'bob@example.com', 'Bob Smith', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'charlie@example.com', 'Charlie Brown', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy')
ON CONFLICT (id) DO NOTHING;

INSERT INTO meetings (id, title, description, organizer_id, status) VALUES
    ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'Team Lunch', 'Let''s grab lunch together!', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'pending')
ON CONFLICT (id) DO NOTHING;

INSERT INTO time_slots (id, meeting_id, start_time, end_time) VALUES
    ('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'd0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', '2026-03-01 12:00:00+00', '2026-03-01 13:00:00+00'),
    ('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a66', 'd0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', '2026-03-02 12:00:00+00', '2026-03-02 13:00:00+00')
ON CONFLICT (id) DO NOTHING;

INSERT INTO participants (meeting_id, user_id, rsvp_status) VALUES
    ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'accepted'),
    ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'invited')
ON CONFLICT DO NOTHING;

INSERT INTO votes (time_slot_id, user_id) VALUES
    ('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'),
    ('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22'),
    ('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a66', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11')
ON CONFLICT DO NOTHING;
