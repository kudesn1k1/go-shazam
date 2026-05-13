import { test, expect } from '@playwright/test';
import { resetServer, seedSong } from './helpers/reset';

test.describe('Catalog browsing', () => {
  test.beforeEach(async () => {
    await resetServer();
    // Seed 30 songs across 3 artists with monotonically increasing duration.
    for (let i = 1; i <= 30; i++) {
      const artist = `Artist ${'ABC'[i % 3]}`;
      await seedSong({
        title: `Song ${String(i).padStart(2, '0')}`,
        artist,
        duration: 180000 + i * 1000,
        source_id: `src-${i}`,
      });
    }
  });

  test('catalog shows seeded songs with default pagination', async ({ page }) => {
    await page.goto('/catalog');

    // 20 rows per page by default
    const rows = page.locator('table.data-table tbody tr');
    await expect(rows).toHaveCount(20);

    // Pagination control should be visible (30 items > 20 per page)
    await expect(page.locator('.pagination')).toBeVisible();
  });

  test('clicking next page navigates to page 2 with remaining 10 rows', async ({ page }) => {
    await page.goto('/catalog');
    await expect(page.locator('table.data-table tbody tr')).toHaveCount(20);

    // Last page-btn is the "›" next button
    const buttons = page.locator('.pagination .page-btn');
    await buttons.last().click();

    await expect(page.locator('table.data-table tbody tr')).toHaveCount(10);
    await expect(page).toHaveURL(/page=2/);
  });

  test('filtering by artist narrows the result set', async ({ page }) => {
    await page.goto('/catalog');
    await expect(page.locator('table.data-table tbody tr')).toHaveCount(20);

    // SongFilters.vue exposes an artist input — find by placeholder or by label
    // First look for an input named "artist"; fall back to typing into the
    // generic search box.
    const artistInput = page.locator('input[placeholder*="rtist" i], input[name="artist"]').first();
    if (await artistInput.count() > 0) {
      await artistInput.fill('Artist A');
      // Wait for refetch — table updates without page reload.
      await page.waitForTimeout(500);
      const count = await page.locator('table.data-table tbody tr').count();
      expect(count).toBeGreaterThan(0);
      expect(count).toBeLessThanOrEqual(10);
    } else {
      // No artist filter UI? Use the URL directly.
      await page.goto('/catalog?artist=Artist+A');
      const count = await page.locator('table.data-table tbody tr').count();
      expect(count).toBeGreaterThan(0);
      expect(count).toBeLessThanOrEqual(10);
    }
  });
});
