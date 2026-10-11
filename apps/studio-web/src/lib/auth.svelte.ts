export interface User {
	id: string;
	email: string;
	name: string;
	picture: string;
}

// Current session, filled by loadSession(). `checked` turns true once the
// server has answered, so the shell can wait instead of flashing.
export const session = $state<{ user: User | null; checked: boolean }>({
	user: null,
	checked: false
});

export async function loadSession(): Promise<User | null> {
	try {
		const res = await fetch('/api/auth/me');
		session.user = res.ok ? ((await res.json()) as User) : null;
	} catch {
		session.user = null;
	}
	session.checked = true;
	return session.user;
}

export async function logout(): Promise<void> {
	await fetch('/api/auth/logout', { method: 'POST' }).catch(() => {});
	session.user = null;
}

// Full-page navigation: the server runs the OAuth redirect chain.
export const googleLoginUrl = '/api/auth/google/login';
