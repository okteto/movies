import { test, expect } from '@playwright/test';

test('environment variables are set', async () => {
  expect(process.env.OKTETO_NAMESPACE).toBeDefined();
  expect(process.env.OKTETO_DOMAIN).toBeDefined();
  expect(process.env.OKTETO_NAMESPACE).not.toBe('');
  expect(process.env.OKTETO_DOMAIN).not.toBe('');
});


test('movies has title', async ({ page }) => {
  await page.goto(`https://movies-${process.env.OKTETO_NAMESPACE}.${process.env.OKTETO_DOMAIN}`);

  // The page title
  await expect(page).toHaveTitle('Movies');
});