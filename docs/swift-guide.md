# Swift 사용 가이드

SwiftUI macOS 앱에서 라이브러리로 임포트해 사용할 수 있는 Swift 포팅입니다. Go 버전과 동일한 3계층(Line / Word / Char) 비교 API를 제공합니다.

- SPM 타겟명: `FileDiff`
- 최소 지원: macOS 13+, Swift 6.0+

## 설치

`Package.swift` 의 의존성에 추가합니다.

```swift
// Package.swift
dependencies: [
    .package(url: "https://github.com/byron1st/file-diff.git", branch: "main")
],
targets: [
    .target(name: "YourApp", dependencies: [
        .product(name: "FileDiff", package: "file-diff")
    ])
]
```

Xcode 프로젝트에서는 **File → Add Package Dependencies...** 메뉴로 저장소 URL을 추가한 뒤 `FileDiff` 라이브러리를 타겟에 연결하면 됩니다.

## 사용법

### 라인 수준 비교

```swift
import FileDiff

let lines1 = ["hello", "world", "foo"]
let lines2 = ["hello", "changed", "foo"]

// Myers / Patience / Histogram 매처 선택 가능
let result = HistogramMatcher().match(lines1, lines2, policy: .default)

for range in result.changes {
    print("left[\(range.left.lowerBound)..<\(range.left.upperBound)] -> " +
          "right[\(range.right.lowerBound)..<\(range.right.upperBound)]")
}

for range in result.unchanged {
    print("equal: left[\(range.left)] = right[\(range.right)]")
}
```

### 워드 수준 비교

바이트 오프셋 기준으로 변경된 단어 구간을 반환합니다.

```swift
import FileDiff

let fragments = compareWords(
    "the quick brown fox",
    "the slow brown dog",
    policy: .default
)

for f in fragments {
    // f.left, f.right: UTF-8 바이트 오프셋 범위 (Range<Int>)
    print("changed: \(f.left) -> \(f.right)")
}
```

### 문자 수준 비교

```swift
import FileDiff

let diff = compareChars("abcdef", "abXdeZ")

for range in diff.changes {
    print("changed at bytes: \(range.left) -> \(range.right)")
}
```

### 다단계 비교 (라인 + 인라인 워드)

라인 비교 후 변경 라인에 워드 비교를 적용한 결과를 한 번에 얻습니다.

```swift
import FileDiff

let lineFragments = compareLineFragments(
    lines1, lines2,
    matcher: HistogramMatcher(),
    policy: .default
)

for lf in lineFragments {
    print("lines \(lf.leftLines) -> \(lf.rightLines)")
    for inner in lf.inner {
        // inner 오프셋은 lf.leftOffsets / lf.rightOffsets 기준 상대값
        print("  inline \(inner.left) -> \(inner.right)")
    }
}
```

## 알고리즘

Go 버전과 동일하게 세 가지 라인 매처를 제공합니다.

| 알고리즘 | 타입 | 특징 |
|---------|------|------|
| Myers | `MyersMatcher` | O(ND) 알고리즘. 짧은 라인 제외 후 전체 비교하는 2단계 최적화 적용. |
| Patience | `PatienceMatcher` | 고유 라인을 앵커로 사용한 LIS 기반 매칭. 비고유 영역은 Myers로 폴백. |
| Histogram | `HistogramMatcher` | 라인 빈도 기반 앵커 선택. 반복 구조가 많은 파일에 유리 (권장). |

## 비교 정책

`ComparisonPolicy` 열거형을 사용합니다.

| 케이스 | 설명 |
|-------|------|
| `.default` | 원문 그대로 비교 |
| `.trimWhitespaces` | 각 라인의 앞뒤 공백을 제거한 후 비교 |
| `.ignoreWhitespaces` | 모든 공백 차이를 무시 |

```swift
let result = HistogramMatcher().match(lines1, lines2, policy: .ignoreWhitespaces)
```

## 오프셋 규약

모든 오프셋은 **UTF-8 바이트 기준** 이며 Go 버전과 동일합니다. 문자열을 오프셋으로 슬라이싱할 때는 `String.utf8` 뷰를 사용하거나 `Data` 를 경유해 주세요.
