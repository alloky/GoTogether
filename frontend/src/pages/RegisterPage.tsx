import { Card, Typography } from 'antd';
import { Link, useNavigate } from 'react-router-dom';
import { RegisterForm } from '../components/auth/RegisterForm';

export function RegisterPage() {
  const navigate = useNavigate();

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f0f2f5' }}>
      <Card style={{ width: 400 }}>
        <Typography.Title level={2} style={{ textAlign: 'center' }}>GoTogether</Typography.Title>
        <Typography.Paragraph type="secondary" style={{ textAlign: 'center' }}>
          Create your account
        </Typography.Paragraph>
        <RegisterForm onSuccess={() => navigate('/dashboard')} />
        <div style={{ textAlign: 'center', marginTop: 16 }}>
          Already have an account? <Link to="/login">Log in</Link>
        </div>
      </Card>
    </div>
  );
}
