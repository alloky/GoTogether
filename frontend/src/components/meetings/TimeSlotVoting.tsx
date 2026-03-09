import { Checkbox, Button, Space, Typography, Progress } from 'antd';
import { useState } from 'react';
import { TimeSlot } from '../../api/types';
import { useVote } from '../../hooks/useMeetings';
import dayjs from 'dayjs';

interface Props {
  meetingId: string;
  timeSlots: TimeSlot[];
  totalParticipants: number;
  currentUserId: string;
  disabled?: boolean;
}

export function TimeSlotVoting({ meetingId, timeSlots, totalParticipants, currentUserId, disabled }: Props) {
  const voteMutation = useVote();

  // Find slots the current user has voted for
  const userVotedSlots = timeSlots
    .filter(s => s.voters?.some(v => v.id === currentUserId))
    .map(s => s.id);

  const [selectedSlots, setSelectedSlots] = useState<string[]>(userVotedSlots);

  const handleVote = async () => {
    await voteMutation.mutateAsync({ id: meetingId, input: { timeSlotIds: selectedSlots } });
  };

  const maxVotes = Math.max(...timeSlots.map(s => s.voteCount), 1);

  return (
    <div data-testid="time-slot-voting">
      <Typography.Title level={5}>Vote for Time Slots</Typography.Title>
      <Space direction="vertical" style={{ width: '100%' }}>
        {timeSlots.map((slot) => (
          <div key={slot.id} style={{ padding: '8px 0', borderBottom: '1px solid #f0f0f0' }}>
            <Checkbox
              checked={selectedSlots.includes(slot.id)}
              disabled={disabled}
              onChange={(e) => {
                if (e.target.checked) {
                  setSelectedSlots([...selectedSlots, slot.id]);
                } else {
                  setSelectedSlots(selectedSlots.filter(id => id !== slot.id));
                }
              }}
              data-testid={`vote-slot-${slot.id}`}
            >
              <Typography.Text strong>
                {dayjs(slot.startTime).format('ddd, MMM D, YYYY h:mm A')} &ndash; {dayjs(slot.endTime).format('h:mm A')}
              </Typography.Text>
            </Checkbox>
            <div style={{ marginLeft: 24, marginTop: 4 }}>
              <Progress
                percent={totalParticipants > 0 ? Math.round((slot.voteCount / totalParticipants) * 100) : 0}
                size="small"
                format={() => `${slot.voteCount} vote${slot.voteCount !== 1 ? 's' : ''}`}
              />
              {slot.voters && slot.voters.length > 0 && (
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                  {slot.voters.map(v => v.displayName).join(', ')}
                </Typography.Text>
              )}
            </div>
          </div>
        ))}

        {!disabled && (
          <Button
            type="primary"
            onClick={handleVote}
            loading={voteMutation.isPending}
            data-testid="submit-votes"
          >
            Submit Votes
          </Button>
        )}
      </Space>
    </div>
  );
}
