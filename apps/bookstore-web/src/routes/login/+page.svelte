<script lang="ts">
	import type { PageProps } from './$types';
	let { data }: PageProps = $props();
</script>

<svelte:head
	><title>로그인 — Novelity 서점</title><meta name="robots" content="noindex" /></svelte:head
>
<section class="login-card">
	<h1>Novelity 로그인</h1>
	<p>서점과 작업실에서 같은 계정을 사용합니다.</p>
	{#if data.errorMessage}<p role="alert" class="notice">{data.errorMessage}</p>{/if}
	{#if data.user}
		<p>{data.user.name} 님으로 로그인되어 있습니다.</p>
		<a class="primary-link" href="/">서점으로</a>
	{:else}
		<a class="primary-link" href={data.loginUrl}>Google로 로그인</a>
	{/if}
	<p><a href="/">서점 둘러보기</a></p>
</section>
