import { createApiClient, type User } from '@novelity/api-client';
export type { User } from '@novelity/api-client';
export { googleLoginUrl } from '@novelity/api-client';
const api = createApiClient();

// Current session, filled by loadSession(). `checked` turns true once the
// server has answered, so the shell can wait instead of flashing.
export const session = $state<{ user: User | null; checked: boolean }>({
	user: null,
	checked: false
});

export async function loadSession(): Promise<User | null> {
	try {
		session.user = await api.me();
	} catch {
		session.user = null;
	}
	session.checked = true;
	return session.user;
}

export async function logout(): Promise<void> {
	await api.logout();
	session.user = null;
}
