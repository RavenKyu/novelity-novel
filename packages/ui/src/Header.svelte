<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ChevronRightOutline } from 'flowbite-svelte-icons';

	interface Props {
		breadcrumbs?: { label: string; href?: string }[];
		actions?: Snippet;
	}

	let { breadcrumbs = [], actions }: Props = $props();
</script>

<header
	class="top-0 right-0 px-6 fixed z-20 flex h-[var(--header-height)] items-center justify-between border-b border-[var(--color-header-border)] bg-[var(--color-header)]"
	style="left: var(--sidebar-width)"
>
	<!-- Breadcrumbs -->
	<nav class="min-w-0 gap-1.5 text-sm flex items-center">
		{#each breadcrumbs as crumb, i (i)}
			{#if i > 0}
				<ChevronRightOutline class="h-3.5 w-3.5 shrink-0 text-[var(--color-sidebar-text)]" />
			{/if}
			{#if crumb.href && i < breadcrumbs.length - 1}
				<a
					href={crumb.href}
					class="truncate text-[var(--color-sidebar-text-active)] hover:underline">{crumb.label}</a
				>
			{:else}
				<span class="truncate text-[var(--color-sidebar-text)]">{crumb.label}</span>
			{/if}
		{/each}
	</nav>

	{#if actions}
		<div class="gap-2 flex shrink-0 items-center">
			{@render actions()}
		</div>
	{/if}
</header>
