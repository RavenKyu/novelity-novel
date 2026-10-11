<script lang="ts">
	import { goto } from '$app/navigation';
	import { ArrowRightToBracketOutline, ChevronSortOutline } from 'flowbite-svelte-icons';
	import { logout, session } from '$lib/auth.svelte';

	let open = $state(false);
	let root: HTMLDivElement | undefined = $state();

	const user = $derived(session.user);
	const initial = $derived((user?.name || user?.email || '?').charAt(0).toUpperCase());

	function onWindowClick(e: MouseEvent) {
		if (open && root && !root.contains(e.target as Node)) open = false;
	}

	async function handleLogout() {
		open = false;
		await logout();
		await goto('/login');
	}
</script>

<svelte:window onclick={onWindowClick} />

{#if user}
	<div class="user-menu-root" bind:this={root}>
		<button
			type="button"
			class="user-chip"
			class:open
			onclick={() => (open = !open)}
			aria-haspopup="true"
			aria-expanded={open}
		>
			{#if user.picture}
				<img src={user.picture} alt={user.name} class="avatar" referrerpolicy="no-referrer" />
			{:else}
				<span class="avatar avatar-letter">{initial}</span>
			{/if}
			<span class="info sidebar-label">
				<span class="name">{user.name || user.email}</span>
				<span class="email">{user.email}</span>
			</span>
			<span class="chev sidebar-label"><ChevronSortOutline class="h-3.5 w-3.5" /></span>
		</button>

		{#if open}
			<div class="menu" role="menu">
				<div class="menu-header">
					<div class="menu-name">{user.name || user.email}</div>
					<div class="menu-email">{user.email}</div>
				</div>
				<button type="button" class="menu-item danger" role="menuitem" onclick={handleLogout}>
					<ArrowRightToBracketOutline class="h-4 w-4" />로그아웃
				</button>
			</div>
		{/if}
	</div>
{/if}

<style>
	.user-menu-root {
		position: relative;
		padding: 10px 12px;
		border-top: 1px solid var(--color-sidebar-border);
	}
	.user-chip {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		padding: 4px 6px;
		border-radius: 8px;
		background: transparent;
		border: none;
		cursor: pointer;
		color: inherit;
		text-align: left;
	}
	.user-chip:hover,
	.user-chip.open {
		background: var(--color-sidebar-hover);
	}
	.avatar {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		object-fit: cover;
		flex-shrink: 0;
	}
	.avatar-letter {
		background: var(--color-accent);
		color: #fff;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		font-size: 12px;
		font-weight: 600;
	}
	.info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}
	.name {
		font-size: 12px;
		font-weight: 500;
		color: var(--color-sidebar-text-active);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.email {
		font-size: 10px;
		color: var(--color-sidebar-text);
		opacity: 0.7;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.chev {
		flex-shrink: 0;
		width: 24px;
		height: 24px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		color: var(--color-sidebar-text);
	}
	.menu {
		position: absolute;
		left: 12px;
		right: 12px;
		bottom: calc(100% - 4px);
		min-width: 200px;
		background: var(--color-bg-elevated);
		color: var(--color-text-primary);
		border: 1px solid var(--color-border);
		border-radius: 10px;
		padding: 6px;
		box-shadow: 0 12px 32px rgba(0, 0, 0, 0.25);
		z-index: 50;
	}
	.menu-header {
		padding: 10px 10px 8px;
		border-bottom: 1px solid var(--color-border);
		margin-bottom: 6px;
	}
	.menu-name {
		font-weight: 600;
		font-size: 13px;
	}
	.menu-email {
		font-size: 11px;
		color: var(--color-text-dimmed);
		margin-top: 1px;
	}
	.menu-item {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 10px;
		border-radius: 6px;
		font-size: 13px;
		background: transparent;
		border: none;
		cursor: pointer;
		width: 100%;
		text-align: left;
		color: var(--color-text-primary);
	}
	.menu-item:hover {
		background: var(--color-accent-subtle);
	}
	.menu-item.danger {
		color: var(--color-error);
	}
</style>
