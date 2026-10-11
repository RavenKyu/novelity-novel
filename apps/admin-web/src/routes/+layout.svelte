<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { loadSession, logout, session } from '$lib/auth.svelte';
	import { Sidebar, UserMenu } from '@novelity/ui';
	import { HomeOutline } from 'flowbite-svelte-icons';
	import { localReturnTo } from '@novelity/api-client';

	let { children } = $props();
	const isLogin = $derived(page.url.pathname === '/login');

	$effect(() => {
		if (!session.checked) {
			void loadSession();
		} else if (!session.user && !isLogin) {
			const query = new URLSearchParams();
			const target = page.url.pathname + page.url.search + page.url.hash;
			if (target !== '/') query.set('returnTo', target);
			if (session.forbidden) query.set('error', 'forbidden');
			void goto(query.size ? `/login?${query}` : '/login', { replaceState: true });
		} else if (session.user && isLogin) {
			void goto(localReturnTo(page.url.searchParams.get('returnTo')), { replaceState: true });
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<meta name="robots" content="noindex, nofollow" />
	<title>Novelity Admin</title>
</svelte:head>

{#if isLogin}
	{#if session.checked && !session.user}{@render children()}{/if}
{:else if session.user?.roles.includes('admin')}
	<div class="flex min-h-screen bg-[var(--color-body)]">
		<Sidebar
			title="Novelity Admin"
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
