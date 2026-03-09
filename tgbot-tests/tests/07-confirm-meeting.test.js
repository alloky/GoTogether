/**
 * CJM: Confirm Meeting Time Slot (Manual Selection)
 *
 * Organizer manually picks a time slot to confirm the meeting.
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

  organizer = harness.createUser('ConfirmOrganizer');
  await organizer.sendCommand('/start');
  await organizer.sleep(1000);

  // Create meeting with 2 time slots
  const token = await api.getToken(organizer.userId, organizer.name);
  const meeting = await api.createMeeting(token, {
    title: 'Confirm Test Meeting',
    description: 'Meeting to test confirmation',
    isPublic: true,
    tags: ['confirm-test'],
    timeSlots: [
      { startTime: '2026-11-01T10:00:00Z', endTime: '2026-11-01T11:00:00Z' },
      { startTime: '2026-11-02T14:00:00Z', endTime: '2026-11-02T15:00:00Z' },
    ],
    participantEmails: [],
    participantIds: [],
  });
  meetingId = meeting.id;

  // Vote on a slot so there's data
  await api.vote(token, meetingId, [meeting.timeSlots[1].id]);
});

afterAll(async () => {
  await harness.stop();
});

describe('Confirm Meeting (Manual Slot Selection)', () => {
  test('organizer navigates to meeting detail', async () => {
    await organizer.sendCommand('/meetings');
    await organizer.pressButtonByText('Confirm Test Meeting');
    const resp = await organizer.getLastResponse();
    expect(resp.text).toContain('Confirm Test Meeting');
    expect(resp.text).toContain('organizer');
  });

  test('organizer sees Confirm button', async () => {
    const confirmBtn = await organizer.findButton('Confirm Meeting');
    expect(confirmBtn).not.toBeNull();
  });

  test('organizer opens confirm interface', async () => {
    await organizer.pressButtonByText('Confirm Meeting');
    const resp = await organizer.getLastResponse();
    expect(resp.text).toContain('Confirm');
    expect(resp.text).toContain('Confirm Test Meeting');

    // Should show Auto-pick and individual slots
    const autoBtn = await organizer.findButton('Auto-pick');
    expect(autoBtn).not.toBeNull();
  });

  test('organizer manually picks a specific slot', async () => {
    // Pick the second time slot (which has a vote)
    const history = await organizer.getHistory();
    const lastMsg = history[history.length - 1];
    let slotBtnData = null;

    if (lastMsg.reply_markup?.inline_keyboard) {
      for (const row of lastMsg.reply_markup.inline_keyboard) {
        for (const btn of row) {
          // Find a cfsl button that is NOT auto
          if (btn.callback_data && btn.callback_data.includes('|') && !btn.callback_data.includes('auto')) {
            slotBtnData = btn.callback_data;
            break;
          }
        }
        if (slotBtnData) break;
      }
    }

    expect(slotBtnData).not.toBeNull();
    await organizer.pressButton(slotBtnData);
    await organizer.sleep(1000);

    // Should see confirmation message
    const responses = await organizer.getLastResponses(3);
    const hasConfirm = responses.some(r => r.text.includes('Meeting confirmed'));
    expect(hasConfirm).toBe(true);
  });

  test('verify meeting is confirmed via API', async () => {
    const token = await api.getToken(organizer.userId, organizer.name);
    const meeting = await api.getMeeting(token, meetingId);
    expect(meeting.status).toBe('confirmed');
    expect(meeting.confirmedSlotId).toBeDefined();
  });
});
