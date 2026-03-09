/**
 * CJM: Error Handling
 *
 * Tests that the bot handles error conditions gracefully:
 * - Viewing a deleted/non-existent meeting
 * - Permission errors (participant trying to confirm)
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');
const crypto = require('crypto');

let harness;
let organizer;
let participant;
let api;
let meetingId;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  organizer = harness.createUser('ErrorOrganizer');
  await organizer.sendCommand('/start');
  await organizer.sleep(1000);

  participant = harness.createUser('ErrorParticipant');
  await participant.sendCommand('/start');
  await participant.sleep(1000);

  const orgToken = await api.getToken(organizer.userId, organizer.name);
  const partEmail = `tg_${participant.userId}@telegram.local`;
  const users = await api.searchUsers(orgToken, participant.name);
  const partUser = users.find(u => u.email === partEmail);

  const meeting = await api.createMeeting(orgToken, {
    title: 'Error Handling Meeting',
    description: 'Meeting for testing error paths',
    isPublic: true,
    tags: [],
    timeSlots: [
      { startTime: '2027-07-01T10:00:00Z', endTime: '2027-07-01T11:00:00Z' },
    ],
    participantEmails: [],
    participantIds: partUser ? [partUser.id] : [],
  });
  meetingId = meeting.id;
});

afterAll(async () => {
  await harness.stop();
});

describe('Error Handling', () => {
  test('participant does NOT see Confirm or Delete buttons', async () => {
    await participant.sendCommand('/meetings');
    await participant.pressButtonByText('Error Handling Meeting');

    const resp = await participant.getLastResponse();
    expect(resp.text).toContain('Error Handling Meeting');

    // Participant should see Vote and RSVP buttons but NOT Confirm or Delete
    const confirmBtn = await participant.findButton('Confirm Meeting');
    const deleteBtn = await participant.findButton('Delete');
    expect(confirmBtn).toBeNull();
    expect(deleteBtn).toBeNull();

    // But should see Accept/Decline (participant-specific)
    const acceptBtn = await participant.findButton('Accept');
    expect(acceptBtn).not.toBeNull();
  });

  test('navigating to a deleted meeting shows error message', async () => {
    // First, delete the meeting via API
    const token = await api.getToken(organizer.userId, organizer.name);
    await api.deleteMeeting(token, meetingId);

    // Now try to view it via the bot using a raw callback with \f prefix (telebot format)
    await organizer.pressButton(`\fview|${meetingId}`);
    await organizer.sleep(500);

    const resp = await organizer.getLastResponse();
    // Bot should show "Failed to load meeting. It may have been deleted."
    expect(resp.text).toContain('Failed to load meeting');
  });

  test('viewing with a completely fake meeting ID shows error', async () => {
    const fakeId = crypto.randomUUID();
    await organizer.pressButton(`\fview|${fakeId}`);
    await organizer.sleep(500);

    const resp = await organizer.getLastResponse();
    expect(resp.text).toContain('Failed to load meeting');
  });
});
