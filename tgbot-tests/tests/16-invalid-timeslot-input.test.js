/**
 * CJM: Invalid Time Slot Input Handling
 *
 * Tests that the bot properly rejects invalid time slot formats
 * and re-prompts the user without breaking the conversation flow.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let user;
let api;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();
  user = harness.createUser('InvalidSlotUser');
  await user.sendCommand('/start');
  await user.sleep(1000);
});

afterAll(async () => {
  await harness.stop();
});

describe('Invalid Time Slot Input', () => {
  test('start meeting creation and reach time slot step', async () => {
    await user.sendCommand('/new');
    await user.expectResponseContaining('title');

    await user.sendMessage('Slot Validation Meeting');
    await user.expectResponseContaining('description');

    await user.sendCommand('/skip');
    await user.expectResponseContaining('visibility');

    await user.pressButtonByText('Public');
    await user.expectResponseContaining('time slot');
  });

  test('missing separator is rejected', async () => {
    await user.sendMessage('2026-06-15 10:00 11:00');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Invalid');
  });

  test('invalid start time is rejected', async () => {
    await user.sendMessage('2026-13-01 25:00 - 26:00');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('invalid');
  });

  test('end before start is rejected', async () => {
    await user.sendMessage('2026-06-15 14:00 - 10:00');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('end time must be after start time');
  });

  test('valid slot after failures is accepted', async () => {
    await user.sendMessage('2026-06-15 10:00 - 11:00');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Time slot 1 added');
  });

  test('finish meeting creation', async () => {
    await user.sendCommand('/done');
    await user.expectResponseContaining('tags');
    await user.sendCommand('/skip');
    await user.expectResponseContaining('Invite participants');
    await user.sendCommand('/done');
    await user.sleep(1000);
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Meeting created');
  });

  test('verify meeting has exactly 1 time slot via API', async () => {
    const token = await api.getToken(user.userId, user.name);
    const meetings = await api.listMyMeetings(token);
    const listed = meetings.find(m => m.title === 'Slot Validation Meeting');
    expect(listed).toBeDefined();

    const meeting = await api.getMeeting(token, listed.id);
    expect(meeting.timeSlots).toHaveLength(1);
  });
});
