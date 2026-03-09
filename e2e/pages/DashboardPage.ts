import { Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class DashboardPage extends BasePage {
  readonly newMeetingButton: Locator;
  readonly meetingList: Locator;
  readonly emptyState: Locator;

  constructor(page: any) {
    super(page);
    this.newMeetingButton = this.getByTestId('new-meeting-btn');
    this.meetingList = this.getByTestId('meeting-list');
    this.emptyState = this.getByTestId('meetings-empty');
  }

  async goto() {
    await this.page.goto('/dashboard');
  }

  async clickNewMeeting() {
    await this.newMeetingButton.click();
  }

  async clickMeeting(title: string) {
    await this.page.getByText(title).click();
  }

  async getMeetingCount() {
    return await this.meetingList.locator('.ant-list-item').count();
  }
}
