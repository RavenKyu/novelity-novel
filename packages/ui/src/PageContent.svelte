<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();
</script>

<main
	class="min-w-0 flex min-h-screen flex-1 flex-col"
	style="padding-top: var(--header-height); margin-left: var(--sidebar-width)"
>
	<div class="min-h-0 p-6 flex flex-1 flex-col">
		{@render children()}
	</div>
</main>
