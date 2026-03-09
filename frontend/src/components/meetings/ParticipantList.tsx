import { List, Tag } from 'antd';
import { Participant } from '../../api/types';

const statusColors: Record<string, string> = {
  invited: 'default',
  accepted: 'green',
  declined: 'red',
};

export function ParticipantList({ participants }: { participants: Participant[] }) {
  return (
    <List
      size="small"
      dataSource={participants}
      data-testid="participant-list"
      renderItem={(p) => (
        <List.Item>
          <span>{p.user?.displayName || p.userId}</span>
          <Tag color={statusColors[p.rsvpStatus]}>{p.rsvpStatus}</Tag>
        </List.Item>
      )}
    />
  );
}
