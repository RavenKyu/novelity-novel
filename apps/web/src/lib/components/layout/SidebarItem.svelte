<script lang="ts">
	import type { Component } from 'svelte';

	interface Props {
		href: string;
		label: string;
		icon?: Component<{ class?: string }>;
		active?: boolean;
	}

	let { href, label, icon: Icon, active = false }: Props = $props();
</script>

<a
	{href}
	class="flex h-8 items-center gap-2.5 rounded-sm pr-3 pl-4 text-sm transition-colors duration-100"
	class:bg-[var(--color-sidebar-active)]={active}
	class:font-medium={active}
	class:text-[var(--color-sidebar-text-active)]={active}
	class:text-[var(--color-sidebar-text)]={!active}
	class:hover:bg-[var(--color-sidebar-hover)]={!active}
>
	{#if Icon}
		<Icon class="h-[18px] w-[18px] shrink-0" />
	{/if}
	<span class="sidebar-label truncate">{label}</span>
</a>
