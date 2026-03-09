/**
 * CJM: Back Button Navigation
 *
 * Tests the "Back to Meetings" button on the meeting detail page
 * and verifies it navigates back to the meetings list.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let user;
let api;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  user = harness.createUser('BackNavUser');
  await user.sendCommand('/start');
  await user.sleep(1000);

  // Create a meeting so user has something to navigate to
  const token = await api.getToken(user.userId, user.name);
  await api.createMeeting(token, {
    title: 'Back Nav Meeting',
    description: 'Testing back navigation',
    isPublic: true,
    tags: [],
    timeSlots: [
      { startTime: '2027-04-10T10:00:00Z', endTime: '2027-04-10T11:00:00Z' },
    ],
    participantEmails: [],
    participantIds: [],
  });
});

afterAll(async () => {
  await harness.stop();
});

describe('Back Button Navigation', () => {
  test('navigate to meeting detail', async () => {
    await user.sendCommand('/meetings');
    await user.pressButtonByText('Back Nav Meeting');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Back Nav Meeting');
    expect(resp.text).toContain('Testing back navigation');
  });

  test('pressing Back returns to meetings list', async () => {
    const backBtn = await user.findButton('Back to Meetings');
    expect(backBtn).not.toBeNull();

    await user.pressButtonByText('Back to Meetings');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Your Meetings');
    expect(resp.text).toContain('Back Nav Meeting');
  });

  test('can navigate back to detail and back again', async () => {
    // Go back into detail
    await user.pressButtonByText('Back Nav Meeting');
    const detail = await user.getLastResponse();
    expect(detail.text).toContain('Back Nav Meeting');

    // Go back to list again
    await user.pressButtonByText('Back to Meetings');
    const list = await user.getLastResponse();
    expect(list.text).toContain('Your Meetings');
  });
});
