import { Form, Input, Button, Alert } from 'antd';
import { MailOutlined, LockOutlined } from '@ant-design/icons';
import { useState } from 'react';
import { useAuth } from '../../hooks/useAuth';

export function LoginForm({ onSuccess }: { onSuccess: () => void }) {
  const { login } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: { email: string; password: string }) => {
    setLoading(true);
    setError(null);
    try {
      await login(values.email, values.password);
      onSuccess();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Form name="login" onFinish={onFinish} layout="vertical" size="large">
      {error && <Alert message={error} type="error" showIcon style={{ marginBottom: 16 }} data-testid="login-error" />}
      <Form.Item name="email" rules={[{ required: true, message: 'Please enter your email' }]}>
        <Input prefix={<MailOutlined />} placeholder="Email" data-testid="login-email" />
      </Form.Item>
      <Form.Item name="password" rules={[{ required: true, message: 'Please enter your password' }]}>
        <Input.Password prefix={<LockOutlined />} placeholder="Password" data-testid="login-password" />
      </Form.Item>
      <Form.Item>
        <Button type="primary" htmlType="submit" loading={loading} block data-testid="login-submit">
          Log In
        </Button>
      </Form.Item>
    </Form>
  );
}
