import { test, expect } from '@playwright/test';
import { RegisterPage } from '../pages/RegisterPage';
import { LoginPage } from '../pages/LoginPage';
import { uniqueEmail } from '../fixtures/test-users';

test.describe('Authentication', () => {
  test('should register a new user and redirect to dashboard', async ({ page }) => {
    const registerPage = new RegisterPage(page);
    await registerPage.goto();

    const email = uniqueEmail('reg');
    await registerPage.register('Test User', email, 'password123');

    await expect(page).toHaveURL(/\/dashboard/);
  });

  test('should login with registered user', async ({ page }) => {
    const email = uniqueEmail('login');

    // First register
    const registerPage = new RegisterPage(page);
    await registerPage.goto();
    await registerPage.register('Login User', email, 'password123');
    await expect(page).toHaveURL(/\/dashboard/);

    // Logout
    await page.getByTestId('logout-btn').click();
    await expect(page).toHaveURL(/\/login/);

    // Login
    const loginPage = new LoginPage(page);
    await loginPage.login(email, 'password123');
    await expect(page).toHaveURL(/\/dashboard/);
  });

  test('should show error for wrong credentials', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('nonexistent@test.com', 'wrongpassword');

    await expect(loginPage.errorMessage).toBeVisible();
  });

  test('should navigate between login and register', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.registerLink.click();
    await expect(page).toHaveURL(/\/register/);

    const registerPage = new RegisterPage(page);
    await registerPage.loginLink.click();
    await expect(page).toHaveURL(/\/login/);
  });

  test('should redirect to login when not authenticated', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });
});
