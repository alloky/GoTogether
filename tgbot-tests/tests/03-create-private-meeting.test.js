/**
 * CJM: Create a Private Meeting
 *
 * Same flow as public, but selects "Private" visibility.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let user;
let api;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();
  user = harness.createUser('PrivateCreator');
  await user.sendCommand('/start');
  await user.sleep(1000);
});

afterAll(async () => {
  await harness.stop();
});

describe('Create Private Meeting', () => {
  test('create private meeting with minimal info', async () => {
    // Start creation
    await user.sendCommand('/new');
    await user.expectResponseContaining('title');

    // Title
    await user.sendMessage('Secret Planning Session');
    await user.expectResponseContaining('description');

    // Skip description
    await user.sendCommand('/skip');
    await user.expectResponseContaining('visibility');

    // Choose Private
    await user.pressButtonByText('Private');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Private');
    expect(resp.text).toContain('time slot');

    // Add one slot
    await user.sendMessage('2026-07-01 09:00 - 10:00');
    await user.expectResponseContaining('Time slot 1 added');

    // Done with slots
    await user.sendCommand('/done');
    await user.expectResponseContaining('tags');

    // Skip tags
    await user.sendCommand('/skip');
    await user.expectResponseContaining('Invite participants');

    // Skip participants
    await user.sendCommand('/done');
    await user.sleep(1000);

    const final = await user.getLastResponse();
    expect(final.text).toContain('Meeting created');
    expect(final.text).toContain('Secret Planning Session');
    expect(final.text).toContain('Private');
  });

  test('verify private meeting is stored correctly via API', async () => {
    const token = await api.getToken(user.userId, user.name);
    const meetings = await api.listMyMeetings(token);
    const listed = meetings.find(m => m.title === 'Secret Planning Session');
    expect(listed).toBeDefined();

    const meeting = await api.getMeeting(token, listed.id);
    expect(meeting.isPublic).toBe(false);
    expect(meeting.timeSlots).toHaveLength(1);
  });
});
