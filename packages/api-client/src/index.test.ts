import test from "node:test";
import assert from "node:assert/strict";
import {
  createApiClient,
  ApiError,
  googleLoginUrl,
  localReturnTo,
} from "./index.ts";

test("anonymous me and server failure are distinct", async () => {
  assert.equal(
    await createApiClient(async () => new Response(null, { status: 401 })).me(),
    null,
  );
  await assert.rejects(
    createApiClient(async () => new Response(null, { status: 503 })).me(),
    (error: unknown) => error instanceof ApiError && error.status === 503,
  );
});

test("logout uses POST against the supplied API base and reports failures", async () => {
  let requested = "";
  let method = "";
  const client = createApiClient(async (url, init) => {
    requested = String(url);
    method = init?.method ?? "";
    return new Response(null, { status: 204 });
  }, "https://studio.test/api");
  await client.logout();
  assert.equal(requested, "https://studio.test/api/auth/logout");
  assert.equal(method, "POST");
  await assert.rejects(
    createApiClient(async () => new Response(null, { status: 500 })).logout(),
  );
});

test("local navigation rejects external and malformed return paths", () => {
  for (const value of [
    null,
    "https://evil.test/",
    "//evil.test/",
    "/\\evil.test/",
    "/login",
    "/\n/evil.test",
  ])
    assert.equal(localReturnTo(value), "/");
  assert.equal(
    localReturnTo("/works?tab=draft#chapter"),
    "/works?tab=draft#chapter",
  );
  assert.equal(
    new URL(
      googleLoginUrl("/works?x=1"),
      "https://studio.test",
    ).searchParams.get("returnTo"),
    "/works?x=1",
  );
});
