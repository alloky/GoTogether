/**
 * CJM: Vote Toggle and Submit
 *
 * Tests the vote toggle → submit → API verification flow.
 * Opens vote UI, verifies slot buttons are present with unchecked icons,
 * toggles a slot, submits, and confirms votes are recorded.
 *
 * Note: telegram-test-api doesn't support editMessageText (used by cbVoteToggle
 * for in-place button updates), so visual icon flip can't be verified here.
 * The VoteStore state and API-level vote persistence ARE verified end-to-end.
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

  const organizer = harness.createUser('ToggleOrganizer');
  await organizer.sendCommand('/start');
  await organizer.sleep(1000);

  voter = harness.createUser('ToggleVoter');
  await voter.sendCommand('/start');
  await voter.sleep(1000);

  const orgToken = await api.getToken(organizer.userId, organizer.name);
  const voterEmail = `tg_${voter.userId}@telegram.local`;
  const users = await api.searchUsers(orgToken, voter.name);
  const voterUser = users.find(u => u.email === voterEmail);

  const meeting = await api.createMeeting(orgToken, {
    title: 'Toggle Vote Meeting',
    description: 'Test vote toggle and submit',
    isPublic: true,
    tags: [],
    timeSlots: [
      { startTime: '2027-03-05T09:00:00Z', endTime: '2027-03-05T10:00:00Z' },
      { startTime: '2027-03-06T14:00:00Z', endTime: '2027-03-06T15:00:00Z' },
    ],
    participantEmails: [],
    participantIds: voterUser ? [voterUser.id] : [],
  });
  meetingId = meeting.id;
});

afterAll(async () => {
  await harness.stop();
});

describe('Vote Toggle and Submit', () => {
  test('vote UI shows toggle buttons with unchecked icons', async () => {
    await voter.sendCommand('/meetings');
    await voter.pressButtonByText('Toggle Vote Meeting');
    await voter.pressButtonByText('Vote on Times');

    const resp = await voter.getLastResponse();
    expect(resp.text).toContain('Vote');

    // Collect vtog buttons
    const vtogButtons = [];
    if (resp.reply_markup?.inline_keyboard) {
      for (const row of resp.reply_markup.inline_keyboard) {
        for (const btn of row) {
          if (btn.callback_data && btn.callback_data.includes('vtog')) {
            vtogButtons.push(btn);
          }
        }
      }
    }
    expect(vtogButtons.length).toBe(2);
    // Both should initially show unchecked icon
    expect(vtogButtons[0].text).toContain('⬜');
    expect(vtogButtons[1].text).toContain('⬜');
  });

  test('toggle a slot and submit votes', async () => {
    // Find the first vtog button
    const history = await voter.getHistory();
    let firstVtogData = null;
    for (let i = history.length - 1; i >= 0; i--) {
      const markup = history[i].reply_markup;
      if (markup?.inline_keyboard) {
        for (const row of markup.inline_keyboard) {
          for (const btn of row) {
            if (btn.callback_data && btn.callback_data.includes('vtog')) {
              firstVtogData = btn.callback_data;
              break;
            }
          }
          if (firstVtogData) break;
        }
        break;
      }
    }
    expect(firstVtogData).not.toBeNull();

    // Toggle the slot
    await voter.pressButton(firstVtogData);
    await voter.sleep(500);

    // Submit votes
    await voter.pressButtonByText('Submit Votes');
    const responses = await voter.getLastResponses(3);
    const hasConfirm = responses.some(r => r.text.includes('Votes submitted'));
    expect(hasConfirm).toBe(true);
  });

  test('verify votes are recorded via API', async () => {
    const token = await api.getToken(voter.userId, voter.name);
    const meeting = await api.getMeeting(token, meetingId);
    const totalVotes = meeting.timeSlots.reduce((sum, s) => sum + s.voteCount, 0);
    expect(totalVotes).toBeGreaterThan(0);
  });
});
