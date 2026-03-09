import { test, expect } from '@playwright/test';
import { RegisterPage } from '../pages/RegisterPage';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { CreateMeetingPage } from '../pages/CreateMeetingPage';
import { MeetingDetailPage } from '../pages/MeetingDetailPage';
import { uniqueEmail } from '../fixtures/test-users';

test.describe('Voting Flow', () => {
  test('organizer can create meeting, vote, and confirm', async ({ page }) => {
    // Register as organizer
    const registerPage = new RegisterPage(page);
    await registerPage.goto();
    const email = uniqueEmail('organizer');
    await registerPage.register('Organizer', email, 'password123');
    await expect(page).toHaveURL(/\/dashboard/);

    // Create meeting
    const dashboard = new DashboardPage(page);
    await dashboard.clickNewMeeting();

    const createPage = new CreateMeetingPage(page);
    await createPage.fillTitle('Vote Test Meeting');
    await createPage.addTimeSlot('2026-04-01 10:00', '2026-04-01 11:00');
    await createPage.submit();

    await expect(page).toHaveURL(/\/meetings\/[a-f0-9-]+/);

    // Vote on slot
    const detailPage = new MeetingDetailPage(page);
    await expect(detailPage.timeSlotVoting).toBeVisible();
    await detailPage.voteOnSlots([0]);

    // Wait for vote to be submitted
    await page.waitForTimeout(1000);

    // Confirm meeting
    await detailPage.confirmMeeting();

    // Status should change to confirmed
    await expect(detailPage.meetingStatus).toHaveText('confirmed');
  });
});
