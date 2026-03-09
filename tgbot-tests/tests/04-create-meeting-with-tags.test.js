/**
 * CJM: Create Meeting with Custom Tags
 *
 * Verifies that tags are properly associated with the meeting.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let user;
let api;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();
  user = harness.createUser('TagCreator');
  await user.sendCommand('/start');
  await user.sleep(1000);
});

afterAll(async () => {
  await harness.stop();
});

describe('Create Meeting with Tags', () => {
  test('create meeting with multiple comma-separated tags', async () => {
    await user.sendCommand('/new');
    await user.expectResponseContaining('title');

    await user.sendMessage('Design Review');
    await user.expectResponseContaining('description');

    await user.sendMessage('Review new UI mockups');
    await user.expectResponseContaining('visibility');

    await user.pressButtonByText('Public');
    await user.expectResponseContaining('time slot');

    await user.sendMessage('2026-08-10 14:00 - 16:00');
    await user.expectResponseContaining('Time slot 1 added');

    await user.sendCommand('/done');
    await user.expectResponseContaining('tags');

    // Send comma-separated tags
    await user.sendMessage('design, review, ui, frontend');
    await user.expectResponseContaining('Invite participants');

    // Finish
    await user.sendCommand('/done');
    await user.sleep(1000);

    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Meeting created');
    expect(resp.text).toContain('Design Review');
  });

  test('verify tags are stored via API', async () => {
    const token = await api.getToken(user.userId, user.name);
    const meetings = await api.listMyMeetings(token);
    const meeting = meetings.find(m => m.title === 'Design Review');
    expect(meeting).toBeDefined();
    expect(meeting.tags).toHaveLength(4);
    expect(meeting.tags).toContain('design');
    expect(meeting.tags).toContain('review');
    expect(meeting.tags).toContain('ui');
    expect(meeting.tags).toContain('frontend');
  });
});
