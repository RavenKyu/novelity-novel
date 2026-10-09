import { test as base, type BrowserContext } from '@playwright/test';
import { createHash, randomBytes } from 'node:crypto';
import { execFileSync } from 'node:child_process';

/**
 * Real Google sign-in can't run unattended, so specs that need a logged-in user
 * insert a user + session straight into MongoDB (the same shape core-api's
 * store writes: session _id = sha256(token)) and hand the browser the cookie.
 * Everything seeded uses googleSub "e2e-…" and is removed after the test.
 */

const mongoUser = process.env.MONGO_ROOT_USER ?? 'novelity';
const mongoPassword = process.env.MONGO_ROOT_PASSWORD ?? 'novelity-dev';

function mongosh(script: string): string {
	return execFileSync(
		'docker',
		['compose', 'exec', '-T', 'nn-mongodb', 'mongosh', '-u', mongoUser, '-p', mongoPassword, '--quiet', '--eval', script],
		{ encoding: 'utf8' }
	).trim();
}

export interface SeededUser {
	sub: string;
	email: string;
	name: string;
	token: string;
}

export function seedSession(name = 'E2E 작가'): SeededUser {
	const sub = `e2e-${randomBytes(6).toString('hex')}`;
	const email = `${sub}@example.com`;
	const token = randomBytes(32).toString('base64url');
	const hash = createHash('sha256').update(token).digest('hex');
	mongosh(`
		const d = db.getSiblingDB('novelity');
		const now = new Date();
		const u = d.users.findOneAndUpdate(
			{ googleSub: ${JSON.stringify(sub)} },
			{ $set: { email: ${JSON.stringify(email)}, name: ${JSON.stringify(name)}, picture: '', lastLoginAt: now },
			  $setOnInsert: { createdAt: now } },
			{ upsert: true, returnDocument: 'after' });
		d.sessions.insertOne({ _id: ${JSON.stringify(hash)}, userId: u._id, createdAt: now,
			expiresAt: new Date(now.getTime() + 3600e3) });
	`);
	return { sub, email, name, token };
}

export function removeSeeded(sub: string): void {
	mongosh(`
		const d = db.getSiblingDB('novelity');
		const u = d.users.findOne({ googleSub: ${JSON.stringify(sub)} });
		if (u) { d.sessions.deleteMany({ userId: u._id }); d.users.deleteOne({ _id: u._id }); }
	`);
}

export async function signIn(context: BrowserContext, baseURL: string, user: SeededUser): Promise<void> {
	await context.addCookies([
		{ name: 'nn_session', value: user.token, url: baseURL, httpOnly: true, sameSite: 'Lax' }
	]);
}

/** `user` is a seeded, signed-in user for the test's browser context. */
export const test = base.extend<{ user: SeededUser }>({
	user: async ({ context, baseURL }, use) => {
		const user = seedSession();
		await signIn(context, baseURL!, user);
		await use(user);
		removeSeeded(user.sub);
	}
});

export { expect } from '@playwright/test';
