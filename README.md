# file-diff

JetBrains IntelliJ IDEA의 diff 엔진을 Go로 포팅한 라이브러리입니다. 라인, 워드, 문자 수준의 다단계 비교를 지원합니다.

## 설치

```bash
go get github.com/byron1st/file-diff
```

## 사용법

### 라인 수준 비교

두 텍스트를 줄 단위로 비교합니다. 세 가지 알고리즘 중 선택할 수 있습니다.

```go
import (
    "strings"

    "github.com/byron1st/file-diff/go/comparison"
)

text1 := "hello\nworld\nfoo"
text2 := "hello\nchanged\nfoo"

lines1 := strings.Split(text1, "\n")
lines2 := strings.Split(text2, "\n")

// 알고리즘 선택
matcher := &comparison.HistogramMatcher{}

// 비교 수행
diff := matcher.Match(lines1, lines2, comparison.PolicyDefault)

// 변경된 범위 조회
for _, r := range diff.Changes() {
    // r.Start1, r.End1: 첫 번째 텍스트의 변경 라인 범위 [Start, End)
    // r.Start2, r.End2: 두 번째 텍스트의 변경 라인 범위 [Start, End)
    fmt.Printf("left[%d:%d] -> right[%d:%d]\n", r.Start1, r.End1, r.Start2, r.End2)
}

// 변경되지 않은 범위 조회
for _, r := range diff.Unchanged() {
    fmt.Printf("equal: left[%d:%d] = right[%d:%d]\n", r.Start1, r.End1, r.Start2, r.End2)
}
```

### 워드 수준 비교

변경된 라인 내에서 단어 단위로 세밀한 차이를 확인합니다.

```go
import "github.com/byron1st/file-diff/go/comparison"

text1 := "the quick brown fox"
text2 := "the slow brown dog"

fragments, err := comparison.CompareWords(text1, text2, comparison.PolicyDefault)
if err != nil {
    // handle error
}

for _, f := range fragments {
    // f.StartOffset1, f.EndOffset1: text1 내 바이트 오프셋
    // f.StartOffset2, f.EndOffset2: text2 내 바이트 오프셋
    fmt.Printf("changed: %q -> %q\n",
        text1[f.StartOffset1:f.EndOffset1],
        text2[f.StartOffset2:f.EndOffset2])
}
```

### 문자 수준 비교

바이트(코드 포인트) 단위의 비교입니다.

```go
import "github.com/byron1st/file-diff/go/comparison"

diff, err := comparison.CompareChars("abcdef", "abXdeZ")
if err != nil {
    // handle error
}

for _, r := range diff.Changes() {
    fmt.Printf("changed at byte offsets: [%d:%d) -> [%d:%d)\n",
        r.Start1, r.End1, r.Start2, r.End2)
}
```

### 다단계 비교 (라인 + 워드)

라인 비교 후, 변경된 라인에 대해 워드 비교를 적용한 `[]fragment.LineFragment` 를 바로 얻을 수 있습니다. GUI diff 뷰어 등에서 유용합니다.

```go
import (
    "strings"

    "github.com/byron1st/file-diff/go/comparison"
)

lines1 := strings.Split(text1, "\n")
lines2 := strings.Split(text2, "\n")

fragments, err := comparison.CompareLineFragments(
    lines1,
    lines2,
    &comparison.HistogramMatcher{},
    comparison.PolicyDefault,
)
if err != nil {
    // handle error
}
```

필요하면 여전히 `LineMatcher.Match()` 와 `CompareWords()` 를 직접 조합하는 저수준 패턴도 사용할 수 있습니다.

## 알고리즘

`LineMatcher` 인터페이스를 구현하는 세 가지 알고리즘을 제공합니다.

| 알고리즘 | 타입 | 특징 |
|---------|------|------|
| Myers | `MyersMatcher` | O(ND) 알고리즘. 2단계 비교(짧은 라인 제외 후 전체 비교)를 통한 최적화 적용. |
| Patience | `PatienceMatcher` | 양쪽에 고유한 라인을 앵커로 사용하여 LIS 기반 매칭. 비고유 영역은 Myers로 폴백. 구조적 변경(메서드 이동, 리팩터링)에 적합. |
| Histogram | `HistogramMatcher` | Patience를 확장하여 라인 빈도 기반 앵커 선택. 반복 구조가 많은 파일(JSON, YAML 등)에서 Patience보다 우수. |

```go
var matcher comparison.LineMatcher

matcher = &comparison.MyersMatcher{}      // Myers
matcher = &comparison.PatienceMatcher{}   // Patience
matcher = &comparison.HistogramMatcher{}  // Histogram (권장)
```

## 비교 정책 (ComparisonPolicy)

공백 처리 방식을 제어합니다.

| 정책 | 상수 | 설명 |
|-----|------|------|
| 기본 | `PolicyDefault` | 원문 그대로 비교 |
| 공백 트림 | `PolicyTrimWhitespaces` | 각 라인의 앞뒤 공백을 제거한 후 비교 |
| 공백 무시 | `PolicyIgnoreWhitespaces` | 모든 공백 차이를 무시 |

```go
diff := matcher.Match(lines1, lines2, comparison.PolicyIgnoreWhitespaces)
```

## 주요 타입

### `DiffIterable` / `FairDiffIterable`

비교 결과를 나타내는 인터페이스입니다.

```go
type DiffIterable interface {
    Length1() int            // 첫 번째 시퀀스 길이
    Length2() int            // 두 번째 시퀀스 길이
    Changes() []util.Range   // 변경된 범위
    Unchanged() []util.Range // 변경되지 않은 범위
}
```

`FairDiffIterable`는 `DiffIterable`를 확장하며, 변경되지 않은 범위에서 양쪽의 요소 수가 동일함을 보장합니다 (`End1-Start1 == End2-Start2`).

### `util.Range`

반개방 구간 `[Start, End)`으로 양쪽 시퀀스의 범위를 표현합니다.

```go
type Range struct {
    Start1, End1 int  // 첫 번째 시퀀스
    Start2, End2 int  // 두 번째 시퀀스
}
```

### `fragment.DiffFragment`

워드/문자 수준의 변경을 바이트 오프셋으로 표현합니다.

```go
type DiffFragment struct {
    StartOffset1, EndOffset1 int  // 첫 번째 텍스트 내 바이트 오프셋
    StartOffset2, EndOffset2 int  // 두 번째 텍스트 내 바이트 오프셋
}
```

### `fragment.LineFragment`

라인 수준의 변경을 표현하며, 내부에 워드/문자 수준 변경(`InnerFragments`)을 포함할 수 있습니다.

```go
type LineFragment struct {
    StartLine1, EndLine1 int       // 라인 범위
    StartLine2, EndLine2 int
    StartOffset1, EndOffset1 int   // 바이트 오프셋 범위
    StartOffset2, EndOffset2 int
    InnerFragments []DiffFragment  // 워드/문자 수준 세부 변경 (nil 가능)
}
```

## 라이선스

이 프로젝트는 JetBrains의 IntelliJ IDEA Community Edition에서 포팅되었습니다. 원본 코드는 Apache 2.0 라이선스를 따르며, 일부 구성 요소는 MIT 라이선스를 따릅니다. 자세한 내용은 [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md)를 참조하세요.
