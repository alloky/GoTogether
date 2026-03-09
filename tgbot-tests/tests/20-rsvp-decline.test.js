/**
 * CJM: RSVP Decline
 *
 * Tests the Decline path for RSVP (complementing test 09 which tests Accept).
 * Verifies both the bot response and the backend state.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let organizer;
let participant;
let api;
let meetingId;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  organizer = harness.createUser('DeclineOrganizer');
  await organizer.sendCommand('/start');
  await organizer.sleep(1000);

  participant = harness.createUser('DeclineParticipant');
  await participant.sendCommand('/start');
  await participant.sleep(1000);

  const orgToken = await api.getToken(organizer.userId, organizer.name);
  const partEmail = `tg_${participant.userId}@telegram.local`;
  const users = await api.searchUsers(orgToken, participant.name);
  const partUser = users.find(u => u.email === partEmail);

  const meeting = await api.createMeeting(orgToken, {
    title: 'Decline RSVP Meeting',
    description: 'Meeting to test decline',
    isPublic: true,
    tags: [],
    timeSlots: [
      { startTime: '2027-02-20T10:00:00Z', endTime: '2027-02-20T11:00:00Z' },
    ],
    participantEmails: [],
    participantIds: partUser ? [partUser.id] : [],
  });
  meetingId = meeting.id;
});

afterAll(async () => {
  await harness.stop();
});

describe('RSVP Decline', () => {
  test('participant sees Accept/Decline buttons', async () => {
    await participant.sendCommand('/meetings');
    await participant.pressButtonByText('Decline RSVP Meeting');

    const acceptBtn = await participant.findButton('Accept');
    const declineBtn = await participant.findButton('Decline');
    expect(acceptBtn).not.toBeNull();
    expect(declineBtn).not.toBeNull();
  });

  test('participant declines the invitation', async () => {
    await participant.pressButtonByText('Decline');
    await participant.sleep(500);

    const responses = await participant.getLastResponses(3);
    const hasDecline = responses.some(r => r.text.includes('declined'));
    expect(hasDecline).toBe(true);
  });

  test('verify RSVP status is declined via API', async () => {
    const token = await api.getToken(organizer.userId, organizer.name);
    const meeting = await api.getMeeting(token, meetingId);
    const part = meeting.participants.find(
      p => p.user?.email === `tg_${participant.userId}@telegram.local`
    );
    expect(part).toBeDefined();
    expect(part.rsvpStatus).toBe('declined');
  });
});
