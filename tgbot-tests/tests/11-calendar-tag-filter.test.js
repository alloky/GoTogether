/**
 * CJM: Calendar Tag Filter
 *
 * User filters calendar events by tag using inline buttons.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let user;
let api;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  user = harness.createUser('TagFilterUser');
  await user.sendCommand('/start');
  await user.sleep(1000);

  const token = await api.getToken(user.userId, user.name);

  // Create two confirmed meetings with different tags
  const m1 = await api.createMeeting(token, {
    title: 'Work Meeting',
    isPublic: true,
    tags: ['work'],
    timeSlots: [
      { startTime: '2027-03-10T10:00:00Z', endTime: '2027-03-10T11:00:00Z' },
    ],
    participantEmails: [],
    participantIds: [],
  });
  await api.confirmMeeting(token, m1.id, m1.timeSlots[0].id);

  const m2 = await api.createMeeting(token, {
    title: 'Social Event',
    isPublic: true,
    tags: ['social'],
    timeSlots: [
      { startTime: '2027-03-11T18:00:00Z', endTime: '2027-03-11T20:00:00Z' },
    ],
    participantEmails: [],
    participantIds: [],
  });
  await api.confirmMeeting(token, m2.id, m2.timeSlots[0].id);
});

afterAll(async () => {
  await harness.stop();
});

describe('Calendar Tag Filter', () => {
  test('/calendar shows all confirmed meetings', async () => {
    await user.sendCommand('/calendar');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Calendar');
    expect(resp.text).toContain('Work Meeting');
    expect(resp.text).toContain('Social Event');
  });

  test('calendar has tag filter buttons', async () => {
    const workBtn = await user.findButton('work');
    const socialBtn = await user.findButton('social');
    // At least one tag button should exist
    expect(workBtn || socialBtn).toBeTruthy();
  });

  test('filtering by work tag shows only work meetings', async () => {
    const workBtnData = await user.findButton('work');
    expect(workBtnData).not.toBeNull();
    await user.pressButton(workBtnData);
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Work Meeting');
    expect(resp.text).toContain('Filtered by');
    // Social Event should NOT appear when filtered by work
    expect(resp.text).not.toContain('Social Event');
  });

  test('clearing filter shows all meetings again', async () => {
    const clearBtn = await user.findButton('Clear filter');
    expect(clearBtn).not.toBeNull();
    await user.pressButton(clearBtn);
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('Work Meeting');
    expect(resp.text).toContain('Social Event');
    expect(resp.text).not.toContain('Filtered by');
  });
});
