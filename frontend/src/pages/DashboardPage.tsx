import { Typography, Button } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { MeetingList } from '../components/meetings/MeetingList';
import { useMeetingsList } from '../hooks/useMeetings';

export function DashboardPage() {
  const navigate = useNavigate();
  const { data: meetings, isLoading } = useMeetingsList();

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Typography.Title level={3} style={{ margin: 0 }}>My Meetings</Typography.Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/meetings/new')} data-testid="new-meeting-btn">
          New Meeting
        </Button>
      </div>
      <MeetingList meetings={meetings || []} loading={isLoading} />
    </div>
  );
}
