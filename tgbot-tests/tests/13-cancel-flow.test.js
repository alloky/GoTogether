/**
 * CJM: Cancel Meeting Creation Flow
 *
 * User starts creating a meeting, then cancels mid-flow.
 */

const { createTestHarness } = require('../helpers/setup');

let harness;
let user;

beforeAll(async () => {
  harness = await createTestHarness();
  user = harness.createUser('CancelUser');
  await user.sendCommand('/start');
  await user.sleep(1000);
});

afterAll(async () => {
  await harness.stop();
});

describe('Cancel Meeting Creation', () => {
  test('start creation and cancel after title step', async () => {
    await user.sendCommand('/new');
    await user.expectResponseContaining('title');

    await user.sendCommand('/cancel');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Cancelled');
  });

  test('after cancel, text input shows hint instead of continuing flow', async () => {
    await user.sendMessage('random text');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('/meetings');
  });

  test('cancel at description step', async () => {
    await user.sendCommand('/new');
    await user.sendMessage('Some Title');
    await user.expectResponseContaining('description');

    await user.sendCommand('/cancel');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Cancelled');
  });
});
