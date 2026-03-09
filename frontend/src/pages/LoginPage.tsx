import { Card, Typography } from 'antd';
import { Link, useNavigate } from 'react-router-dom';
import { LoginForm } from '../components/auth/LoginForm';

export function LoginPage() {
  const navigate = useNavigate();

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f0f2f5' }}>
      <Card style={{ width: 400 }}>
        <Typography.Title level={2} style={{ textAlign: 'center' }}>GoTogether</Typography.Title>
        <Typography.Paragraph type="secondary" style={{ textAlign: 'center' }}>
          Plan meetings with friends
        </Typography.Paragraph>
        <LoginForm onSuccess={() => navigate('/dashboard')} />
        <div style={{ textAlign: 'center', marginTop: 16 }}>
          Don't have an account? <Link to="/register">Register</Link>
        </div>
      </Card>
    </div>
  );
}
