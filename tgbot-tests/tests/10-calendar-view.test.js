/**
 * CJM: Calendar View
 *
 * User views the calendar which shows confirmed meetings grouped by date.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let user;
let api;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  user = harness.createUser('CalendarViewer');
  await user.sendCommand('/start');
  await user.sleep(1000);

  // Create and confirm a meeting so it appears in the calendar
  const token = await api.getToken(user.userId, user.name);
  const meeting = await api.createMeeting(token, {
    title: 'Calendar Event',
    description: 'Should appear in calendar',
    isPublic: true,
    tags: ['calendar-test'],
    timeSlots: [
      { startTime: '2027-02-10T10:00:00Z', endTime: '2027-02-10T11:00:00Z' },
    ],
    participantEmails: [],
    participantIds: [],
  });
  // Confirm it
  await api.confirmMeeting(token, meeting.id, meeting.timeSlots[0].id);
});

afterAll(async () => {
  await harness.stop();
});

describe('Calendar View', () => {
  test('/calendar shows confirmed meetings', async () => {
    await user.sendCommand('/calendar');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Calendar');
    expect(resp.text).toContain('Calendar Event');
  });

  test('calendar shows time information', async () => {
    const resp = await user.getLastResponse();
    // Should have formatted time
    expect(resp.text).toContain('10:00');
    expect(resp.text).toContain('11:00');
  });
});
