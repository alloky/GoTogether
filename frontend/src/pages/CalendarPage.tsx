import { Typography, Calendar, Badge, Spin, Tag, List, Select, Space, Switch } from 'antd';
import { useNavigate } from 'react-router-dom';
import { useAllMeetings, useAllTags } from '../hooks/useMeetings';
import { useAuth } from '../hooks/useAuth';
import { Meeting, TimeSlot } from '../api/types';
import dayjs, { Dayjs } from 'dayjs';
import { useState, useMemo } from 'react';

interface CalendarEvent {
  meeting: Meeting;
  slot: TimeSlot;
}

function getConfirmedEventsForDate(meetings: Meeting[], date: Dayjs): CalendarEvent[] {
  const events: CalendarEvent[] = [];
  const dateStr = date.format('YYYY-MM-DD');

  for (const meeting of meetings) {
    if (meeting.status !== 'confirmed' || !meeting.confirmedSlotId) continue;
    const confirmedSlot = meeting.timeSlots?.find(s => s.id === meeting.confirmedSlotId);
    if (!confirmedSlot) continue;
    if (dayjs(confirmedSlot.startTime).format('YYYY-MM-DD') === dateStr) {
      events.push({ meeting, slot: confirmedSlot });
    }
  }
  return events;
}

const tagColors = ['blue', 'cyan', 'geekblue', 'purple', 'magenta', 'volcano', 'gold', 'lime', 'green'];

export function CalendarPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { data: meetings, isLoading } = useAllMeetings();
  const { data: allTags } = useAllTags();

  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [showMyEvents, setShowMyEvents] = useState(false);
  const [showPublicOnly, setShowPublicOnly] = useState(false);
  const [showPrivateOnly, setShowPrivateOnly] = useState(false);

  const filteredMeetings = useMemo(() => {
    let result = meetings || [];

    if (showMyEvents && user) {
      result = result.filter(m =>
        m.organizerId === user.id ||
        m.participants?.some(p => p.userId === user.id)
      );
    }

    if (showPublicOnly) {
      result = result.filter(m => m.isPublic);
    }

    if (showPrivateOnly) {
      result = result.filter(m => !m.isPublic);
    }

    if (selectedTags.length > 0) {
      result = result.filter(m =>
        selectedTags.some(tag => (m.tags || []).includes(tag))
      );
    }

    return result;
  }, [meetings, selectedTags, showMyEvents, showPublicOnly, showPrivateOnly, user]);

  if (isLoading) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  }

  const tagOptions = (allTags || []).map((tag) => ({ label: tag, value: tag }));

  const dateCellRender = (date: Dayjs) => {
    const events = getConfirmedEventsForDate(filteredMeetings, date);
    if (events.length === 0) return null;

    return (
      <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
        {events.map((event) => (
          <li key={event.meeting.id} style={{ marginBottom: 2 }}>
            <Badge
              status="success"
              text={
                <span
                  style={{ fontSize: 11, cursor: 'pointer' }}
                  onClick={(e) => {
                    e.stopPropagation();
                    navigate(`/meetings/${event.meeting.id}`);
                  }}
                >
                  {event.meeting.title}
                </span>
              }
            />
          </li>
        ))}
      </ul>
    );
  };

  // Upcoming confirmed events only
  const upcomingEvents = filteredMeetings
    .filter((m) => m.status === 'confirmed' && m.confirmedSlotId)
    .map((meeting) => {
      const slot = meeting.timeSlots?.find(s => s.id === meeting.confirmedSlotId);
      return slot ? { meeting, slot } : null;
    })
    .filter((e): e is CalendarEvent => e !== null)
    .filter((e) => dayjs(e.slot.startTime).isAfter(dayjs().startOf('day')))
    .sort((a, b) => dayjs(a.slot.startTime).unix() - dayjs(b.slot.startTime).unix())
    .slice(0, 20);

  return (
    <div data-testid="calendar-page">
      <Typography.Title level={3} style={{ marginBottom: 16 }}>
        Calendar
      </Typography.Title>

      {/* Filters */}
      <div style={{ marginBottom: 16, padding: 16, background: '#fafafa', borderRadius: 8, border: '1px solid #f0f0f0' }}>
        <Typography.Text strong style={{ display: 'block', marginBottom: 12 }}>Filters</Typography.Text>
        <Space wrap size="middle">
          <Space>
            <Typography.Text>My Events:</Typography.Text>
            <Switch
              size="small"
              checked={showMyEvents}
              onChange={(checked) => setShowMyEvents(checked)}
            />
          </Space>
          <Space>
            <Typography.Text>Public Only:</Typography.Text>
            <Switch
              size="small"
              checked={showPublicOnly}
              onChange={(checked) => { setShowPublicOnly(checked); if (checked) setShowPrivateOnly(false); }}
            />
          </Space>
          <Space>
            <Typography.Text>Private Only:</Typography.Text>
            <Switch
              size="small"
              checked={showPrivateOnly}
              onChange={(checked) => { setShowPrivateOnly(checked); if (checked) setShowPublicOnly(false); }}
            />
          </Space>
          <Select
            mode="multiple"
            placeholder="Filter by tags..."
            style={{ minWidth: 200 }}
            value={selectedTags}
            onChange={setSelectedTags}
            options={tagOptions}
            allowClear
          />
        </Space>
      </div>

      <Calendar cellRender={(date, info) => {
        if (info.type === 'date') return dateCellRender(date);
        return null;
      }} />

      <Typography.Title level={4} style={{ marginTop: 32 }}>
        Upcoming Confirmed Events
      </Typography.Title>

      {upcomingEvents.length === 0 ? (
        <Typography.Text type="secondary">No upcoming confirmed events.</Typography.Text>
      ) : (
        <List
          dataSource={upcomingEvents}
          renderItem={({ meeting, slot }) => (
            <List.Item
              key={meeting.id}
              style={{ cursor: 'pointer' }}
              onClick={() => navigate(`/meetings/${meeting.id}`)}
            >
              <List.Item.Meta
                title={
                  <span>
                    {meeting.title}{' '}
                    <Tag color="green">confirmed</Tag>
                    {!meeting.isPublic && <Tag color="orange">private</Tag>}
                    {(meeting.tags || []).map((tag, i) => (
                      <Tag key={tag} color={tagColors[i % tagColors.length]}>{tag}</Tag>
                    ))}
                  </span>
                }
                description={
                  <>
                    {dayjs(slot.startTime).format('ddd, MMM D, YYYY h:mm A')} &ndash;{' '}
                    {dayjs(slot.endTime).format('h:mm A')}
                    {meeting.organizer && (
                      <span style={{ marginLeft: 12, color: '#888' }}>
                        by {meeting.organizer.displayName}
                      </span>
                    )}
                  </>
                }
              />
            </List.Item>
          )}
        />
      )}
    </div>
  );
}
