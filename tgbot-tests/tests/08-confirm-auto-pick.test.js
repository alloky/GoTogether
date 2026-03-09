/**
 * CJM: Confirm Meeting with Auto-Pick
 *
 * Organizer uses the "Auto-pick Best" button which picks the most-voted slot.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let organizer;
let api;
let meetingId;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  organizer = harness.createUser('AutoConfirmer');
  await organizer.sendCommand('/start');
  await organizer.sleep(1000);

  const token = await api.getToken(organizer.userId, organizer.name);
  const meeting = await api.createMeeting(token, {
    title: 'Auto Confirm Meeting',
    description: 'Test auto-pick confirmation',
    isPublic: true,
    tags: [],
    timeSlots: [
      { startTime: '2026-12-01T10:00:00Z', endTime: '2026-12-01T11:00:00Z' },
      { startTime: '2026-12-02T14:00:00Z', endTime: '2026-12-02T15:00:00Z' },
    ],
    participantEmails: [],
    participantIds: [],
  });
  meetingId = meeting.id;
});

afterAll(async () => {
  await harness.stop();
});

describe('Confirm Meeting (Auto-Pick)', () => {
  test('organizer uses auto-pick to confirm', async () => {
    await organizer.sendCommand('/meetings');
    await organizer.pressButtonByText('Auto Confirm Meeting');
    await organizer.pressButtonByText('Confirm Meeting');
    await organizer.pressButtonByText('Auto-pick');
    await organizer.sleep(1000);

    const responses = await organizer.getLastResponses(3);
    const hasConfirm = responses.some(r => r.text.includes('Meeting confirmed'));
    expect(hasConfirm).toBe(true);
  });

  test('verify meeting is confirmed', async () => {
    const token = await api.getToken(organizer.userId, organizer.name);
    const meeting = await api.getMeeting(token, meetingId);
    expect(meeting.status).toBe('confirmed');
  });
});
