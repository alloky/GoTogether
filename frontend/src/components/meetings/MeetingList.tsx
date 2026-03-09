import { List, Tag, Typography, Empty } from 'antd';
import { useNavigate } from 'react-router-dom';
import { Meeting } from '../../api/types';
import dayjs from 'dayjs';

const statusColors: Record<string, string> = {
  pending: 'blue',
  confirmed: 'green',
  cancelled: 'red',
};

export function MeetingList({ meetings, loading }: { meetings: Meeting[]; loading: boolean }) {
  const navigate = useNavigate();

  if (!loading && meetings.length === 0) {
    return <Empty description="No meetings yet. Create one!" data-testid="meetings-empty" />;
  }

  return (
    <List
      loading={loading}
      dataSource={meetings}
      data-testid="meeting-list"
      renderItem={(meeting) => (
        <List.Item
          key={meeting.id}
          onClick={() => navigate(`/meetings/${meeting.id}`)}
          style={{ cursor: 'pointer' }}
          data-testid={`meeting-item-${meeting.id}`}
        >
          <List.Item.Meta
            title={
              <span>
                {meeting.title}{' '}
                <Tag color={statusColors[meeting.status]} data-testid="meeting-status">
                  {meeting.status}
                </Tag>
                {!meeting.isPublic && <Tag color="orange">private</Tag>}
                {(meeting.tags || []).map((tag) => (
                  <Tag key={tag} color="cyan">{tag}</Tag>
                ))}
              </span>
            }
            description={
              <Typography.Text type="secondary">
                {meeting.description || 'No description'} &middot; Created {dayjs(meeting.createdAt).format('MMM D, YYYY')}
              </Typography.Text>
            }
          />
        </List.Item>
      )}
    />
  );
}
