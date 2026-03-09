import { Form, Input, Button, DatePicker, Space, Alert, Select, Typography, Switch } from 'antd';
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons';
import { useState, useRef, useCallback } from 'react';
import { useCreateMeeting, useAllTags } from '../../hooks/useMeetings';
import { searchUsers } from '../../api/auth';
import { User } from '../../api/types';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;

export function CreateMeetingForm({ onSuccess }: { onSuccess: (id: string) => void }) {
  const [form] = Form.useForm();
  const createMeeting = useCreateMeeting();
  const { data: existingTags } = useAllTags();
  const [error, setError] = useState<string | null>(null);
  const [userOptions, setUserOptions] = useState<User[]>([]);
  const [searchLoading, setSearchLoading] = useState(false);
  const searchTimer = useRef<ReturnType<typeof setTimeout>>();

  const handleUserSearch = useCallback((value: string) => {
    if (searchTimer.current) clearTimeout(searchTimer.current);
    if (!value || value.length < 2) {
      setUserOptions([]);
      return;
    }
    searchTimer.current = setTimeout(async () => {
      setSearchLoading(true);
      try {
        const users = await searchUsers(value);
        setUserOptions(users);
      } catch {
        setUserOptions([]);
      } finally {
        setSearchLoading(false);
      }
    }, 300);
  }, []);

  const onFinish = async (values: any) => {
    setError(null);
    try {
      const timeSlots = values.timeSlots.map((slot: any) => ({
        startTime: slot.range[0].toISOString(),
        endTime: slot.range[1].toISOString(),
      }));

      const participantIds: string[] = values.participants || [];
      const tags: string[] = values.tags || [];

      const meeting = await createMeeting.mutateAsync({
        title: values.title,
        description: values.description || '',
        isPublic: values.isPublic !== false,
        tags,
        timeSlots,
        participantEmails: [],
        participantIds,
      });

      onSuccess(meeting.id);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create meeting');
    }
  };

  const tagOptions = (existingTags || []).map((tag) => ({ label: tag, value: tag }));

  return (
    <Form form={form} name="create-meeting" onFinish={onFinish} layout="vertical" size="large">
      {error && <Alert message={error} type="error" showIcon style={{ marginBottom: 16 }} />}

      <Form.Item name="title" label="Title" rules={[{ required: true, message: 'Please enter a title' }]}>
        <Input placeholder="e.g. Team Lunch" data-testid="meeting-title" />
      </Form.Item>

      <Form.Item name="description" label="Description">
        <Input.TextArea rows={3} placeholder="What's the meeting about?" data-testid="meeting-description" />
      </Form.Item>

      <Form.Item name="isPublic" label="Visibility" valuePropName="checked" initialValue={true}>
        <Switch
          checkedChildren="Public"
          unCheckedChildren="Private"
          data-testid="meeting-visibility"
        />
      </Form.Item>

      <Form.Item name="tags" label="Tags">
        <Select
          mode="tags"
          placeholder="Select or type new tags..."
          options={tagOptions}
          data-testid="meeting-tags"
          tokenSeparators={[',']}
        />
      </Form.Item>

      <Form.List name="timeSlots" initialValue={[{}]}>
        {(fields, { add, remove }) => (
          <>
            <Typography.Text strong style={{ display: 'block', marginBottom: 8 }}>Time Slots</Typography.Text>
            {fields.map(({ key, name }) => (
              <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                <Form.Item
                  name={[name, 'range']}
                  rules={[{ required: true, message: 'Select time range' }]}
                >
                  <RangePicker
                    showTime={{ format: 'HH:mm' }}
                    format="YYYY-MM-DD HH:mm"
                    data-testid={`time-slot-${name}`}
                    disabledDate={(current) => current && current < dayjs().startOf('day')}
                  />
                </Form.Item>
                {fields.length > 1 && (
                  <MinusCircleOutlined onClick={() => remove(name)} data-testid={`remove-slot-${name}`} />
                )}
              </Space>
            ))}
            <Form.Item>
              <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />} data-testid="add-time-slot">
                Add Time Slot
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>

      <Form.Item name="participants" label="Invite Participants">
        <Select
          mode="multiple"
          placeholder="Search by name..."
          filterOption={false}
          onSearch={handleUserSearch}
          loading={searchLoading}
          notFoundContent={searchLoading ? 'Searching...' : 'Type at least 2 characters to search'}
          showSearch
          data-testid="participant-search"
          optionLabelProp="label"
        >
          {userOptions.map((user) => (
            <Select.Option key={user.id} value={user.id} label={user.displayName}>
              <div data-testid={`user-option-${user.id}`}>
                <span style={{ fontWeight: 500 }}>{user.displayName}</span>
                <span style={{ color: '#888', marginLeft: 8 }}>{user.email}</span>
              </div>
            </Select.Option>
          ))}
        </Select>
      </Form.Item>

      <Form.Item>
        <Button type="primary" htmlType="submit" loading={createMeeting.isPending} block data-testid="create-meeting-submit">
          Create Meeting
        </Button>
      </Form.Item>
    </Form>
  );
}
