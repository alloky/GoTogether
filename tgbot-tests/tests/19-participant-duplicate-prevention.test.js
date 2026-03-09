/**
 * CJM: Participant Duplicate Prevention
 *
 * Tests that adding the same user twice shows a duplicate warning
 * and the participant appears only once in the final meeting.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let organizer;
let participant;
let api;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  organizer = harness.createUser('DupOrganizer');
  await organizer.sendCommand('/start');
  await organizer.sleep(1000);

  participant = harness.createUser('DupTarget');
  await participant.sendCommand('/start');
  await participant.sleep(1000);
});

afterAll(async () => {
  await harness.stop();
});

describe('Participant Duplicate Prevention', () => {
  test('create meeting and reach participant step', async () => {
    await organizer.sendCommand('/new');
    await organizer.expectResponseContaining('title');

    await organizer.sendMessage('Dup Prevention Meeting');
    await organizer.expectResponseContaining('description');

    await organizer.sendCommand('/skip');
    await organizer.expectResponseContaining('visibility');

    await organizer.pressButtonByText('Public');
    await organizer.expectResponseContaining('time slot');

    await organizer.sendMessage('2026-09-25 10:00 - 11:00');
    await organizer.expectResponseContaining('Time slot 1 added');

    await organizer.sendCommand('/done');
    await organizer.expectResponseContaining('tags');

    await organizer.sendCommand('/skip');
    await organizer.expectResponseContaining('Invite participants');
  });

  test('add participant first time succeeds', async () => {
    await organizer.sendMessage('DupTarget');
    await organizer.sleep(500);
    await organizer.pressButtonByText('DupTarget');
    const resp = await organizer.getLastResponse();
    expect(resp.text).toContain('Participant added');
    expect(resp.text).toContain('1 total');
  });

  test('adding same participant again shows duplicate warning', async () => {
    await organizer.sendMessage('DupTarget');
    await organizer.sleep(500);
    await organizer.pressButtonByText('DupTarget');
    const resp = await organizer.getLastResponse();
    expect(resp.text).toContain('Already added');
  });

  test('finish creation and verify only one participant via API', async () => {
    await organizer.sendCommand('/done');
    await organizer.sleep(1000);
    const resp = await organizer.getLastResponse();
    expect(resp.text).toContain('Meeting created');
    expect(resp.text).toContain('1 participant');

    const token = await api.getToken(organizer.userId, organizer.name);
    const meetings = await api.listMyMeetings(token);
    const listed = meetings.find(m => m.title === 'Dup Prevention Meeting');
    expect(listed).toBeDefined();

    const meeting = await api.getMeeting(token, listed.id);
    expect(meeting.participants).toHaveLength(1);
  });
});
