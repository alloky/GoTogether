/**
 * CJM: Meetings List Pagination
 *
 * Creates 7 meetings for a user, then verifies that the meetings list
 * paginates at 5 per page with Prev/Next buttons.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let user;
let api;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  user = harness.createUser('PaginationUser');
  await user.sendCommand('/start');
  await user.sleep(1000);

  // Create 7 meetings via API
  const token = await api.getToken(user.userId, user.name);
  for (let i = 1; i <= 7; i++) {
    await api.createMeeting(token, {
      title: `Page Meeting ${i}`,
      isPublic: true,
      tags: [],
      timeSlots: [
        {
          startTime: `2027-05-${String(i).padStart(2, '0')}T10:00:00Z`,
          endTime: `2027-05-${String(i).padStart(2, '0')}T11:00:00Z`,
        },
      ],
      participantEmails: [],
      participantIds: [],
    });
  }
});

afterAll(async () => {
  await harness.stop();
});

describe('Meetings Pagination', () => {
  test('first page shows 5 meetings and Next button', async () => {
    await user.sendCommand('/meetings');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Your Meetings');

    // Count meeting buttons (each meeting is a row button)
    const buttons = resp.reply_markup.inline_keyboard.flat();
    const meetingBtns = buttons.filter(b => b.text.includes('Page Meeting'));
    expect(meetingBtns.length).toBe(5);

    // Should have Next button
    const nextBtn = buttons.find(b => b.text.includes('Next'));
    expect(nextBtn).toBeDefined();

    // Should NOT have Prev button on first page
    const prevBtn = buttons.find(b => b.text.includes('Prev'));
    expect(prevBtn).toBeUndefined();
  });

  test('pressing Next shows second page with remaining meetings', async () => {
    await user.pressButtonByText('Next');
    const resp = await user.getLastResponse();

    const buttons = resp.reply_markup.inline_keyboard.flat();
    const meetingBtns = buttons.filter(b => b.text.includes('Page Meeting'));
    expect(meetingBtns.length).toBe(2);

    // Should have Prev button
    const prevBtn = buttons.find(b => b.text.includes('Prev'));
    expect(prevBtn).toBeDefined();

    // Should NOT have Next button on last page
    const nextBtn = buttons.find(b => b.text.includes('Next'));
    expect(nextBtn).toBeUndefined();
  });

  test('pressing Prev returns to first page', async () => {
    await user.pressButtonByText('Prev');
    const resp = await user.getLastResponse();

    const buttons = resp.reply_markup.inline_keyboard.flat();
    const meetingBtns = buttons.filter(b => b.text.includes('Page Meeting'));
    expect(meetingBtns.length).toBe(5);

    const nextBtn = buttons.find(b => b.text.includes('Next'));
    expect(nextBtn).toBeDefined();
  });

  test('verify total meeting count via API', async () => {
    const token = await api.getToken(user.userId, user.name);
    const meetings = await api.listMyMeetings(token);
    const pageMeetings = meetings.filter(m => m.title.startsWith('Page Meeting'));
    expect(pageMeetings.length).toBe(7);
  });
});
