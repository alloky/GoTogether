/**
 * CJM: RSVP Accept/Decline
 *
 * A participant accepts or declines a meeting invitation via inline buttons.
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

  organizer = harness.createUser('RSVPOrganizer');
  await organizer.sendCommand('/start');
  await organizer.sleep(1000);

  participant = harness.createUser('RSVPParticipant');
  await participant.sendCommand('/start');
  await participant.sleep(1000);

  const orgToken = await api.getToken(organizer.userId, organizer.name);
  const partEmail = `tg_${participant.userId}@telegram.local`;
  const users = await api.searchUsers(orgToken, participant.name);
  const partUser = users.find(u => u.email === partEmail);

  const meeting = await api.createMeeting(orgToken, {
    title: 'RSVP Test Meeting',
    description: 'Meeting to test RSVP',
    isPublic: true,
    tags: [],
    timeSlots: [
      { startTime: '2027-01-15T10:00:00Z', endTime: '2027-01-15T11:00:00Z' },
    ],
    participantEmails: [],
    participantIds: partUser ? [partUser.id] : [],
  });
  meetingId = meeting.id;
});

afterAll(async () => {
  await harness.stop();
});

describe('RSVP Accept/Decline', () => {
  test('participant sees Accept/Decline buttons', async () => {
    await participant.sendCommand('/meetings');
    await participant.pressButtonByText('RSVP Test Meeting');

    const acceptBtn = await participant.findButton('Accept');
    const declineBtn = await participant.findButton('Decline');
    expect(acceptBtn).not.toBeNull();
    expect(declineBtn).not.toBeNull();
  });

  test('participant accepts the invitation', async () => {
    await participant.pressButtonByText('Accept');
    await participant.sleep(500);

    const responses = await participant.getLastResponses(3);
    const hasAccept = responses.some(r => r.text.includes('accepted'));
    expect(hasAccept).toBe(true);
  });

  test('verify RSVP status via API', async () => {
    const token = await api.getToken(organizer.userId, organizer.name);
    const meeting = await api.getMeeting(token, meetingId);
    const part = meeting.participants.find(
      p => p.user?.email === `tg_${participant.userId}@telegram.local`
    );
    expect(part).toBeDefined();
    expect(part.rsvpStatus).toBe('accepted');
  });
});
