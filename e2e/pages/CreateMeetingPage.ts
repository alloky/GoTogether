import { Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class CreateMeetingPage extends BasePage {
  readonly titleInput: Locator;
  readonly descriptionInput: Locator;
  readonly addTimeSlotButton: Locator;
  readonly participantSelect: Locator;
  readonly submitButton: Locator;

  constructor(page: any) {
    super(page);
    this.titleInput = this.getByTestId('meeting-title');
    this.descriptionInput = this.getByTestId('meeting-description');
    this.addTimeSlotButton = this.getByTestId('add-time-slot');
    this.participantSelect = this.page.getByRole('combobox', { name: /Invite Participants/ });
    this.submitButton = this.getByTestId('create-meeting-submit');
  }

  async goto() {
    await this.page.goto('/meetings/new');
  }

  async fillTitle(title: string) {
    await this.titleInput.fill(title);
  }

  async fillDescription(description: string) {
    await this.descriptionInput.fill(description);
  }

  async addTimeSlot(startDate: string, endDate: string) {
    const rangePickers = this.page.locator('.ant-picker-range');
    const lastPicker = rangePickers.last();

    const startInput = lastPicker.locator('input').first();
    await startInput.click();
    await startInput.fill(startDate);
    await startInput.press('Enter');

    const endInput = lastPicker.locator('input').last();
    await endInput.fill(endDate);
    await endInput.press('Escape');

    await this.page.waitForTimeout(300);
  }

  async searchAndAddParticipant(name: string) {
    // Click and type into the Select's combobox input
    await this.participantSelect.click();
    await this.participantSelect.fill(name);

    // Wait for debounce (300ms) + API response
    await this.page.waitForTimeout(800);

    // Wait for dropdown options to appear and click the first one
    const option = this.page.locator('.ant-select-item-option').first();
    await option.waitFor({ state: 'visible', timeout: 5000 });
    await option.click();

    // Click outside to close dropdown
    await this.page.waitForTimeout(200);
    await this.titleInput.click();
  }

  async submit() {
    await this.submitButton.scrollIntoViewIfNeeded();
    await Promise.all([
      this.page.waitForURL(/\/meetings\/[a-f0-9-]+/),
      this.submitButton.click(),
    ]);
  }
}
