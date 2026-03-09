import { Typography } from 'antd';
import { useNavigate } from 'react-router-dom';
import { CreateMeetingForm } from '../components/meetings/CreateMeetingForm';

export function CreateMeetingPage() {
  const navigate = useNavigate();

  return (
    <div style={{ maxWidth: 600 }}>
      <Typography.Title level={3}>Create a Meeting</Typography.Title>
      <CreateMeetingForm onSuccess={(id) => navigate(`/meetings/${id}`)} />
    </div>
  );
}
