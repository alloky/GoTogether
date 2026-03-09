import { Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class LoginPage extends BasePage {
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly loginButton: Locator;
  readonly errorMessage: Locator;
  readonly registerLink: Locator;

  constructor(page: any) {
    super(page);
    this.emailInput = this.getByTestId('login-email');
    this.passwordInput = this.getByTestId('login-password');
    this.loginButton = this.getByTestId('login-submit');
    this.errorMessage = this.getByTestId('login-error');
    this.registerLink = this.page.getByRole('link', { name: 'Register' });
  }

  async goto() {
    await this.page.goto('/login');
  }

  async login(email: string, password: string) {
    await this.emailInput.fill(email);
    await this.passwordInput.fill(password);
    await this.loginButton.click();
  }
}
