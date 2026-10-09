<script lang="ts">
	import { page } from '$app/state';
	import { HomeOutline } from 'flowbite-svelte-icons';
	import SidebarItem from './SidebarItem.svelte';
	import UserMenu from './UserMenu.svelte';

	const navigation = [{ label: '홈', href: '/', icon: HomeOutline }];

	function isActive(href: string): boolean {
		const path = page.url.pathname;
		return href === '/' ? path === '/' : path === href || path.startsWith(href + '/');
	}
</script>

<aside
	class="fixed top-0 left-0 z-10 flex h-screen flex-col border-r border-[var(--color-sidebar-border)] bg-[var(--color-sidebar)]"
>
	<!-- Logo -->
	<div
		class="flex h-[var(--header-height)] shrink-0 items-center border-b border-[var(--color-sidebar-border)] px-4"
	>
		<div class="flex items-center gap-2.5">
			<div
				class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-[var(--color-accent)]"
			>
				<span class="text-xs font-bold text-white">N</span>
			</div>
			<span
				class="sidebar-label text-sm font-semibold tracking-tight text-[var(--color-sidebar-text-active)]"
				>Novelity Admin</span
			>
		</div>
	</div>

	<!-- Navigation -->
	<nav class="flex-1 overflow-y-auto px-2 py-2">
		{#each navigation as item (item.href)}
			<SidebarItem
				href={item.href}
				label={item.label}
				icon={item.icon}
				active={isActive(item.href)}
			/>
		{/each}
	</nav>

	<!-- User chip + dropdown -->
	<UserMenu />

	<!-- Bottom -->
	<div
		class="sidebar-label border-t border-[var(--color-sidebar-border)] px-4 py-2 text-[10px] text-[var(--color-sidebar-text)] opacity-50"
	>
		Novelity v{__APP_VERSION__}
	</div>
</aside>

<style>
	aside {
		width: var(--sidebar-width);
	}

	/* Narrow window: icon-width rail (see --sidebar-width in layout.css) that
	   expands over the content on hover. */
	@media (max-width: 1279px) {
		aside {
			transition: width 0.15s ease;
			overflow: hidden;
		}
		aside:hover {
			width: var(--sidebar-width-expanded);
			z-index: 40;
			box-shadow: 4px 0 16px rgba(0, 0, 0, 0.12);
		}
		aside:not(:hover) :global(.sidebar-label) {
			display: none;
		}
		aside:not(:hover) :global(.user-menu-root) {
			padding-left: 8px;
			padding-right: 8px;
		}
		/* Center icons in the collapsed 56px rail */
		aside:not(:hover) nav :global(a) {
			padding-left: 11px;
			padding-right: 11px;
		}
	}
</style>
