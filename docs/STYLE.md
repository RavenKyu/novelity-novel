# Style Guide (admin web)

> 마지막 갱신: 2026-10-09
> 출처: argos 프론트엔드(`argos/frontend/src/app.css`, `lib/components/layout/*`, `lib/components/ui/*`)의 스타일을 차용했다.
> 토큰 원본: `apps/web/src/routes/layout.css`의 `@theme`. 값을 바꿀 때는 그 파일과 이 문서를 함께 고친다.

## 원칙

- **어두운 크롬, 밝은 본문.** 사이드바와 헤더는 같은 짙은 남색(`#1e2028`)이고, 본문은 연회색(`#f4f5f7`), 카드는 흰색이다.
- **색은 토큰으로만 쓴다.** `bg-[var(--color-surface)]`, `text-[var(--color-text-secondary)]`처럼 쓰고, `bg-gray-50` 같은 Tailwind 팔레트 색을 직접 쓰지 않는다. 예외는 neutral 배지의 `bg-gray-100`, `bg-gray-400`으로 argos와 같다.
- **조밀하게 만든다.** 기본 글자는 14px, 사이드바 항목 높이 32px(`h-8`), 헤더 44px, 버튼 `h-7`/`h-8`이다. 그림자는 최소로 쓰고, 경계는 1px 보더로 표현한다.
- **다크 모드는 없다.** argos와 같이 라이트 단일 테마다. 크롬이 이미 어둡기 때문이다. 필요해지면 토큰을 `:root.dark`에서 재정의하는 방식으로 추가한다.
- **UI 문구는 한국어로 쓴다** (예: 메뉴 "홈", 상태 "정상"/"연결 안 됨").

## 토큰

| 그룹 | 토큰 | 값 |
|------|------|-----|
| 크롬 | `--color-sidebar`, `--color-header` | `#1e2028` |
| | `--color-sidebar-border`, `--color-header-border` | `rgba(255,255,255,.08)` |
| | `--color-sidebar-text` / `-text-active` | `rgba(230,230,245,.65)` / `#e6e6f5` |
| | `--color-sidebar-hover` / `-active` | `rgba(255,255,255,.06)` / `rgba(99,132,255,.14)` |
| 본문 | `--color-body` | `#f4f5f7` |
| 표면 | `--color-surface` / `-hover` / `-border` | `#ffffff` / `#fafbfc` / `#e2e5ea` |
| | `--color-bg-elevated`, `--color-border` | `#ffffff`, `#e2e5ea` |
| 텍스트 | `--color-text-primary` / `-secondary` / `-dimmed` | `#1a1c23` / `#5a6070` / `#8b90a0` |
| 액센트 | `--color-accent` / `-hover` / `-subtle` | `#3871e0` / `#2c5dbd` / `rgba(56,113,224,.08)` |
| 상태 | `--color-success` / `-bg` | `#1a7f37` / `#dafbe1` |
| | `--color-error` / `-bg` | `#cf222e` / `#ffebe9` |
| | `--color-warning` / `-bg` | `#9a6700` / `#fff8c5` |
| 입력 | `--color-input-border` / `-focus` | `#d0d4dc` / `#3871e0` |
| 크기 | `--header-height` | `44px` |
| | `--sidebar-width-expanded` / `-collapsed` | `260px` / `56px` |

- 폰트: `system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif`, 14px, line-height 1.5, antialiased.
- flowbite-svelte의 `primary-50…900` 스케일은 액센트(`#3871e0` = 500)를 중심으로 다시 정의했다. 그래서 flowbite 컴포넌트의 기본 primary 색이 argos 액센트와 맞는다.

## 레이아웃

```
┌────────────┬──────────────────────────────────────┐
│ ▣ Logo     │ 홈 › 하위        (breadcrumbs) [actions] │  ← Header 44px, #1e2028
├────────────┼──────────────────────────────────────┤
│ ⌂ 홈       │                                      │
│            │  PageContent  p-6, max-w 1280px      │  ← body #f4f5f7
│            │   ┌ Card ─────────┐                   │
│ v0.1.0     │   └───────────────┘                   │
└────────────┴──────────────────────────────────────┘
  Sidebar 260px (1280px 미만: 56px 아이콘 레일, hover 시 260px 오버레이)
```

| 컴포넌트 | 위치 | 규칙 |
|----------|------|------|
| Sidebar | `lib/components/layout/Sidebar.svelte` | 상단 로고(28px 액센트 사각형 + 이니셜), 내비게이션, 하단 버전 표시. 1280px 미만에서는 `.sidebar-label`을 숨긴다 |
| SidebarItem | `lib/components/layout/SidebarItem.svelte` | `h-8 pl-4 pr-3 rounded-sm text-sm`, 아이콘 18px. 활성 항목은 `sidebar-active` 배경에 `font-medium` |
| Header | `lib/components/layout/Header.svelte` | 페이지마다 렌더한다. 브레드크럼은 `ChevronRight`로 구분하고 오른쪽에 `actions` 스니펫을 둔다 |
| PageContent | `lib/components/layout/PageContent.svelte` | `--header-height`·`--sidebar-width`만큼 띄우고 `p-6` |

셸 없는 화면(현재 `/login`)은 `+layout.svelte`의 `isChromeless`에 경로를 추가한다. argos 로그인처럼 본문 중앙에 `max-w-sm` 카드를 두고, 로고는 48px 액센트 사각형, 오류는 `error-bg` 박스로 표시한다.
사이드바 하단 사용자 칩·메뉴(`UserMenu.svelte`)는 argos UserMenu의 치수와 색을 그대로 따른다.

새 페이지는 다음 형태로 만든다.

```svelte
<Header breadcrumbs={[{ label: '상위', href: '/x' }, { label: '현재' }]} />
<PageContent>…</PageContent>
```

## 콘텐츠 요소

- **카드:** `bg-[var(--color-surface)] border border-[var(--color-surface-border)] rounded-lg p-5`이고 그림자는 없다(`shadow-none`). 클릭할 수 있는 카드는 hover 시 `shadow-sm`을 준다. flowbite `Card`를 쓸 때도 이 클래스로 덮어쓴다.
- **배지:** `rounded-full px-2 py-0.5 text-xs font-medium`. 변형마다 `-bg` 배경 토큰과 본색 토큰을 짝지어 쓴다. 필요하면 앞에 1.5×1.5 점을 붙인다.

  | 변형 | 배경 | 글자·점 |
  |------|------|------|
  | success | `--color-success-bg` | `--color-success` |
  | error | `--color-error-bg` | `--color-error` |
  | warning | `--color-warning-bg` | `--color-warning` |
  | info | `--color-accent-subtle` | `--color-accent` |
  | neutral | `gray-100` | `--color-text-secondary` / `gray-400` |
- **버튼:** `rounded-md font-medium`, md는 `h-8 px-3 text-sm`, sm은 `h-7 px-2.5 text-xs`.
  - primary: 액센트 배경에 흰 글자
  - secondary: surface 배경에 보더
  - ghost: hover 시 `accent-subtle` 배경과 액센트 글자
  - danger: error 글자에 hover 시 `error-bg`
- **섹션 제목:** 18px, weight 400, line-height 28px, margin `32px 0 16px`.
- **아이콘:** `flowbite-svelte-icons`의 Outline 계열을 쓴다. argos의 2px stroke 아이콘과 톤이 같다. 사이드바는 18px, 브레드크럼 구분자는 14px.
- **전환:** 색 전환은 `duration-100`, 사이드바 폭 전환은 `0.15s ease`.

## 컴포넌트 라이브러리와의 관계

flowbite-svelte는 기능 컴포넌트(모달, 드롭다운, 테이블, 폼 등)로 쓰고, 겉모양은 위 토큰으로 맞춘다.
앱 셸(사이드바·헤더·본문)은 argos 구조를 그대로 따르기 위해 자체 컴포넌트로 만들었다.
