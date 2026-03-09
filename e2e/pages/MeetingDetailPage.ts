import { Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class MeetingDetailPage extends BasePage {
  readonly meetingDetail: Locator;
  readonly meetingStatus: Locator;
  readonly confirmButton: Locator;
  readonly submitVotesButton: Locator;
  readonly rsvpAcceptButton: Locator;
  readonly rsvpDeclineButton: Locator;
  readonly backButton: Locator;
  readonly participantList: Locator;
  readonly timeSlotVoting: Locator;

  constructor(page: any) {
    super(page);
    this.meetingDetail = this.getByTestId('meeting-detail');
    this.meetingStatus = this.getByTestId('meeting-status');
    this.confirmButton = this.getByTestId('confirm-meeting');
    this.submitVotesButton = this.getByTestId('submit-votes');
    this.rsvpAcceptButton = this.getByTestId('rsvp-accept');
    this.rsvpDeclineButton = this.getByTestId('rsvp-decline');
    this.backButton = this.getByTestId('back-to-dashboard');
    this.participantList = this.getByTestId('participant-list');
    this.timeSlotVoting = this.getByTestId('time-slot-voting');
  }

  async getStatus() {
    return await this.meetingStatus.textContent();
  }

  async voteOnSlots(indices: number[]) {
    const checkboxes = this.timeSlotVoting.locator('.ant-checkbox-input');
    for (const idx of indices) {
      await checkboxes.nth(idx).check();
    }
    await this.submitVotesButton.click();
  }

  async confirmMeeting() {
    await this.confirmButton.click();
  }

  async goBack() {
    await this.backButton.click();
  }
}
