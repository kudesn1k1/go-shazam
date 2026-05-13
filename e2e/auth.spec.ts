import { test, expect } from '@playwright/test';
import { resetServer } from './helpers/reset';

test.describe('Authentication flow', () => {
  test.beforeEach(async () => {
    await resetServer();
  });

  test('register, reload restores session, logout clears it', async ({ page }) => {
    await page.goto('/');

    // Open Create-account modal
    await page.getByRole('button', { name: 'Create account' }).click();
    await page.locator('#email').fill('e2e-auth@example.com');
    await page.locator('#password').fill('correct-horse-battery-staple');
    await page.locator('#confirm').fill('correct-horse-battery-staple');
    await page.getByRole('button', { name: 'Create account' }).last().click();

    // Authenticated UI: Log out button visible
    await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible({ timeout: 10000 });

    // Reload — session must restore via httpOnly refresh cookie
    await page.reload();
    await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible({ timeout: 10000 });

    // Logout
    await page.getByRole('button', { name: 'Log out' }).click();
    await expect(page.getByRole('button', { name: 'Create account' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Log in' })).toBeVisible();

    // Reload again — must NOT be logged in
    await page.reload();
    await expect(page.getByRole('button', { name: 'Create account' })).toBeVisible();
  });

  test('login with wrong password shows error toast', async ({ page }) => {
    // First register a user
    await page.goto('/');
    await page.getByRole('button', { name: 'Create account' }).click();
    await page.locator('#email').fill('wrong-pw@example.com');
    await page.locator('#password').fill('right-password-123');
    await page.locator('#confirm').fill('right-password-123');
    await page.getByRole('button', { name: 'Create account' }).last().click();
    await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible({ timeout: 10000 });

    // Logout and try wrong password
    await page.getByRole('button', { name: 'Log out' }).click();
    await expect(page.getByRole('button', { name: 'Log in' })).toBeVisible();

    await page.getByRole('button', { name: 'Log in' }).click();
    await page.locator('#email').fill('wrong-pw@example.com');
    await page.locator('#password').fill('definitely-not-the-right-password');
    await page.getByRole('button', { name: 'Sign in' }).click();

    // Toast appears with error
    await expect(page.locator('.toast')).toBeVisible({ timeout: 5000 });
    // Still NOT logged in
    await expect(page.getByRole('button', { name: 'Log out' })).toHaveCount(0);
  });
});
