/**
 * CJM: Vote Submit with No Slots Selected
 *
 * Tests that submitting votes without toggling any slot
 * shows an error message and does not alter the backend state.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let voter;
let api;
let meetingId;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  const organizer = harness.createUser('NoVoteOrganizer');
  await organizer.sendCommand('/start');
  await organizer.sleep(1000);

  voter = harness.createUser('NoVoteVoter');
  await voter.sendCommand('/start');
  await voter.sleep(1000);

  const orgToken = await api.getToken(organizer.userId, organizer.name);
  const voterEmail = `tg_${voter.userId}@telegram.local`;
  const users = await api.searchUsers(orgToken, voter.name);
  const voterUser = users.find(u => u.email === voterEmail);

  const meeting = await api.createMeeting(orgToken, {
    title: 'No Vote Meeting',
    description: 'Test empty vote submit',
    isPublic: true,
    tags: [],
    timeSlots: [
      { startTime: '2027-03-01T09:00:00Z', endTime: '2027-03-01T10:00:00Z' },
      { startTime: '2027-03-02T14:00:00Z', endTime: '2027-03-02T15:00:00Z' },
    ],
    participantEmails: [],
    participantIds: voterUser ? [voterUser.id] : [],
  });
  meetingId = meeting.id;
});

afterAll(async () => {
  await harness.stop();
});

describe('Vote Submit with No Slots Selected', () => {
  test('voter opens vote interface', async () => {
    await voter.sendCommand('/meetings');
    await voter.pressButtonByText('No Vote Meeting');
    await voter.pressButtonByText('Vote on Times');

    const resp = await voter.getLastResponse();
    expect(resp.text).toContain('Vote');

    const submitBtn = await voter.findButton('Submit Votes');
    expect(submitBtn).not.toBeNull();
  });

  test('submitting without toggling shows error', async () => {
    await voter.pressButtonByText('Submit Votes');
    const resp = await voter.getLastResponse();
    expect(resp.text).toContain('No slots selected');
  });

  test('verify no votes were recorded via API', async () => {
    const token = await api.getToken(voter.userId, voter.name);
    const meeting = await api.getMeeting(token, meetingId);
    const totalVotes = meeting.timeSlots.reduce((sum, s) => sum + s.voteCount, 0);
    expect(totalVotes).toBe(0);
  });
});
