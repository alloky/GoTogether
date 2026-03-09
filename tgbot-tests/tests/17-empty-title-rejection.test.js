/**
 * CJM: Empty Title Rejection
 *
 * Tests that the bot rejects an empty title and re-prompts,
 * then accepts a valid title and completes the flow.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let user;
let api;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();
  user = harness.createUser('EmptyTitleUser');
  await user.sendCommand('/start');
  await user.sleep(1000);
});

afterAll(async () => {
  await harness.stop();
});

describe('Empty Title Rejection', () => {
  test('/skip on title step is rejected', async () => {
    await user.sendCommand('/new');
    await user.expectResponseContaining('title');

    // Try to skip the title
    await user.sendCommand('/skip');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Title cannot be empty');
  });

  test('valid title after rejection is accepted', async () => {
    await user.sendMessage('Valid Title After Rejection');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('description');
  });

  test('complete the meeting creation', async () => {
    await user.sendCommand('/skip');
    await user.expectResponseContaining('visibility');

    await user.pressButtonByText('Public');
    await user.expectResponseContaining('time slot');

    await user.sendMessage('2026-08-20 10:00 - 11:00');
    await user.expectResponseContaining('Time slot 1 added');

    await user.sendCommand('/done');
    await user.expectResponseContaining('tags');

    await user.sendCommand('/skip');
    await user.expectResponseContaining('Invite participants');

    await user.sendCommand('/done');
    await user.sleep(1000);

    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Meeting created');
  });

  test('verify meeting has the valid title via API', async () => {
    const token = await api.getToken(user.userId, user.name);
    const meetings = await api.listMyMeetings(token);
    const meeting = meetings.find(m => m.title === 'Valid Title After Rejection');
    expect(meeting).toBeDefined();
  });
});
