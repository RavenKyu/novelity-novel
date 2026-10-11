import { ApiError, createApiClient, type User } from '@novelity/api-client';
const api = createApiClient();

export const session = $state<{ user: User | null; checked: boolean; forbidden: boolean }>({
	user: null,
	checked: false,
	forbidden: false
});

export async function loadSession(): Promise<void> {
	session.forbidden = false;
	try {
		const user = await api.me();
		session.forbidden = !!user && !user.roles.includes('admin');
		session.user = user?.roles.includes('admin') ? user : null;
	} catch (error) {
		session.user = null;
		session.forbidden = error instanceof ApiError && error.status === 403;
	}
	session.checked = true;
}

export async function logout(): Promise<void> {
	await api.logout();
	session.user = null;
	session.forbidden = false;
}
