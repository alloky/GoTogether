CREATE TABLE IF NOT EXISTS meeting_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meeting_id UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    UNIQUE(meeting_id, tag)
);

CREATE INDEX idx_meeting_tags_meeting_id ON meeting_tags(meeting_id);
CREATE INDEX idx_meeting_tags_tag ON meeting_tags(tag);
