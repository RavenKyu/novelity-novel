<script lang="ts">
	import { onMount } from 'svelte';
	import { Card } from 'flowbite-svelte';
	import { Header, PageContent } from '@novelity/ui';
	import { loadSession } from '$lib/auth.svelte';

	let status = $state('확인 중');
	onMount(async () => {
		try {
			const response = await fetch('/api/admin/users');
			if (response.status === 401 || response.status === 403) await loadSession();
			status = response.ok ? '정상' : '연결 안 됨';
		} catch {
			status = '연결 안 됨';
		}
	});
</script>

<Header breadcrumbs={[{ label: '관리자 홈' }]} />
<PageContent>
	<div class="flex w-full max-w-[1280px] flex-col gap-5">
		<h1 class="text-xl font-semibold text-[var(--color-text-primary)]">관리자 홈</h1>
		<Card
			class="max-w-none rounded-lg border-[var(--color-surface-border)] bg-[var(--color-surface)] p-5 shadow-none"
		>
			<h2 class="text-sm font-semibold text-[var(--color-text-primary)]">관리자 연결: {status}</h2>
			<p class="mt-2 text-sm text-[var(--color-text-secondary)]">
				사용자, 작품, 신고 관리 기능은 준비 중입니다.
			</p>
		</Card>
	</div>
</PageContent>
