import { Form, Input, Button, Alert } from 'antd';
import { MailOutlined, LockOutlined, UserOutlined } from '@ant-design/icons';
import { useState } from 'react';
import { useAuth } from '../../hooks/useAuth';

export function RegisterForm({ onSuccess }: { onSuccess: () => void }) {
  const { register } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: { email: string; displayName: string; password: string }) => {
    setLoading(true);
    setError(null);
    try {
      await register(values.email, values.displayName, values.password);
      onSuccess();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Form name="register" onFinish={onFinish} layout="vertical" size="large">
      {error && <Alert message={error} type="error" showIcon style={{ marginBottom: 16 }} data-testid="register-error" />}
      <Form.Item name="displayName" rules={[{ required: true, message: 'Please enter your name' }]}>
        <Input prefix={<UserOutlined />} placeholder="Display Name" data-testid="register-name" />
      </Form.Item>
      <Form.Item name="email" rules={[{ required: true, type: 'email', message: 'Please enter a valid email' }]}>
        <Input prefix={<MailOutlined />} placeholder="Email" data-testid="register-email" />
      </Form.Item>
      <Form.Item name="password" rules={[{ required: true, min: 6, message: 'Password must be at least 6 characters' }]}>
        <Input.Password prefix={<LockOutlined />} placeholder="Password" data-testid="register-password" />
      </Form.Item>
      <Form.Item>
        <Button type="primary" htmlType="submit" loading={loading} block data-testid="register-submit">
          Register
        </Button>
      </Form.Item>
    </Form>
  );
}
