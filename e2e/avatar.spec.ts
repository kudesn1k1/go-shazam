import { test, expect } from '@playwright/test';
import path from 'path';
import { fileURLToPath } from 'url';
import { resetServer } from './helpers/reset';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

test.describe('Avatar upload', () => {
  test.beforeEach(async () => {
    await resetServer();
  });

  test('user can upload PNG and confirm it as their avatar', async ({ page }) => {
    // Register
    await page.goto('/');
    await page.getByRole('button', { name: 'Create account' }).click();
    await page.locator('#email').fill('avatar-user@example.com');
    await page.locator('#password').fill('strong-password-123');
    await page.locator('#confirm').fill('strong-password-123');
    await page.getByRole('button', { name: 'Create account' }).last().click();
    await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible({ timeout: 10000 });

    // Go to profile
    await page.goto('/profile');
    await expect(page.getByRole('heading', { name: 'Profile' })).toBeVisible();

    // The AvatarUploader has a hidden <input type="file"> inside a <label class="pill">Choose photo</label>
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(path.resolve(__dirname, 'fixtures', 'test-avatar.png'));

    // After pick → state goes to 'uploading' then 'previewing'.
    // Wait for the "Use this photo" button to appear (state='previewing').
    await expect(page.getByRole('button', { name: 'Use this photo' })).toBeVisible({ timeout: 10000 });

    // Confirm
    await page.getByRole('button', { name: 'Use this photo' }).click();

    // After confirm the AvatarUploader emits 'changed' → ProfilePage refetches the user.
    // The committed avatar's UserAvatar should now show a real <img src> pointing to /api/files/...
    await expect(page.locator('img[src*="/api/files/"]')).toBeVisible({ timeout: 10000 });
  });
});
