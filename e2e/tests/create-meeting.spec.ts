import { test, expect } from '@playwright/test';
import { RegisterPage } from '../pages/RegisterPage';
import { DashboardPage } from '../pages/DashboardPage';
import { CreateMeetingPage } from '../pages/CreateMeetingPage';
import { uniqueEmail } from '../fixtures/test-users';

test.describe('Create Meeting', () => {
  test('should create a new meeting and see it on dashboard', async ({ page }) => {
    // Register
    const registerPage = new RegisterPage(page);
    await registerPage.goto();
    await registerPage.register('Meeting Creator', uniqueEmail('creator'), 'password123');
    await expect(page).toHaveURL(/\/dashboard/);

    // Navigate to create meeting
    const dashboard = new DashboardPage(page);
    await dashboard.clickNewMeeting();
    await expect(page).toHaveURL(/\/meetings\/new/);

    // Fill form
    const createPage = new CreateMeetingPage(page);
    await createPage.fillTitle('Team Standup');
    await createPage.fillDescription('Daily standup meeting');

    // Add a time slot
    await createPage.addTimeSlot('2026-03-15 09:00', '2026-03-15 09:30');

    // Submit
    await createPage.submit();

    // Should redirect to meeting detail
    await expect(page).toHaveURL(/\/meetings\/[a-f0-9-]+/);
    await expect(page.getByText('Team Standup')).toBeVisible();
  });

  test('should show empty state on fresh dashboard', async ({ page }) => {
    const registerPage = new RegisterPage(page);
    await registerPage.goto();
    await registerPage.register('Fresh User', uniqueEmail('fresh'), 'password123');
    await expect(page).toHaveURL(/\/dashboard/);

    const dashboard = new DashboardPage(page);
    await expect(dashboard.emptyState).toBeVisible();
  });
});
