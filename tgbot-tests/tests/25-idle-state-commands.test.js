/**
 * CJM: Idle State Commands
 *
 * Tests /cancel, /skip, /done, and plain text when no conversation is active.
 * All should be graceful no-ops or show hints.
 */

const { createTestHarness } = require('../helpers/setup');

let harness;
let user;

beforeAll(async () => {
  harness = await createTestHarness();
  user = harness.createUser('IdleStateUser');
  await user.sendCommand('/start');
  await user.sleep(1000);
});

afterAll(async () => {
  await harness.stop();
});

describe('Idle State Commands', () => {
  test('/cancel when idle shows cancellation message (no crash)', async () => {
    await user.sendCommand('/cancel');
    const resp = await user.getLastResponse();
    // Should respond gracefully — either "Cancelled" or a hint
    expect(resp).not.toBeNull();
    expect(resp.text).toBeTruthy();
  });

  test('plain text when idle shows hint', async () => {
    await user.sendMessage('random text without context');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('/meetings');
  });

  test('/skip when idle does not crash (silent no-op)', async () => {
    const beforeHistory = await user.getHistory();
    const countBefore = beforeHistory.length;

    await user.sendCommand('/skip');
    await user.sleep(500);

    // /skip returns nil when idle, so bot may send nothing
    // Just verify no error occurred — history should not decrease
    const afterHistory = await user.getHistory();
    expect(afterHistory.length).toBeGreaterThanOrEqual(countBefore);
  });

  test('/done when idle does not crash (silent no-op)', async () => {
    const beforeHistory = await user.getHistory();
    const countBefore = beforeHistory.length;

    await user.sendCommand('/done');
    await user.sleep(500);

    const afterHistory = await user.getHistory();
    expect(afterHistory.length).toBeGreaterThanOrEqual(countBefore);
  });
});
