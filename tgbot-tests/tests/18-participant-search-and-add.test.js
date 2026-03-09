/**
 * CJM: Participant Search and Add
 *
 * Tests the multi-step participant search flow during meeting creation:
 * search for users, add them via inline buttons, finish with /done.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let organizer;
let api;
let participant1;
let participant2;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  organizer = harness.createUser('SearchOrganizer');
  await organizer.sendCommand('/start');
  await organizer.sleep(1000);

  // Create two additional users with unique names (include harness user ID to avoid DB collisions)
  participant1 = harness.createUser('Findable' + harness.nextUserId);
  await participant1.sendCommand('/start');
  await participant1.sleep(1000);

  participant2 = harness.createUser('Findable' + harness.nextUserId);
  await participant2.sendCommand('/start');
  await participant2.sleep(1000);
});

afterAll(async () => {
  await harness.stop();
});

describe('Participant Search and Add', () => {
  test('create meeting up to participant search step', async () => {
    await organizer.sendCommand('/new');
    await organizer.expectResponseContaining('title');

    await organizer.sendMessage('Meeting With Participants');
    await organizer.expectResponseContaining('description');

    await organizer.sendCommand('/skip');
    await organizer.expectResponseContaining('visibility');

    await organizer.pressButtonByText('Public');
    await organizer.expectResponseContaining('time slot');

    await organizer.sendMessage('2026-09-20 14:00 - 15:00');
    await organizer.expectResponseContaining('Time slot 1 added');

    await organizer.sendCommand('/done');
    await organizer.expectResponseContaining('tags');

    await organizer.sendCommand('/skip');
    await organizer.expectResponseContaining('Invite participants');
  });

  test('search for first participant returns results with Add buttons', async () => {
    await organizer.sendMessage(participant1.name);
    await organizer.sleep(500);
    const resp = await organizer.getLastResponse();
    expect(resp.text).toContain(participant1.name);

    // Should have an Add button for the user
    const addBtn = await organizer.findButton(participant1.name);
    expect(addBtn).not.toBeNull();

    // Should have a Done adding button
    const doneBtn = await organizer.findButton('Done adding');
    expect(doneBtn).not.toBeNull();
  });

  test('add first participant', async () => {
    await organizer.pressButtonByText(participant1.name);
    const resp = await organizer.getLastResponse();
    expect(resp.text).toContain('Participant added');
    expect(resp.text).toContain('1 total');
  });

  test('search and add second participant', async () => {
    await organizer.sendMessage(participant2.name);
    await organizer.sleep(500);
    const resp = await organizer.getLastResponse();
    expect(resp.text).toContain(participant2.name);

    await organizer.pressButtonByText(participant2.name);
    const resp2 = await organizer.getLastResponse();
    expect(resp2.text).toContain('Participant added');
    expect(resp2.text).toContain('2 total');
  });

  test('finish creation with /done', async () => {
    await organizer.sendCommand('/done');
    await organizer.sleep(1000);
    const resp = await organizer.getLastResponse();
    expect(resp.text).toContain('Meeting created');
    expect(resp.text).toContain('2 participant');
  });

  test('verify participants are stored via API', async () => {
    const token = await api.getToken(organizer.userId, organizer.name);
    const meetings = await api.listMyMeetings(token);
    const listed = meetings.find(m => m.title === 'Meeting With Participants');
    expect(listed).toBeDefined();

    const meeting = await api.getMeeting(token, listed.id);
    expect(meeting.participants).toHaveLength(2);

    // Verify participant display names match the ones we added
    const names = meeting.participants.map(p => p.user?.displayName);
    expect(names).toContain(participant1.name);
    expect(names).toContain(participant2.name);
  });
});
