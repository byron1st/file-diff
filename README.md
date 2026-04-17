# file-diff

JetBrains IntelliJ IDEA의 다단계 diff 엔진을 포팅한 라이브러리입니다. 라인 / 워드 / 문자 세 계층의 비교를 제공하며, 구조적 변경과 인라인 토큰 변경을 모두 가독성 높게 표현하는 것을 목표로 합니다.

라인 매칭 단계는 Myers(O(ND)), Patience, Histogram 세 가지 알고리즘을 플러거블한 인터페이스로 제공합니다.

## 언어별 가이드

두 구현은 동일한 API 형태와 동작(UTF-8 바이트 오프셋 규약 포함)을 공유합니다.

- [Go 가이드](docs/go-guide.md) — Go 1.26+, 표준 라이브러리만 사용.
- [Swift 가이드](docs/swift-guide.md) — Swift Package Manager, macOS 13+, Swift 6.0+ (타겟명 `FileDiff`).

## 라이선스

이 프로젝트는 JetBrains의 IntelliJ IDEA Community Edition에서 포팅되었습니다. 원본 코드는 Apache 2.0 라이선스를 따르며, 일부 구성 요소는 MIT 라이선스를 따릅니다. 자세한 내용은 [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md)와 [LICENSE](LICENSE)를 참조하세요.
