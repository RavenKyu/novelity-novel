<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();
</script>

<main
	class="flex min-h-screen min-w-0 flex-1 flex-col"
	style="padding-top: var(--header-height); margin-left: var(--sidebar-width)"
>
	<div class="flex min-h-0 flex-1 flex-col p-6">
		{@render children()}
	</div>
</main>
