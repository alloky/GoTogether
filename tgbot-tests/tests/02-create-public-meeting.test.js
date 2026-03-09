/**
 * CJM: Create a Public Meeting
 *
 * Full 6-step conversation flow:
 * 1. /new → title
 * 2. description
 * 3. visibility (public)
 * 4. time slots
 * 5. tags
 * 6. participants (/done to skip)
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let user;
let api;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();
  user = harness.createUser('PublicCreator');
  // Register user first
  await user.sendCommand('/start');
  await user.sleep(1000);
});

afterAll(async () => {
  await harness.stop();
});

describe('Create Public Meeting', () => {
  test('Step 1: /new prompts for title', async () => {
    await user.sendCommand('/new');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Create New Meeting');
    expect(resp.text).toContain('title');
  });

  test('Step 2: Send title, prompts for description', async () => {
    await user.sendMessage('Weekly Team Standup');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('description');
  });

  test('Step 3: Send description, prompts for visibility', async () => {
    await user.sendMessage('Our weekly standup meeting');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('visibility');

    // Should have Public/Private buttons
    const publicBtn = await user.findButton('Public');
    const privateBtn = await user.findButton('Private');
    expect(publicBtn).not.toBeNull();
    expect(privateBtn).not.toBeNull();
  });

  test('Step 4: Choose Public, prompts for time slots', async () => {
    await user.pressButtonByText('Public');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Public');
    expect(resp.text).toContain('time slot');
  });

  test('Step 5: Add time slot, then /done', async () => {
    await user.sendMessage('2026-06-15 10:00 - 11:00');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Time slot 1 added');

    // Add a second slot
    await user.sendMessage('2026-06-16 14:00 - 15:00');
    const resp2 = await user.getLastResponse();
    expect(resp2.text).toContain('Time slot 2 added');

    // Done with slots
    await user.sendCommand('/done');
    const resp3 = await user.getLastResponse();
    expect(resp3.text).toContain('tags');
  });

  test('Step 6: Add tags, then participants', async () => {
    await user.sendMessage('standup, weekly');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Invite participants');
  });

  test('Step 7: Skip participants, meeting created', async () => {
    await user.sendCommand('/done');
    // Wait a bit more for the meeting to be created via API
    await user.sleep(1000);
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Meeting created');
    expect(resp.text).toContain('Weekly Team Standup');
  });

  test('verify meeting exists in backend with correct properties', async () => {
    const token = await api.getToken(user.userId, user.name);
    const meetings = await api.listMyMeetings(token);
    const listed = meetings.find(m => m.title === 'Weekly Team Standup');
    expect(listed).toBeDefined();

    // Use getMeeting for full detail (timeSlots, participants)
    const meeting = await api.getMeeting(token, listed.id);
    expect(meeting.description).toBe('Our weekly standup meeting');
    expect(meeting.isPublic).toBe(true);
    expect(meeting.timeSlots).toHaveLength(2);
    expect(meeting.tags).toContain('standup');
    expect(meeting.tags).toContain('weekly');
  });
});
