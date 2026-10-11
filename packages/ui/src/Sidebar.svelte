<script lang="ts">
	import type { Component, Snippet } from 'svelte';
	import SidebarItem from './SidebarItem.svelte';
	interface Props {
		title: string;
		version: string;
		pathname: string;
		navigation: { label: string; href: string; icon?: Component<{ class?: string }> }[];
		userMenu?: Snippet;
	}
	let { title, version, pathname, navigation, userMenu }: Props = $props();
	function isActive(href: string): boolean {
		return href === '/' ? pathname === '/' : pathname === href || pathname.startsWith(href + '/');
	}
</script>

<aside
	class="top-0 left-0 fixed z-10 flex h-screen flex-col border-r border-[var(--color-sidebar-border)] bg-[var(--color-sidebar)]"
>
	<!-- Logo -->
	<div
		class="px-4 flex h-[var(--header-height)] shrink-0 items-center border-b border-[var(--color-sidebar-border)]"
	>
		<div class="gap-2.5 flex items-center">
			<div
				class="h-7 w-7 rounded-lg flex shrink-0 items-center justify-center bg-[var(--color-accent)]"
			>
				<span class="text-xs font-bold text-white">N</span>
			</div>
			<span
				class="sidebar-label text-sm font-semibold tracking-tight text-[var(--color-sidebar-text-active)]"
				>{title}</span
			>
		</div>
	</div>

	<!-- Navigation -->
	<nav class="px-2 py-2 flex-1 overflow-y-auto">
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
	{#if userMenu}{@render userMenu()}{/if}

	<!-- Bottom -->
	<div
		class="sidebar-label px-4 py-2 border-t border-[var(--color-sidebar-border)] text-[10px] text-[var(--color-sidebar-text)] opacity-50"
	>
		Novelity v{version}
	</div>
</aside>

<style>
	aside {
		width: var(--sidebar-width);
	}

	/* Narrow window: icon-width rail (see --sidebar-width in tokens.css) that
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
