/**
 * CJM: User Registration & Welcome
 *
 * Tests the /start command which auto-registers the user
 * and shows the welcome menu with navigation buttons.
 */

const { createTestHarness } = require('../helpers/setup');

let harness;
let user;

beforeAll(async () => {
  harness = await createTestHarness();
});

afterAll(async () => {
  await harness.stop();
});

describe('/start - Registration & Welcome', () => {
  test('should welcome user and show main menu', async () => {
    user = harness.createUser('Alice Start');
    await user.sendCommand('/start');

    const resp = await user.getLastResponse();
    expect(resp).not.toBeNull();
    expect(resp.text).toContain('Welcome');
    expect(resp.text).toContain('GoTogether');
  });

  test('should show inline menu buttons', async () => {
    const resp = await user.getLastResponse();
    expect(resp.reply_markup).not.toBeNull();
    expect(resp.reply_markup.inline_keyboard).toBeDefined();

    const buttons = resp.reply_markup.inline_keyboard
      .flat()
      .map(b => b.text);
    expect(buttons.some(b => b.includes('My Meetings'))).toBe(true);
    expect(buttons.some(b => b.includes('Calendar'))).toBe(true);
    expect(buttons.some(b => b.includes('New Meeting'))).toBe(true);
    expect(buttons.some(b => b.includes('Help'))).toBe(true);
  });

  test('should handle /help command', async () => {
    const helpUser = harness.createUser('HelpUser');
    await helpUser.sendCommand('/start');
    await helpUser.sendCommand('/help');

    const resp = await helpUser.getLastResponse();
    expect(resp.text).toContain('Commands');
    expect(resp.text).toContain('/start');
    expect(resp.text).toContain('/meetings');
    expect(resp.text).toContain('/new');
  });
});
