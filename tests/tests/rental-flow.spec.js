import { test, expect } from '@playwright/test';

// The movie used for the rent/return round trip. The flow returns it at the end,
// so the environment is left the way it was found and the test can be re-run.
const MOVIE_ID = 1;
const MOVIE_TITLE = 'Moby Dock';
const RENT_PRICE = 12.99;

// rent -> kafka -> worker -> postgres is asynchronous, so rental state is polled
// instead of being read once.
const PROPAGATION_TIMEOUT = 30_000;

async function rentedMovies(request) {
  const response = await request.get('/rentals');
  expect(response.status()).toBe(200);
  return (await response.json()) ?? [];
}

async function rentedIds(request) {
  return (await rentedMovies(request)).map((movie) => movie.id);
}

test('services report healthy', async ({ request }) => {
  const catalog = await request.get('/catalog/healthz');
  expect(catalog.status()).toBe(200);

  const rent = await request.get('/rent');
  expect(rent.status()).toBe(200);
  expect(await rent.json()).toEqual({ status: 'ok' });
});

test('api serves users from postgres', async ({ request }) => {
  const all = await request.get('/users');
  expect(all.status()).toBe(200);
  expect((await all.json()).length).toBeGreaterThan(0);

  const one = await request.get('/users/1');
  expect(one.status()).toBe(200);
  expect((await one.json()).Userid).toBe(1);

  const missing = await request.get('/users/99999999');
  expect(missing.status()).toBe(404);
});

// Exercises every backend service in one pass: rent (java) publishes to kafka,
// worker (go) consumes and writes to postgres, and api (go) reads it back and
// joins it against catalog (node).
test('renting and returning a movie round trips through every service', async ({ request }) => {
  test.setTimeout(120_000);

  await test.step('start from a clean slate', async () => {
    await request.post('/rent/return', { data: { catalog_id: String(MOVIE_ID) } });
    await expect
      .poll(() => rentedIds(request), { timeout: PROPAGATION_TIMEOUT })
      .not.toContain(MOVIE_ID);
  });

  await test.step('rent publishes the movie to the rentals topic', async () => {
    const response = await request.post('/rent', {
      data: { catalog_id: String(MOVIE_ID), price: RENT_PRICE },
    });
    expect(response.status()).toBe(200);
  });

  await test.step('worker writes the rental to postgres', async () => {
    await expect
      .poll(() => rentedIds(request), { timeout: PROPAGATION_TIMEOUT })
      .toContain(MOVIE_ID);
  });

  await test.step('api joins the rental against the catalog', async () => {
    const rented = (await rentedMovies(request)).find((movie) => movie.id === MOVIE_ID);
    expect(rented.original_title).toBe(MOVIE_TITLE);
    expect(rented.price).toBeCloseTo(RENT_PRICE, 2);
  });

  await test.step('returning the movie removes it again', async () => {
    const response = await request.post('/rent/return', {
      data: { catalog_id: String(MOVIE_ID) },
    });
    expect(response.status()).toBe(200);
    await expect
      .poll(() => rentedIds(request), { timeout: PROPAGATION_TIMEOUT })
      .not.toContain(MOVIE_ID);
  });
});
