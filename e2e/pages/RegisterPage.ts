import { Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class RegisterPage extends BasePage {
  readonly nameInput: Locator;
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly registerButton: Locator;
  readonly errorMessage: Locator;
  readonly loginLink: Locator;

  constructor(page: any) {
    super(page);
    this.nameInput = this.getByTestId('register-name');
    this.emailInput = this.getByTestId('register-email');
    this.passwordInput = this.getByTestId('register-password');
    this.registerButton = this.getByTestId('register-submit');
    this.errorMessage = this.getByTestId('register-error');
    this.loginLink = this.page.getByRole('link', { name: 'Log in' });
  }

  async goto() {
    await this.page.goto('/register');
  }

  async register(name: string, email: string, password: string) {
    await this.nameInput.fill(name);
    await this.emailInput.fill(email);
    await this.passwordInput.fill(password);
    await this.registerButton.click();
  }
}
