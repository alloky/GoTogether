/**
 * CJM: View Meetings List & Navigation
 *
 * User creates some meetings, then views the list via /meetings.
 * Tests pagination and meeting detail view.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let user;
let api;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();
  user = harness.createUser('ListViewer');
  await user.sendCommand('/start');
  await user.sleep(1000);

  // Create a meeting via API so user has something to view
  const token = await api.getToken(user.userId, user.name);
  await api.createMeeting(token, {
    title: 'Test Meeting for List',
    description: 'A test meeting',
    isPublic: true,
    tags: ['test'],
    timeSlots: [
      {
        startTime: '2026-09-01T10:00:00Z',
        endTime: '2026-09-01T11:00:00Z',
      },
    ],
    participantEmails: [],
    participantIds: [],
  });
});

afterAll(async () => {
  await harness.stop();
});

describe('View Meetings List', () => {
  test('/meetings shows the list of meetings', async () => {
    await user.sendCommand('/meetings');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Your Meetings');
    expect(resp.text).toContain('Test Meeting for List');
  });

  test('meetings list has inline buttons for each meeting', async () => {
    const resp = await user.getLastResponse();
    expect(resp.reply_markup).not.toBeNull();

    const buttons = resp.reply_markup.inline_keyboard.flat();
    const meetingBtn = buttons.find(b => b.text.includes('Test Meeting for List'));
    expect(meetingBtn).toBeDefined();
  });

  test('clicking a meeting shows its details', async () => {
    await user.pressButtonByText('Test Meeting for List');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Test Meeting for List');
    expect(resp.text).toContain('A test meeting');
    expect(resp.text).toContain('organizer');
  });

  test('displayed detail matches backend API data', async () => {
    const token = await api.getToken(user.userId, user.name);
    const meetings = await api.listMyMeetings(token);
    const meeting = meetings.find(m => m.title === 'Test Meeting for List');
    expect(meeting).toBeDefined();

    const resp = await user.getLastResponse();
    expect(resp.text).toContain(meeting.title);
    expect(resp.text).toContain(meeting.description);
  });
});
