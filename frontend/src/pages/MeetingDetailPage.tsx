import { Spin, Button, Result } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { useMeetingDetail } from '../hooks/useMeetings';
import { MeetingDetail } from '../components/meetings/MeetingDetail';

export function MeetingDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: meeting, isLoading, error } = useMeetingDetail(id || '');

  if (isLoading) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  }

  if (error || !meeting) {
    return (
      <Result
        status="404"
        title="Meeting not found"
        extra={<Button onClick={() => navigate('/dashboard')}>Back to Dashboard</Button>}
      />
    );
  }

  return (
    <div>
      <Button
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/dashboard')}
        style={{ marginBottom: 16 }}
        data-testid="back-to-dashboard"
      >
        Back
      </Button>
      <MeetingDetail meeting={meeting} />
    </div>
  );
}
