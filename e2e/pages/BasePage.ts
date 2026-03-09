import { Page, Locator } from '@playwright/test';

export class BasePage {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  getByTestId(testId: string): Locator {
    return this.page.getByTestId(testId);
  }

  async waitForNavigation(url: string | RegExp) {
    await this.page.waitForURL(url);
  }
}
