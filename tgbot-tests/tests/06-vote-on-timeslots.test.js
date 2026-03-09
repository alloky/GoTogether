/**
 * CJM: Vote on Meeting Time Slots
 *
 * A participant votes on time slots via the inline keyboard.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let organizer;
let voter;
let api;
let meetingId;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  // Create organizer and voter
  organizer = harness.createUser('VoteOrganizer');
  await organizer.sendCommand('/start');
  await organizer.sleep(1000);

  voter = harness.createUser('VoteVoter');
  await voter.sendCommand('/start');
  await voter.sleep(1000);

  // Get tokens
  const orgToken = await api.getToken(organizer.userId, organizer.name);
  const voterToken = await api.getToken(voter.userId, voter.name);

  // Find voter's user ID from backend
  const voterEmail = `tg_${voter.userId}@telegram.local`;
  const users = await api.searchUsers(orgToken, voter.name);
  const voterUser = users.find(u => u.email === voterEmail);

  // Create a meeting with the voter as participant
  const meeting = await api.createMeeting(orgToken, {
    title: 'Vote Test Meeting',
    description: 'Meeting to test voting',
    isPublic: true,
    tags: ['vote-test'],
    timeSlots: [
      { startTime: '2026-10-01T09:00:00Z', endTime: '2026-10-01T10:00:00Z' },
      { startTime: '2026-10-02T14:00:00Z', endTime: '2026-10-02T15:00:00Z' },
    ],
    participantEmails: [],
    participantIds: voterUser ? [voterUser.id] : [],
  });
  meetingId = meeting.id;
});

afterAll(async () => {
  await harness.stop();
});

describe('Vote on Time Slots', () => {
  test('voter views the meeting detail', async () => {
    await voter.sendCommand('/meetings');
    await voter.sleep(500);
    await voter.pressButtonByText('Vote Test Meeting');
    const resp = await voter.getLastResponse();
    expect(resp.text).toContain('Vote Test Meeting');
  });

  test('voter opens vote interface', async () => {
    await voter.pressButtonByText('Vote on Times');
    const resp = await voter.getLastResponse();
    expect(resp.text).toContain('Vote');
    expect(resp.text).toContain('Vote Test Meeting');

    // Should have toggle buttons for each slot
    const submitBtn = await voter.findButton('Submit Votes');
    expect(submitBtn).not.toBeNull();
  });

  test('voter toggles a slot and submits votes', async () => {
    // Toggle the first time slot
    const history = await voter.getHistory();
    const lastMsg = history[history.length - 1];
    if (lastMsg.reply_markup?.inline_keyboard) {
      // Find the first vtog button (time slot toggle)
      for (const row of lastMsg.reply_markup.inline_keyboard) {
        for (const btn of row) {
          if (btn.callback_data && btn.callback_data.includes('|')) {
            // This is a vtog button
            await voter.pressButton(btn.callback_data);
            break;
          }
        }
        break;
      }
    }

    // Submit votes
    await voter.pressButtonByText('Submit Votes');
    const resp = await voter.getLastResponse();
    // After submit, bot sends confirmation then meeting detail
    const responses = await voter.getLastResponses(3);
    const hasVoteConfirm = responses.some(r => r.text.includes('Votes submitted'));
    expect(hasVoteConfirm).toBe(true);
  });

  test('verify votes are stored via API', async () => {
    const token = await api.getToken(organizer.userId, organizer.name);
    const meeting = await api.getMeeting(token, meetingId);
    const totalVotes = meeting.timeSlots.reduce((sum, s) => sum + s.voteCount, 0);
    expect(totalVotes).toBeGreaterThan(0);
  });
});
