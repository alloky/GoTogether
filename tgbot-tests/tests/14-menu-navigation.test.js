/**
 * CJM: Menu Navigation via Inline Buttons
 *
 * User navigates through the bot using the main menu inline buttons.
 */

const { createTestHarness } = require('../helpers/setup');

let harness;
let user;

beforeAll(async () => {
  harness = await createTestHarness();
  user = harness.createUser('MenuNavigator');
  await user.sendCommand('/start');
  await user.sleep(1000);
});

afterAll(async () => {
  await harness.stop();
});

describe('Menu Navigation', () => {
  test('pressing My Meetings button shows meetings list', async () => {
    await user.pressButtonByText('My Meetings');
    const resp = await user.getLastResponse();
    // Either shows meetings or empty state
    const hasMeetings = resp.text.includes('Your Meetings') || resp.text.includes('no meetings');
    expect(hasMeetings).toBe(true);
  });

  test('pressing Help button shows help text', async () => {
    // Go back to start
    await user.sendCommand('/start');
    await user.pressButtonByText('Help');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Commands');
  });

  test('pressing New Meeting button starts creation flow', async () => {
    await user.sendCommand('/start');
    await user.pressButtonByText('New Meeting');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('title');
    // Clean up by cancelling
    await user.sendCommand('/cancel');
  });

  test('pressing Calendar button shows calendar', async () => {
    await user.sendCommand('/start');
    await user.pressButtonByText('Calendar');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Calendar');
  });
});
