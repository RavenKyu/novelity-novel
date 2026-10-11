<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { loadSession, session } from '$lib/auth.svelte';
	import { Sidebar, UserMenu } from '@novelity/ui';
	import { HomeOutline } from 'flowbite-svelte-icons';
	import { logout } from '$lib/auth.svelte';
	import { googleLoginUrl, localReturnTo } from '@novelity/api-client';

	let { children } = $props();

	// /login renders without the studio shell (argos "chromeless" pattern).
	const isChromeless = $derived(page.url.pathname === '/login');

	$effect(() => {
		if (!session.checked) {
			void loadSession();
		} else if (!session.user && !isChromeless) {
			const target = page.url.pathname + page.url.search + page.url.hash;
			void goto(target === '/' ? '/login' : `/login?${new URLSearchParams({ returnTo: target })}`, {
				replaceState: true
			});
		} else if (session.user && isChromeless) {
			const returnTo = page.url.searchParams.get('returnTo');
			if (returnTo && /^https?:\/\//.test(returnTo)) {
				// The API owns the cross-space origin allowlist and rechecks the live session.
				window.location.assign(googleLoginUrl(returnTo));
			} else {
				void goto(localReturnTo(returnTo), { replaceState: true });
			}
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Novelity Studio</title>
</svelte:head>

{#if isChromeless}
	{#if session.checked && !session.user}
		{@render children()}
	{/if}
{:else if session.user}
	<div class="flex min-h-screen bg-[var(--color-body)]">
		<Sidebar
			title="Novelity Studio"
			version={__APP_VERSION__}
			pathname={page.url.pathname}
			navigation={[{ label: '홈', href: '/', icon: HomeOutline }]}
		>
			{#snippet userMenu()}
				<UserMenu
					user={session.user}
					onLogout={async () => {
						await logout();
						await goto('/login');
					}}
				/>
			{/snippet}
		</Sidebar>
		{@render children()}
	</div>
{/if}
