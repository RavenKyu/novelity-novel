<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { createApiClient } from '@novelity/api-client';
	import type { LayoutProps } from './$types';
	import './layout.css';

	let { data, children }: LayoutProps = $props();
	let loggingOut = $state(false);
	let logoutError = $state(false);
	async function logout() {
		loggingOut = true;
		logoutError = false;
		try {
			await createApiClient().logout();
			await invalidateAll();
		} catch {
			logoutError = true;
		} finally {
			loggingOut = false;
		}
	}
</script>

<a class="skip-link" href="#content">본문으로 이동</a>
<header class="store-header">
	<a class="brand" href="/" aria-label="Novelity 서점 홈">Novelity <span>서점</span></a>
	<nav aria-label="서점 메뉴">
		<a href={data.studioUrl}>작업실</a>
		{#if data.sessionUnavailable}
			<span role="status">로그인 상태 확인 불가</span>
		{:else if data.user}
			<span data-testid="session-name">{data.user.name} 님</span>
			<button onclick={logout} disabled={loggingOut}
				>{loggingOut ? '로그아웃 중…' : '로그아웃'}</button
			>
		{:else}
			<a href="/login">로그인</a>
		{/if}
	</nav>
</header>
{#if logoutError}<p role="alert" class="notice">
		로그아웃하지 못했습니다. 다시 시도해 주세요.
	</p>{/if}
<main id="content">{@render children()}</main>
<footer>작업실에서 완성한 이야기, Novelity에서 독자에게.</footer>
