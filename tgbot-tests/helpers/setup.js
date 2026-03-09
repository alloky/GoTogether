/**
 * Per-test-file setup. Each test file creates its own harness instance.
 * User IDs are randomized to avoid conflicts across test files.
 */

const { BotTestHarness } = require('./harness');

async function createTestHarness() {
  const harness = new BotTestHarness();
  // Randomize starting IDs to avoid conflicts across test files
  harness.nextUserId = 1000 + Math.floor(Math.random() * 90000);
  harness.nextChatId = harness.nextUserId;
  await harness.start();
  return harness;
}

module.exports = { createTestHarness };
