import { test, expect } from '@playwright/test';
import { RegisterPage } from '../pages/RegisterPage';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { CreateMeetingPage } from '../pages/CreateMeetingPage';
import { MeetingDetailPage } from '../pages/MeetingDetailPage';
import { uniqueEmail } from '../fixtures/test-users';

test.describe('Full End-to-End Flow', () => {
  test('complete meeting lifecycle: create, invite, vote, confirm', async ({ browser }) => {
    test.setTimeout(90000);
    const testId = Math.random().toString(36).slice(2, 7);
    const bobName = `BobTest-${testId}`;
    const bobEmail = uniqueEmail('bob');

    // === User A (Organizer): Register ===
    const ctxA = await browser.newContext();
    const pageA = await ctxA.newPage();

    const registerA = new RegisterPage(pageA);
    await registerA.goto();
    await registerA.register('Alice Organizer', uniqueEmail('alice'), 'password123');
    await expect(pageA).toHaveURL(/\/dashboard/);

    // === User B: Register ===
    const ctxB = await browser.newContext();
    const pageB = await ctxB.newPage();

    const registerB = new RegisterPage(pageB);
    await registerB.goto();
    await registerB.register(bobName, bobEmail, 'password123');
    await expect(pageB).toHaveURL(/\/dashboard/);

    // === User A: Create meeting with Bob ===
    const dashboardA = new DashboardPage(pageA);
    await dashboardA.clickNewMeeting();

    const createPage = new CreateMeetingPage(pageA);
    await createPage.fillTitle('Full Flow Meeting');
    await createPage.fillDescription('Testing the complete flow');
    await createPage.addTimeSlot('2026-05-01 14:00', '2026-05-01 15:00');
    await createPage.searchAndAddParticipant(bobName);
    await createPage.submit();

    await expect(pageA).toHaveURL(/\/meetings\/[a-f0-9-]+/);

    // Get meeting URL
    const meetingUrl = pageA.url();
    const meetingId = meetingUrl.split('/').pop();

    // === User A: Vote ===
    const detailA = new MeetingDetailPage(pageA);
    await expect(detailA.timeSlotVoting).toBeVisible();
    await detailA.voteOnSlots([0]);
    await pageA.waitForTimeout(1000);

    // === User B: Navigate to the meeting ===
    await pageB.goto(meetingUrl);

    const detailB = new MeetingDetailPage(pageB);
    await expect(detailB.timeSlotVoting).toBeVisible({ timeout: 10000 });
    await detailB.voteOnSlots([0]);
    await pageB.waitForTimeout(1000);

    // === User A: Confirm meeting ===
    await pageA.reload();
    await expect(detailA.confirmButton).toBeVisible();
    await detailA.confirmMeeting();

    await expect(detailA.meetingStatus).toHaveText('confirmed');

    // === User B: See confirmed status ===
    await pageB.reload();
    await expect(detailB.meetingStatus).toHaveText('confirmed');

    await ctxA.close();
    await ctxB.close();
  });
});
