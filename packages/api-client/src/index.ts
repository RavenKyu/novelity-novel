export type Role = "reader" | "author" | "admin";
export interface User {
  id: string;
  email: string;
  name: string;
  picture: string;
  roles: Role[];
  createdAt: string;
  lastLoginAt: string;
}

export class ApiError extends Error {
  status: number;
  constructor(status: number) {
    super(`API request failed: ${status}`);
    this.status = status;
  }
}

// SSR callers supply their own fetch and trusted API base; cookie forwarding stays with the app.
export function createApiClient(fetcher: typeof fetch = fetch, base = "/api") {
  async function request(path: string, init?: RequestInit): Promise<Response> {
    const response = await fetcher(`${base}${path}`, init);
    if (!response.ok) throw new ApiError(response.status);
    return response;
  }
  return {
    async me(): Promise<User | null> {
      const response = await fetcher(`${base}/auth/me`);
      if (response.status === 401) return null;
      if (!response.ok) throw new ApiError(response.status);
      return response.json() as Promise<User>;
    },
    async logout(): Promise<void> {
      await request("/auth/logout", { method: "POST" });
    },
    async health(): Promise<{ status: string; db?: string }> {
      return (await request("/healthz")).json();
    },
  };
}

export function googleLoginUrl(returnTo = "/"): string {
  return `/api/auth/google/login?${new URLSearchParams({ returnTo })}`;
}

// Only local paths are accepted by client navigation. The server validates cross-space origins.
export function localReturnTo(value: string | null): string {
  if (
    !value ||
    !value.startsWith("/") ||
    value.startsWith("//") ||
    /[\\\x00-\x20\x7f]/.test(value)
  )
    return "/";
  try {
    const url = new URL(value, "https://local.invalid");
    if (url.origin !== "https://local.invalid" || url.pathname === "/login")
      return "/";
    return url.pathname + url.search + url.hash;
  } catch {
    return "/";
  }
}
