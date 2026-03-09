/**
 * CJM: Delete Meeting
 *
 * Organizer deletes a meeting via the inline button.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let user;
let api;
let meetingId;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  user = harness.createUser('DeleteUser');
  await user.sendCommand('/start');
  await user.sleep(1000);

  const token = await api.getToken(user.userId, user.name);
  const meeting = await api.createMeeting(token, {
    title: 'Meeting to Delete',
    isPublic: true,
    tags: [],
    timeSlots: [
      { startTime: '2027-04-01T10:00:00Z', endTime: '2027-04-01T11:00:00Z' },
    ],
    participantEmails: [],
    participantIds: [],
  });
  meetingId = meeting.id;
});

afterAll(async () => {
  await harness.stop();
});

describe('Delete Meeting', () => {
  test('organizer views meeting and sees delete button', async () => {
    await user.sendCommand('/meetings');
    await user.pressButtonByText('Meeting to Delete');

    const delBtn = await user.findButton('Delete');
    expect(delBtn).not.toBeNull();
  });

  test('organizer deletes the meeting', async () => {
    await user.pressButtonByText('Delete');
    await user.sleep(1000);

    const responses = await user.getLastResponses(3);
    const hasDeleted = responses.some(r => r.text.includes('deleted'));
    expect(hasDeleted).toBe(true);
  });

  test('verify meeting is deleted via API', async () => {
    const token = await api.getToken(user.userId, user.name);
    try {
      await api.getMeeting(token, meetingId);
      // If we get here, meeting wasn't deleted - fail
      expect(true).toBe(false);
    } catch (e) {
      expect(e.response?.status).toBe(404);
    }
  });
});
