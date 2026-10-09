<script lang="ts">
	import { page } from '$app/state';
	import { Card } from 'flowbite-svelte';
	import { googleLoginUrl } from '$lib/auth.svelte';

	// Codes come from core-api's callback redirect (/login?error=…).
	const messages: Record<string, string> = {
		not_configured: 'Google 로그인이 설정되지 않았습니다. 관리자에게 문의하세요.',
		denied: 'Google 로그인이 취소되었습니다.',
		invalid_state: '로그인 요청이 만료되었거나 올바르지 않습니다. 다시 시도해 주세요.',
		exchange_failed: 'Google 인증에 실패했습니다. 다시 시도해 주세요.',
		invalid_token: 'Google 인증 정보를 확인할 수 없습니다. 다시 시도해 주세요.',
		unverified_email: '이메일이 확인된 Google 계정만 사용할 수 있습니다.',
		server_error: '서버 오류로 로그인하지 못했습니다. 잠시 후 다시 시도해 주세요.'
	};
	const errorCode = $derived(page.url.searchParams.get('error'));
	const error = $derived(errorCode ? (messages[errorCode] ?? messages.server_error) : null);
</script>

<svelte:head>
	<title>로그인 · Novelity Admin</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-[var(--color-body)] px-4">
	<div class="w-full max-w-sm">
		<Card
			class="max-w-none rounded-lg border-[var(--color-surface-border)] bg-[var(--color-surface)] p-5 shadow-none"
		>
			<div class="flex flex-col items-center text-center">
				<div
					class="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-[var(--color-accent)]"
				>
					<span class="text-lg font-bold text-white">N</span>
				</div>
				<h1 class="text-xl font-semibold tracking-tight text-[var(--color-text-primary)]">
					Novelity Admin
				</h1>
				<p class="mt-1 text-sm text-[var(--color-text-secondary)]">
					Google 계정으로 로그인하여 시작하세요.
				</p>
			</div>

			{#if error}
				<div
					class="mt-6 rounded-lg border border-[#ffc9c9] bg-[var(--color-error-bg)] px-3 py-2 text-xs text-[var(--color-error)]"
				>
					{error}
				</div>
			{/if}

			<a
				href={googleLoginUrl}
				data-sveltekit-reload
				class="mt-6 flex w-full items-center justify-center gap-2 rounded-lg border border-[var(--color-input-border)] bg-white py-2.5 text-sm font-semibold text-[var(--color-text-primary)] hover:bg-[var(--color-surface-hover)]"
			>
				<svg width="15" height="15" viewBox="0 0 18 18" aria-hidden="true">
					<path
						fill="#4285F4"
						d="M17.64 9.2c0-.64-.06-1.25-.16-1.84H9v3.48h4.84a4.14 4.14 0 0 1-1.8 2.72v2.26h2.91c1.7-1.57 2.69-3.88 2.69-6.62z"
					/>
					<path
						fill="#34A853"
						d="M9 18c2.43 0 4.47-.8 5.96-2.18l-2.91-2.26c-.8.54-1.84.86-3.05.86-2.34 0-4.33-1.58-5.04-3.71H.96v2.33A9 9 0 0 0 9 18z"
					/>
					<path
						fill="#FBBC05"
						d="M3.96 10.71A5.41 5.41 0 0 1 3.68 9c0-.59.1-1.17.28-1.71V4.96H.96A9 9 0 0 0 0 9c0 1.45.35 2.83.96 4.04l3-2.33z"
					/>
					<path
						fill="#EA4335"
						d="M9 3.58c1.32 0 2.5.45 3.44 1.35l2.58-2.59C13.46.9 11.43 0 9 0A9 9 0 0 0 .96 4.96l3 2.33C4.67 5.17 6.66 3.58 9 3.58z"
					/>
				</svg>
				<span>Google로 로그인</span>
			</a>
		</Card>
	</div>
</div>
