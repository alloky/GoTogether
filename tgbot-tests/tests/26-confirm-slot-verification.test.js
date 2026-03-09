/**
 * CJM: Confirm Meeting Slot Verification (Enhanced)
 *
 * Tests that confirming a meeting with a specific slot correctly stores
 * the exact chosen slot ID in the backend, not just any slot.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let organizer;
let api;
let meetingId;
let slotIds;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  organizer = harness.createUser('SlotVerifyOrganizer');
  await organizer.sendCommand('/start');
  await organizer.sleep(1000);

  const token = await api.getToken(organizer.userId, organizer.name);
  const meeting = await api.createMeeting(token, {
    title: 'Slot Verify Meeting',
    description: 'Verify confirmed slot matches selection',
    isPublic: true,
    tags: [],
    timeSlots: [
      { startTime: '2027-06-01T09:00:00Z', endTime: '2027-06-01T10:00:00Z' },
      { startTime: '2027-06-02T14:00:00Z', endTime: '2027-06-02T15:00:00Z' },
      { startTime: '2027-06-03T16:00:00Z', endTime: '2027-06-03T17:00:00Z' },
    ],
    participantEmails: [],
    participantIds: [],
  });
  meetingId = meeting.id;
  slotIds = meeting.timeSlots.map(s => s.id);

  // Vote on the second slot to make it distinguishable
  await api.vote(token, meetingId, [slotIds[1]]);
});

afterAll(async () => {
  await harness.stop();
});

describe('Confirm Meeting Slot Verification', () => {
  test('organizer navigates to confirm interface', async () => {
    await organizer.sendCommand('/meetings');
    await organizer.pressButtonByText('Slot Verify Meeting');
    await organizer.pressButtonByText('Confirm Meeting');

    const resp = await organizer.getLastResponse();
    expect(resp.text).toContain('Confirm');
    expect(resp.text).toContain('Slot Verify Meeting');
  });

  test('confirm view shows individual slot options with vote counts', async () => {
    const autoBtn = await organizer.findButton('Auto-pick');
    expect(autoBtn).not.toBeNull();

    // Should show slot buttons with vote info
    const resp = await organizer.getLastResponse();
    expect(resp.reply_markup).not.toBeNull();

    const allButtons = resp.reply_markup.inline_keyboard.flat();
    // Filter for cfsl buttons (confirm slot) that aren't auto-pick
    const cfslButtons = allButtons.filter(b =>
      b.callback_data && b.callback_data.includes('cfsl') && !b.callback_data.includes('auto')
    );
    expect(cfslButtons.length).toBe(3);
  });

  test('organizer picks the second slot (with 1 vote)', async () => {
    // Find the button that shows "1 vote" (the second slot)
    const resp = await organizer.getLastResponse();
    let targetBtnData = null;
    for (const row of resp.reply_markup.inline_keyboard) {
      for (const btn of row) {
        if (btn.text && btn.text.includes('1 vote') &&
            btn.callback_data && !btn.callback_data.includes('auto')) {
          targetBtnData = btn.callback_data;
          break;
        }
      }
      if (targetBtnData) break;
    }

    expect(targetBtnData).not.toBeNull();
    await organizer.pressButton(targetBtnData);
    await organizer.sleep(1000);

    const responses = await organizer.getLastResponses(3);
    const hasConfirm = responses.some(r => r.text.includes('Meeting confirmed'));
    expect(hasConfirm).toBe(true);
  });

  test('verify the exact confirmed slot via API', async () => {
    const token = await api.getToken(organizer.userId, organizer.name);
    const meeting = await api.getMeeting(token, meetingId);

    expect(meeting.status).toBe('confirmed');
    expect(meeting.confirmedSlotId).toBeDefined();
    expect(meeting.confirmedSlotId).not.toBeNull();
    // The confirmed slot should be the second one (which had the vote)
    expect(meeting.confirmedSlotId).toBe(slotIds[1]);
  });
});
