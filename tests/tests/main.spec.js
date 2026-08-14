import { test, expect, request as playwrightRequest } from '@playwright/test';

const baseURL = () => `https://movies-${process.env.OKTETO_NAMESPACE}.${process.env.OKTETO_DOMAIN}`;

const session = async (email) => {
  const context = await playwrightRequest.newContext({ baseURL: baseURL() });
  const response = await context.post('/auth/login', { data: { email } });
  expect(response.status()).toBe(200);
  return context;
};

const adminSession = async () => {
  const context = await playwrightRequest.newContext({ baseURL: baseURL() });
  const response = await context.post('/auth/admin-login', { data: { username: 'admin', password: 'admin123' } });
  expect(response.status()).toBe(200);
  return context;
};

const movie = async (context, id) => {
  const response = await context.get('/availability');
  expect(response.status()).toBe(200);
  const movies = await response.json();
  return movies.find(item => String(item.id) === String(id));
};

// rentals go through kafka, so the worker applies them asynchronously
const waitForRental = async (context, id, rented) => {
  await expect.poll(async () => (await movie(context, id)).rented, { timeout: 30000 }).toBe(rented);
};

test('environment variables are set', async () => {
  expect(process.env.OKTETO_NAMESPACE).toBeDefined();
  expect(process.env.OKTETO_DOMAIN).toBeDefined();
  expect(process.env.OKTETO_NAMESPACE).not.toBe('');
  expect(process.env.OKTETO_DOMAIN).not.toBe('');
});


test('movies has title', async ({ page }) => {
  await page.goto(baseURL());

  // The page title
  await expect(page).toHaveTitle('Movies');
});

test('catalog has entries', async ({ request }) => {
    const response = await request.get(`${baseURL()}/catalog`);
    expect(response.status()).toBe(200);
    const data = await response.json();
    expect(data.length).toBeGreaterThanOrEqual(6);

    const expectedTitles = [
    'Moby Dock',
    'The Finalizer',
    'Crash Loop Backoff',
    'Kube',
    'Cloud Atlas',
    'Aliens'
  ];

  const actualTitles = data.map(item => item.original_title);
  expect(actualTitles).toEqual(expect.arrayContaining(expectedTitles));
});

test('every movie in the catalog has a number of copies', async ({ request }) => {
  const response = await request.get(`${baseURL()}/catalog`);
  const data = await response.json();
  data.forEach(item => expect(item.copies).toBeGreaterThan(0));
});

test('users share the catalog and are limited by the number of copies', async () => {
  const stamp = Date.now();
  const one = await session(`one-${stamp}@example.com`);
  const two = await session(`two-${stamp}@example.com`);

  // a movie of its own, so the test does not depend on what the rest of the namespace rented
  const admin = await adminSession();
  const created = await admin.post('/catalog', {
    data: { original_title: `Single Copy ${stamp}`, overview: 'only one copy', price: 1, vote_average: 5, copies: 1 }
  });
  expect(created.status()).toBe(201);
  const { id } = await created.json();

  const target = await movie(one, id);
  expect(target.copies).toBe(1);
  expect(target.available).toBe(1);

  expect((await one.post('/rent', { data: { catalog_id: String(id) } })).status()).toBe(202);
  await waitForRental(one, id, true);

  const denied = await two.post('/rent', { data: { catalog_id: String(id) } });
  expect(denied.status()).toBe(409);
  expect(await denied.text()).toContain('copies');
  expect((await movie(two, id)).available).toBe(0);

  expect((await one.post('/rent/return', { data: { catalog_id: String(id) } })).status()).toBe(202);
  await waitForRental(one, id, false);

  expect((await two.post('/rent', { data: { catalog_id: String(id) } })).status()).toBe(202);
  await waitForRental(two, id, true);
  expect((await one.post('/rent', { data: { catalog_id: String(id) } })).status()).toBe(409);

  // the returned movie stays in the history of the first user, charged at the catalog price
  const history = await (await one.get('/rentals/history')).json();
  const returned = history.filter(rental => rental.movie_id === String(id) && rental.returned_at);
  expect(returned.length).toBe(1);
  expect(returned[0].price).toBe(1);

  expect((await two.post('/rent/return', { data: { catalog_id: String(id) } })).status()).toBe(202);
  await waitForRental(two, id, false);

  expect((await admin.delete(`/catalog/${id}`)).status()).toBe(200);
});

test('anonymous users cannot rent', async ({ request }) => {
  const response = await request.post(`${baseURL()}/rent`, { data: { catalog_id: '1' } });
  expect(response.status()).toBe(401);
});

test('an admin can see the history of a user, ban them and forgive them', async ({ page }) => {
  const email = `banned-${Date.now()}@example.com`;
  const user = await session(email);
  expect((await user.post('/rent', { data: { catalog_id: '6' } })).status()).toBe(202);
  await waitForRental(user, 6, true);

  const admin = await adminSession();
  const rentals = await (await admin.get(`/adminapi/users/${encodeURIComponent(email)}/rentals`)).json();
  expect(rentals.length).toBe(1);
  expect(rentals[0].movie_id).toBe('6');

  expect((await admin.post(`/adminapi/users/${encodeURIComponent(email)}/ban`, { data: { reason: 'testing' } })).status()).toBe(200);
  expect((await user.post('/rent', { data: { catalog_id: '1' } })).status()).toBe(409);

  // the banned user gets a meme and a chance to redeem themselves
  await page.goto(baseURL());
  await page.getByPlaceholder('you@example.com').fill(email);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await expect(page.getByRole('heading', { name: 'You have been banned' })).toBeVisible();
  await expect(page.locator('.Banned__meme')).toBeVisible();
  await page.locator('.Banned__input').fill('I helped a friend fix their cluster');
  await page.getByRole('button', { name: 'Request forgiveness' }).click();
  await expect(page.locator('.Banned__thanks')).toBeVisible();

  const redemptions = await (await admin.get('/adminapi/redemptions')).json();
  const redemption = redemptions.find(item => item.user_email === email);
  expect(redemption.status).toBe('pending');

  expect((await admin.post(`/adminapi/redemptions/${redemption.id}/resolve`, { data: { status: 'approved' } })).status()).toBe(200);
  expect((await (await user.get('/me')).json()).banned).toBe(false);

  expect((await user.post('/rent/return', { data: { catalog_id: '6' } })).status()).toBe(202);
  await waitForRental(user, 6, false);
});

test('an admin can add, edit and remove movies from the catalog', async ({ request }) => {
  const admin = await adminSession();
  const title = `Test Movie ${Date.now()}`;

  const anonymous = await request.post(`${baseURL()}/catalog`, { data: { original_title: title } });
  expect(anonymous.status()).toBe(401);

  const created = await admin.post('/catalog', {
    data: { original_title: title, overview: 'a movie about testing', price: 1.5, vote_average: 7, copies: 2 }
  });
  expect(created.status()).toBe(201);
  const { id } = await created.json();

  const user = await session(`catalog-${Date.now()}@example.com`);
  expect((await movie(user, id)).copies).toBe(2);

  expect((await admin.put(`/catalog/${id}`, { data: { original_title: title, price: 3, copies: 5 } })).status()).toBe(200);
  expect((await movie(user, id)).copies).toBe(5);

  expect((await admin.delete(`/catalog/${id}`)).status()).toBe(200);
  expect(await movie(user, id)).toBeUndefined();
});
