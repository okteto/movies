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

test('catalog has entries', async ({ request }) => {
    const apiUrl = `https://movies-${process.env.OKTETO_NAMESPACE}.${process.env.OKTETO_DOMAIN}/catalog`;
    const response = await request.get(apiUrl);
    expect(response.status()).toBe(200);
    const data = await response.json();
    expect(data.length).toBe(6);

    const expectedTitles = [
    'Moby Dock',
    'The Finalizer',
    'Crash Loop Backoff',
    'Kube',
    'Cloud Atlas',
    'Aliens'
  ];

  const actualTitles = data.map(item => item.original_title);
  expect(actualTitles).toEqual(expectedTitles);  
});