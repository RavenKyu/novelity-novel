import { test, expect } from "./fixtures";

test("signed-in user sees the studio shell and API health", async ({
  page,
  user,
}) => {
  await page.goto("/");
  await expect(page.locator("aside").getByText(user.name)).toBeVisible();
  await expect(page.locator("aside").getByText(user.email)).toBeVisible();
  await expect(page.getByText("정상")).toBeVisible();
});

test("signed-in user visiting /login is sent home", async ({ page, user }) => {
  await page.goto("/login");
  await expect(page).toHaveURL(/\/$/);
  await expect(page.locator("aside").getByText(user.name)).toBeVisible();
});

test("logout ends the session", async ({ page, user }) => {
  await page.goto("/");
  await page.locator("aside").getByText(user.name).click();
  await page.getByRole("menuitem", { name: "로그아웃" }).click();

  await expect(page).toHaveURL(/\/login$/);
  expect((await page.request.get("/api/auth/me")).status()).toBe(401);
});

test("login preserves a local returnTo and the OAuth login link", async ({
  page,
}) => {
  await page.goto("/login?returnTo=%2F%3Ftab%3Ddraft%23chapter");
  const link = page.getByRole("link", { name: "Google로 로그인" });
  const href = await link.getAttribute("href");
  const target = new URL(href!, page.url());
  expect(target.searchParams.get("returnTo")).toBe("/?tab=draft#chapter");
});

test("signed-in login returnTo preserves query and fragment", async ({
  page,
  user,
}) => {
  await page.goto("/login?returnTo=%2F%3Ftab%3Ddraft%23chapter");
  await expect(page).toHaveURL(/\/\?tab=draft#chapter$/);
  await expect(page.locator("aside").getByText(user.name)).toBeVisible();
});

test("public deployment has no admin API routes", async ({ page }) => {
  expect((await page.request.get("/api/admin/healthz")).status()).toBe(404);
  expect((await page.request.get("/api/store/healthz")).status()).toBe(200);
});

// The API stubbed here is not under test: only the SPA hand-off is. The API-side
// validation lives in the Go test TestExistingSessionReturnsAcrossSpacesWithoutGoogle.
test("signed-in login forwards a cross-space returnTo to the API unchanged", async ({
  page,
  user,
}) => {
  const returnTo = "https://novelity.kr/library";
  await page.route("**/api/auth/google/login?*", async (route) => {
    const target = new URL(route.request().url());
    expect(target.searchParams.get("returnTo")).toBe(returnTo);
    await route.fulfill({
      status: 200,
      contentType: "text/html",
      body: "<p>validated return</p>",
    });
  });
  await page.goto("/login?" + new URLSearchParams({ returnTo }));
  await expect(page.getByText("validated return")).toBeVisible();
});
