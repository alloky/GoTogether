import { Typography, Tag, Descriptions, Button, Space, Divider, message, Radio, Select } from 'antd';
import { CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import { useState } from 'react';
import { Meeting } from '../../api/types';
import { useAuth } from '../../hooks/useAuth';
import { useConfirmMeeting, useUpdateRSVP, useSetTags, useAllTags } from '../../hooks/useMeetings';
import { TimeSlotVoting } from './TimeSlotVoting';
import { ParticipantList } from './ParticipantList';
import dayjs from 'dayjs';

const statusColors: Record<string, string> = {
  pending: 'blue',
  confirmed: 'green',
  cancelled: 'red',
};

const tagColors = ['blue', 'cyan', 'geekblue', 'purple', 'magenta', 'volcano', 'gold', 'lime', 'green'];

export function MeetingDetail({ meeting }: { meeting: Meeting }) {
  const { user } = useAuth();
  const confirmMutation = useConfirmMeeting();
  const rsvpMutation = useUpdateRSVP();
  const setTagsMutation = useSetTags();
  const { data: allTags } = useAllTags();
  const [selectedSlotId, setSelectedSlotId] = useState<string | null>(null);
  const [editingTags, setEditingTags] = useState(false);
  const [tagValues, setTagValues] = useState<string[]>(meeting.tags || []);

  const isOrganizer = user?.id === meeting.organizerId;
  const isPending = meeting.status === 'pending';
  const isParticipant = meeting.participants?.some(p => p.userId === user?.id);
  const totalVoters = (meeting.participants?.length || 0) + 1; // +1 for organizer

  const confirmedSlot = meeting.timeSlots?.find(s => s.id === meeting.confirmedSlotId);

  const handleConfirm = async (slotId?: string) => {
    try {
      await confirmMutation.mutateAsync({ id: meeting.id, timeSlotId: slotId });
      message.success('Meeting confirmed!');
    } catch {
      message.error('Failed to confirm meeting');
    }
  };

  const handleRSVP = async (status: 'accepted' | 'declined') => {
    try {
      await rsvpMutation.mutateAsync({ id: meeting.id, status });
      message.success(`RSVP ${status}`);
    } catch {
      message.error('Failed to update RSVP');
    }
  };

  const handleSaveTags = async () => {
    try {
      await setTagsMutation.mutateAsync({ id: meeting.id, tags: tagValues });
      message.success('Tags updated');
      setEditingTags(false);
    } catch {
      message.error('Failed to update tags');
    }
  };

  const tagOptions = (allTags || []).map((tag) => ({ label: tag, value: tag }));

  return (
    <div data-testid="meeting-detail">
      <Typography.Title level={3}>
        {meeting.title}{' '}
        <Tag color={statusColors[meeting.status]} data-testid="meeting-status">{meeting.status}</Tag>
        <Tag color={meeting.isPublic ? 'default' : 'orange'}>{meeting.isPublic ? 'public' : 'private'}</Tag>
      </Typography.Title>

      {/* Tags display */}
      <div style={{ marginBottom: 16 }}>
        {(meeting.tags || []).map((tag, i) => (
          <Tag key={tag} color={tagColors[i % tagColors.length]}>{tag}</Tag>
        ))}
        {isOrganizer && !editingTags && (
          <Tag
            style={{ borderStyle: 'dashed', cursor: 'pointer' }}
            onClick={() => { setTagValues(meeting.tags || []); setEditingTags(true); }}
          >
            + Edit Tags
          </Tag>
        )}
      </div>

      {/* Tag editing for organizer */}
      {isOrganizer && editingTags && (
        <div style={{ marginBottom: 16 }}>
          <Select
            mode="tags"
            style={{ width: '100%', marginBottom: 8 }}
            placeholder="Select or type new tags..."
            value={tagValues}
            onChange={setTagValues}
            options={tagOptions}
            tokenSeparators={[',']}
          />
          <Space>
            <Button type="primary" size="small" onClick={handleSaveTags} loading={setTagsMutation.isPending}>
              Save Tags
            </Button>
            <Button size="small" onClick={() => setEditingTags(false)}>Cancel</Button>
          </Space>
        </div>
      )}

      <Descriptions column={1} bordered size="small">
        <Descriptions.Item label="Organizer">{meeting.organizer?.displayName || 'Unknown'}</Descriptions.Item>
        <Descriptions.Item label="Description">{meeting.description || 'No description'}</Descriptions.Item>
        <Descriptions.Item label="Created">{dayjs(meeting.createdAt).format('MMM D, YYYY h:mm A')}</Descriptions.Item>
        {confirmedSlot && (
          <Descriptions.Item label="Confirmed Time">
            {dayjs(confirmedSlot.startTime).format('ddd, MMM D, YYYY h:mm A')} &ndash; {dayjs(confirmedSlot.endTime).format('h:mm A')}
          </Descriptions.Item>
        )}
      </Descriptions>

      {/* RSVP for participants */}
      {isParticipant && isPending && (
        <>
          <Divider />
          <Space>
            <Button
              type="primary"
              icon={<CheckCircleOutlined />}
              onClick={() => handleRSVP('accepted')}
              loading={rsvpMutation.isPending}
              data-testid="rsvp-accept"
            >
              Accept
            </Button>
            <Button
              danger
              icon={<CloseCircleOutlined />}
              onClick={() => handleRSVP('declined')}
              loading={rsvpMutation.isPending}
              data-testid="rsvp-decline"
            >
              Decline
            </Button>
          </Space>
        </>
      )}

      <Divider />

      {/* Time slots and voting */}
      {meeting.timeSlots && meeting.timeSlots.length > 0 && (
        <TimeSlotVoting
          meetingId={meeting.id}
          timeSlots={meeting.timeSlots}
          totalParticipants={totalVoters}
          currentUserId={user?.id || ''}
          disabled={!isPending}
        />
      )}

      {/* Confirm section for organizer */}
      {isOrganizer && isPending && meeting.timeSlots && meeting.timeSlots.length > 0 && (
        <>
          <Divider />
          <Typography.Title level={5}>Confirm Meeting</Typography.Title>
          <Typography.Paragraph type="secondary">
            Choose a time slot manually, or auto-pick the one with the most votes.
          </Typography.Paragraph>

          <Radio.Group
            value={selectedSlotId}
            onChange={(e) => setSelectedSlotId(e.target.value)}
            style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 16 }}
            data-testid="slot-picker"
          >
            {meeting.timeSlots.map((slot) => (
              <Radio key={slot.id} value={slot.id} data-testid={`pick-slot-${slot.id}`}>
                {dayjs(slot.startTime).format('ddd, MMM D, YYYY h:mm A')} &ndash; {dayjs(slot.endTime).format('h:mm A')}
                <Tag style={{ marginLeft: 8 }}>{slot.voteCount} vote{slot.voteCount !== 1 ? 's' : ''}</Tag>
              </Radio>
            ))}
          </Radio.Group>

          <Space>
            <Button
              type="primary"
              size="large"
              onClick={() => handleConfirm(selectedSlotId || undefined)}
              loading={confirmMutation.isPending}
              data-testid="confirm-meeting"
            >
              {selectedSlotId ? 'Confirm Selected Slot' : 'Auto-pick Best Slot'}
            </Button>
          </Space>
        </>
      )}

      <Divider />

      {/* Participants */}
      <Typography.Title level={5}>Participants</Typography.Title>
      {meeting.participants && meeting.participants.length > 0 ? (
        <ParticipantList participants={meeting.participants} />
      ) : (
        <Typography.Text type="secondary">No participants invited yet.</Typography.Text>
      )}
    </div>
  );
}
