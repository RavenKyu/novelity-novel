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
	class="fixed top-0 right-0 z-20 flex h-[var(--header-height)] items-center justify-between border-b border-[var(--color-header-border)] bg-[var(--color-header)] px-6"
	style="left: var(--sidebar-width)"
>
	<!-- Breadcrumbs -->
	<nav class="flex min-w-0 items-center gap-1.5 text-sm">
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
		<div class="flex shrink-0 items-center gap-2">
			{@render actions()}
		</div>
	{/if}
</header>
