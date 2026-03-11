/**
 * CJM: Back Button Navigation
 *
 * Tests "Back to Meetings" on meeting detail, and "Main Menu" buttons
 * on meetings list, calendar, and help — verifying every menu has a way back.
 */

const { createTestHarness } = require('../helpers/setup');
const { BackendAPI } = require('../helpers/api');

let harness;
let user;
let api;

beforeAll(async () => {
  harness = await createTestHarness();
  api = new BackendAPI();

  user = harness.createUser('BackNavUser');
  await user.sendCommand('/start');
  await user.sleep(1000);

  // Create a meeting so the meetings list is non-empty
  const token = await api.getToken(user.userId, user.name);
  await api.createMeeting(token, {
    title: 'Back Nav Meeting',
    description: 'Testing back navigation',
    isPublic: true,
    tags: [],
    timeSlots: [
      { startTime: '2027-04-10T10:00:00Z', endTime: '2027-04-10T11:00:00Z' },
    ],
    participantEmails: [],
    participantIds: [],
  });
});

afterAll(async () => {
  await harness.stop();
});

describe('Back Button — Meeting Detail', () => {
  test('meeting detail has a Back to Meetings button', async () => {
    await user.sendCommand('/meetings');
    await user.pressButtonByText('Back Nav Meeting');
    const detail = await user.getLastResponse();
    expect(detail.text).toContain('Back Nav Meeting');

    const backBtn = await user.findButton('Back to Meetings');
    expect(backBtn).not.toBeNull();
  });

  test('pressing Back to Meetings returns to meetings list', async () => {
    await user.pressButtonByText('Back to Meetings');
    const list = await user.getLastResponse();
    expect(list.text).toContain('Your Meetings');
    expect(list.text).toContain('Back Nav Meeting');
  });

  test('can navigate detail → back → detail → back repeatedly', async () => {
    await user.pressButtonByText('Back Nav Meeting');
    expect((await user.getLastResponse()).text).toContain('Back Nav Meeting');

    await user.pressButtonByText('Back to Meetings');
    expect((await user.getLastResponse()).text).toContain('Your Meetings');
  });
});

describe('Back Button — Meetings List', () => {
  test('meetings list has a Main Menu button', async () => {
    await user.sendCommand('/meetings');
    const btn = await user.findButton('Main Menu');
    expect(btn).not.toBeNull();
  });

  test('pressing Main Menu from meetings list shows welcome message', async () => {
    await user.pressButtonByText('Main Menu');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('GoTogether');
    // Main menu buttons are present
    const myMeetingsBtn = await user.findButton('My Meetings');
    expect(myMeetingsBtn).not.toBeNull();
  });
});

describe('Back Button — Calendar', () => {
  test('calendar has a Main Menu button', async () => {
    await user.sendCommand('/calendar');
    const btn = await user.findButton('Main Menu');
    expect(btn).not.toBeNull();
  });

  test('pressing Main Menu from calendar shows welcome message', async () => {
    await user.pressButtonByText('Main Menu');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('GoTogether');
    const calBtn = await user.findButton('Calendar');
    expect(calBtn).not.toBeNull();
  });
});

describe('Back Button — Help', () => {
  test('help message has a Main Menu button', async () => {
    await user.sendCommand('/help');
    const btn = await user.findButton('Main Menu');
    expect(btn).not.toBeNull();
  });

  test('pressing Main Menu from help shows welcome message', async () => {
    await user.pressButtonByText('Main Menu');
    const resp = await user.getLastResponse();
    expect(resp.text).toContain('GoTogether');
    const helpBtn = await user.findButton('Help');
    expect(helpBtn).not.toBeNull();
  });
});
