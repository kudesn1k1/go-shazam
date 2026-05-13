import { test, expect } from '@playwright/test';
import { resetServer, seedSong, promote } from './helpers/reset';
import { getLatestEmailHash, countSongs } from './helpers/db';

test.describe('Admin CRUD on songs', () => {
  test.beforeEach(async () => {
    await resetServer();
  });

  test('non-admin user is redirected away from /songs', async ({ page }) => {
    // Register a regular user
    await page.goto('/');
    await page.getByRole('button', { name: 'Create account' }).click();
    await page.locator('#email').fill('regular@example.com');
    await page.locator('#password').fill('strong-password-123');
    await page.locator('#confirm').fill('strong-password-123');
    await page.getByRole('button', { name: 'Create account' }).last().click();
    await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible({ timeout: 10000 });

    // Try to go to admin route — guard should send us back to "/"
    await page.goto('/songs');
    await expect(page).toHaveURL('http://localhost/');
  });

  test('admin lists all songs and deletes one', async ({ page }) => {
    // Register a user, then promote them to admin via the test endpoint
    await page.goto('/');
    await page.getByRole('button', { name: 'Create account' }).click();
    await page.locator('#email').fill('admin-crud@example.com');
    await page.locator('#password').fill('strong-password-123');
    await page.locator('#confirm').fill('strong-password-123');
    await page.getByRole('button', { name: 'Create account' }).last().click();
    await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible({ timeout: 10000 });

    const emailHash = await getLatestEmailHash();
    await promote(emailHash, 'admin');

    // Seed 3 songs to delete
    await seedSong({ title: 'Delete Me 1', artist: 'X', source_id: 'd1' });
    await seedSong({ title: 'Delete Me 2', artist: 'X', source_id: 'd2' });
    await seedSong({ title: 'Delete Me 3', artist: 'X', source_id: 'd3' });

    expect(await countSongs()).toBe(3);

    // Refresh /me so the client-side useAuth picks up the new admin role.
    // The cleanest way is to log out and log back in.
    await page.getByRole('button', { name: 'Log out' }).click();
    await expect(page.getByRole('button', { name: 'Log in' })).toBeVisible();
    await page.getByRole('button', { name: 'Log in' }).click();
    await page.locator('#email').fill('admin-crud@example.com');
    await page.locator('#password').fill('strong-password-123');
    await page.getByRole('button', { name: 'Sign in' }).click();
    await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible();

    // Now navigate to admin /songs
    await page.goto('/songs');
    await expect(page).toHaveURL(/\/songs$/);
    await expect(page.locator('table.data-table tbody tr')).toHaveCount(3);

    // Accept the window.confirm dialog automatically
    page.on('dialog', (dialog) => dialog.accept());

    // Click the first row's Delete button
    await page.getByRole('button', { name: 'Delete' }).first().click();

    // Toast appears; row count drops to 2
    await expect(page.locator('table.data-table tbody tr')).toHaveCount(2, { timeout: 5000 });
    expect(await countSongs()).toBe(2);
  });
});
