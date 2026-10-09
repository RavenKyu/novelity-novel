<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { loadSession, session } from '$lib/auth.svelte';
	import Sidebar from '$lib/components/layout/Sidebar.svelte';

	let { children } = $props();

	// /login renders without the admin shell (argos "chromeless" pattern).
	const isChromeless = $derived(page.url.pathname === '/login');

	$effect(() => {
		if (!session.checked) {
			void loadSession();
		} else if (!session.user && !isChromeless) {
			void goto('/login', { replaceState: true });
		} else if (session.user && isChromeless) {
			void goto('/', { replaceState: true });
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Novelity Admin</title>
</svelte:head>

{#if isChromeless}
	{#if session.checked && !session.user}
		{@render children()}
	{/if}
{:else if session.user}
	<div class="flex min-h-screen bg-[var(--color-body)]">
		<Sidebar />
		{@render children()}
	</div>
{/if}
