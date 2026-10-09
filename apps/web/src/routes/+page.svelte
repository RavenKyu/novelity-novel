<script lang="ts">
	import { onMount } from 'svelte';
	import { Badge, Card } from 'flowbite-svelte';
	import Header from '$lib/components/layout/Header.svelte';
	import PageContent from '$lib/components/layout/PageContent.svelte';

	let api = $state<'checking' | 'ok' | 'down'>('checking');

	onMount(async () => {
		try {
			const res = await fetch('/api/healthz');
			api = res.ok ? 'ok' : 'down';
		} catch {
			api = 'down';
		}
	});

	const status = {
		checking: {
			label: '확인 중',
			cls: 'bg-gray-100 text-[var(--color-text-secondary)]',
			dot: 'bg-gray-400'
		},
		ok: {
			label: '정상',
			cls: 'bg-[var(--color-success-bg)] text-[var(--color-success)]',
			dot: 'bg-[var(--color-success)]'
		},
		down: {
			label: '연결 안 됨',
			cls: 'bg-[var(--color-error-bg)] text-[var(--color-error)]',
			dot: 'bg-[var(--color-error)]'
		}
	};
	const s = $derived(status[api]);
</script>

<Header breadcrumbs={[{ label: '홈' }]} />

<PageContent>
	<div class="flex w-full max-w-[1280px] flex-col gap-5">
		<Card
			class="max-w-sm rounded-lg border-[var(--color-surface-border)] bg-[var(--color-surface)] p-5 shadow-none"
		>
			<p class="mb-2 text-xs font-medium text-[var(--color-text-dimmed)]">core-api</p>
			<Badge rounded class="w-fit gap-1.5 px-2 py-0.5 text-xs font-medium {s.cls}">
				<span class="h-1.5 w-1.5 rounded-full {s.dot}"></span>{s.label}
			</Badge>
		</Card>
	</div>
</PageContent>
