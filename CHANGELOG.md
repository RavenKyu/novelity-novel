# Changelog

릴리스별 변경 내역이다. `Prepare release` 워크플로가 릴리스 PR을 만들 때 새 버전을 맨 위에 추가한다. 내용은 병합된 PR 제목을 라벨별로 묶은 것이고, 릴리스 PR에서 직접 다듬어도 된다. 병합하면 해당 버전 절이 GitHub Release 본문이 된다.

## Unreleased — 세 공간 분리

- 작가 앱을 studio-web으로 옮기고 공통 UI·API 클라이언트와 npm workspaces를 도입했다.
- 공간별 API 선택, identity·역할, 서점·작가 공유 세션과 별도 2시간 관리자 세션의 기반을 추가했다.
- 로그인 복귀 주소·OAuth 출처·흐름 무결성을 검증하고 compose·CI·릴리스·개발 문서를 새 구조에 맞췄다.
- 서점·관리자 웹 골격과 관리자 내부망 배포는 후속 에픽에서 제공한다.
