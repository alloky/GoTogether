import { useState } from 'react';
import { Card, Form, Input, Button, Typography, Alert, Descriptions, Tag } from 'antd';
import { UserOutlined, LinkOutlined } from '@ant-design/icons';
import { useMutation } from '@tanstack/react-query';
import { useAuth } from '../hooks/useAuth';
import { linkTelegram } from '../api/profile';
import { getMe } from '../api/auth';

const { Title } = Typography;

export function ProfilePage() {
  const { user } = useAuth();
  const [success, setSuccess] = useState(false);
  const [linkedUsername, setLinkedUsername] = useState(user?.telegramUsername || '');

  const mutation = useMutation({
    mutationFn: (username: string) => linkTelegram(username),
    onSuccess: async (_data, username) => {
      setSuccess(true);
      setLinkedUsername(username.replace(/^@/, ''));
      // Refresh user data
      try {
        await getMe();
      } catch { /* ignore */ }
    },
  });

  const onFinish = (values: { telegramUsername: string }) => {
    setSuccess(false);
    mutation.mutate(values.telegramUsername);
  };

  const currentUsername = linkedUsername || user?.telegramUsername;

  return (
    <div style={{ maxWidth: 600, margin: '0 auto' }}>
      <Title level={3}>
        <UserOutlined /> Profile
      </Title>

      <Card style={{ marginBottom: 24 }}>
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label="Display Name">{user?.displayName}</Descriptions.Item>
          <Descriptions.Item label="Email">{user?.email}</Descriptions.Item>
          <Descriptions.Item label="Member Since">
            {user?.createdAt ? new Date(user.createdAt).toLocaleDateString() : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="Telegram">
            {user?.telegramId ? (
              <Tag color="green">Linked (ID: {user.telegramId})</Tag>
            ) : currentUsername ? (
              <Tag color="blue">@{currentUsername}</Tag>
            ) : (
              <Tag color="default">Not linked</Tag>
            )}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title={<><LinkOutlined /> Link Telegram Account</>}>
        <p style={{ marginBottom: 16, color: '#666' }}>
          Enter your Telegram @username to link it to your profile.
          For full account linking with meeting sync, use the <code>/link</code> command in the Telegram bot.
        </p>

        {success && (
          <Alert
            message="Telegram username linked successfully!"
            type="success"
            showIcon
            closable
            style={{ marginBottom: 16 }}
            onClose={() => setSuccess(false)}
          />
        )}

        {mutation.isError && (
          <Alert
            message="Failed to link Telegram username"
            description={String((mutation.error as Error)?.message || mutation.error)}
            type="error"
            showIcon
            closable
            style={{ marginBottom: 16 }}
          />
        )}

        <Form layout="inline" onFinish={onFinish}>
          <Form.Item
            name="telegramUsername"
            initialValue={currentUsername ? `@${currentUsername}` : ''}
            rules={[{ required: true, message: 'Enter a username' }]}
          >
            <Input placeholder="@username" prefix="@" style={{ width: 220 }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={mutation.isPending}>
              {currentUsername ? 'Update' : 'Link'}
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
