<script lang="ts">
	import type { PageProps } from './$types';
	let { data }: PageProps = $props();
	const description = '작업실에서 완성한 연재와 책을 웹에서 읽는 Novelity 서점입니다.';
</script>

<svelte:head>
	<title>Novelity 서점 — 이야기를 만나는 곳</title>
	<meta name="description" content={description} />
	<meta property="og:title" content="Novelity 서점 — 이야기를 만나는 곳" />
	<meta property="og:description" content={description} />
	<meta property="og:type" content="website" />
	<meta property="og:locale" content="ko_KR" />
	<meta property="og:site_name" content="Novelity" />
	<meta property="og:url" content={data.canonical + '/'} />
	<link rel="canonical" href={data.canonical + '/'} />
</svelte:head>

<section class="intro">
	<p class="eyebrow">이야기를 만나는 곳</p>
	<h1>완성된 이야기의<br />다음 공간, 서점</h1>
	<p>
		작가의 작업실에서 시작한 이야기를<br />독자가 웹에서 바로 읽을 수 있는 곳을 준비하고 있습니다.
	</p>
</section>
<section aria-labelledby="publication-types" class="publications">
	<h2 id="publication-types">연재도, 한 권의 책도</h2>
	<div class="format-grid">
		<article>
			<p class="eyebrow">회차마다 이어지는 이야기</p>
			<h3>연재</h3>
			<p>다음 회차를 기다리며 읽는 이야기.</p>
		</article>
		<article>
			<p class="eyebrow">한 권으로 완성한 이야기</p>
			<h3>책</h3>
			<p>처음부터 끝까지 함께하는 이야기.</p>
		</article>
	</div>
	<p class="empty-state">아직 공개된 작품이 없습니다. 작품이 발행되면 이곳에서 만날 수 있습니다.</p>
</section>
