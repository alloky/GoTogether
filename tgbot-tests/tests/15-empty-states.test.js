/**
 * CJM: Empty State Handling
 *
 * Tests what happens when a fresh user has no meetings.
 * Note: calendar shows ALL public confirmed meetings, so it may not be empty.
 */

const { createTestHarness } = require('../helpers/setup');

let harness;
let user;

beforeAll(async () => {
  harness = await createTestHarness();
  user = harness.createUser('EmptyStateUser');
  await user.sendCommand('/start');
  await user.sleep(1000);
});

afterAll(async () => {
  await harness.stop();
});

describe('Empty States', () => {
  test('/meetings with no meetings shows empty message', async () => {
    await user.sendCommand('/meetings');
    const resp = await user.getLastResponse();
    // Should show "no meetings" or similar
    expect(
      resp.text.includes('no meetings') || resp.text.includes('/new')
    ).toBe(true);
  });

  test('/calendar shows calendar header', async () => {
    await user.sendCommand('/calendar');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Calendar');
  });
});
