# Project Context: Enhanced Diff Engine

## 프로젝트 개요

JetBrains IDE의 다단계 diff 엔진(Apache 2.0)을 Go/Swift로 포팅하고, 라인 매칭 단계에 Histogram Diff 알고리즘을 결합하여 **구조적 변경과 세부 변경 모두에서 가독성이 높은 diff 엔진**을 만드는 프로젝트.

> **알고리즘 선택 근거**: Patience Diff와 Histogram Diff를 코드 리뷰 가독성 관점에서 비교 검토한 결과, Histogram이 Patience의 상위 호환이면서 성능도 더 빠르므로 Histogram을 최종 목표로 선정. 단, 구현은 Patience → Histogram 순서로 단계적으로 진행하며, 라인 매칭 알고리즘을 인터페이스로 추상화하여 교체 가능하게 설계한다.

---

## 배경 지식: Diff 알고리즘 종류

### 주요 알고리즘

| 알고리즘 | 핵심 아이디어 | 복잡도 | 사용처 |
|---|---|---|---|
| LCS (Longest Common Subsequence) | 두 시퀀스 간 최장 공통 부분 수열을 DP로 탐색 | O(MN) 시간/공간 | 고전적 diff의 근간 |
| Myers Diff (1986) | Edit graph 위에서 최단 edit script(SES)를 greedy 탐색 | O(ND), D=edit distance | Git 기본, JetBrains 기반 |
| Patience Diff | 양쪽 파일에서 한 번씩만 등장하는 고유 라인을 앵커로 매칭 후 재귀 | Myers + 전처리 | `git diff --patience` |
| Histogram Diff | Patience 확장, 라인 출현 빈도(histogram) 활용 | Patience보다 빠름 | `git diff --histogram`, JGit |
| Hunt-Szymanski | 공통 매칭 포인트를 정렬 후 LIS로 LCS 탐색 | 매칭 적을 때 효율적 | 초기 Unix diff |
| Hirschberg | LCS/edit distance를 분할 정복으로 공간 최적화 | O(MN) 시간, O(min(M,N)) 공간 | 대용량 파일 비교 |
| XDiff (LibXDiff) | Myers 기반 + 휴리스틱 최적화 | 실용적 | Git 내부 라이브러리 |

### Myers Diff의 한계 (이 프로젝트가 해결하려는 핵심 문제)

Myers는 수학적으로 최소 편집 거리를 찾지만, **의미적으로 엉뚱한 매칭**을 만들 수 있다.

```java
// Before
public void connect() {
    open();
}

public void disconnect() {
    close();
}

// After - connect 메서드 삭제
public void disconnect() {
    close();
}
```

Myers는 `connect()`의 여는 브레이스와 `disconnect()`의 닫는 브레이스를 매칭해서, 마치 `disconnect` 메서드의 본문이 바뀐 것처럼 보여줄 수 있다. 실제로는 `connect` 메서드 전체가 삭제된 것인데.

---

## 두 가지 접근법 비교

### Patience Diff: "구조적 앵커 포인트 확보" (입력 단계에서 해결)

**동작 방식:**
1. 양쪽 파일에서 **한 번씩만 등장하는 고유 라인**(함수 시그니처 등)을 찾아 앵커로 삼음
2. `{`, `}`, 빈 줄 같은 반복 라인은 무시
3. 앵커 사이 구간을 재귀적으로 처리
4. 고유 라인이 없는 구간에서는 Myers로 폴백

**장점:**
- 함수/메서드 단위의 추가·삭제·이동이 직관적으로 보임
- 코드 블록 경계가 엉뚱한 매칭에 사용되지 않음
- 대규모 리팩토링에서 변경 의도가 명확히 드러남

**한계:**
- 비교 단위가 여전히 **라인** — 라인 내부에서 어떤 토큰이 바뀌었는지는 보여주지 않음
- 고유 라인이 없는 구간(반복적 설정 파일, JSON 등)에서는 Myers로 폴백
- 빈 줄만 있는 구간이나 보일러플레이트가 많은 코드에서 앵커 부족

### JetBrains Diff: "매칭 후 심층 분석" (출력 단계에서 해결)

**동작 방식:**
1. **Line-level diff** — 라인 단위로 변경된 블록 식별 (Myers 기반)
2. **Word-level diff** — 변경된 라인 내에서 단어(word boundary) 단위로 세부 비교
3. **Character-level diff** — 필요시 문자 단위까지 정밀 비교

**Comparison Policy 지원:**
- `DEFAULT`: 공백 차이 포함
- `TRIM_WHITESPACES`: 앞뒤 공백 무시
- `IGNORE_WHITESPACES`: 모든 공백 차이 무시

**장점:**
- 라인 내부 변경의 정확한 위치를 시각적으로 즉시 파악 가능
- 긴 라인에서 한두 글자만 바뀐 경우(변수명/타입 변경 등) 극적으로 유용
- 공백 정책으로 포매팅 변경과 실질적 변경 분리 가능

**한계:**
- 라인 매칭 자체는 Myers 기반이므로 `{`/`}` 오매칭 문제가 발생 가능
- 함수 단위 이동이나 대규모 구조 변경에서 Patience보다 덜 직관적

### 핵심 차이 요약

| 관점 | Patience Diff | JetBrains Diff |
|---|---|---|
| 문제 해결 레이어 | 입력(매칭 전략 개선) | 출력(매칭 후 심층 분석) |
| 비교 깊이 | 라인 단위만 | 라인 → 단어 → 문자 |
| 구조적 변경 가독성 | 우수 (앵커 기반) | 보통 (Myers 한계 계승) |
| 라인 내부 변경 가독성 | 없음 | 우수 (word/char 하이라이팅) |
| 최적 시나리오 | 메서드 추가/삭제/이동 | 변수명·타입·인자 변경 |

---

## 프로젝트 목표: 두 접근법의 결합

이 두 가지는 **상호 보완적**이다. 결합 전략:

1. **라인 매칭 단계**: Myers 대신 **Histogram Diff**를 사용하여 구조적 앵커를 확보 (Patience → Histogram 단계적 구현)
2. **라인 내부 비교 단계**: JetBrains 방식의 **word-level → character-level 다단계 비교**를 적용
3. **Comparison Policy**: JetBrains의 공백 처리 정책(DEFAULT, TRIM, IGNORE) 유지

이렇게 하면 구조적 변경(메서드 추가/삭제/이동)과 세부 변경(변수명/타입 변경) 모두에서 가독성이 높은 diff를 생성할 수 있다.

---

## Patience vs Histogram 비교 분석 (알고리즘 선택 근거)

### 관계

Histogram은 Patience의 **향상 버전**이다. 완전히 다른 알고리즘이 아니라 Patience의 아이디어를 계승하면서 약점을 보완한 관계.

### 동작 방식 차이

- **Patience**: 양쪽 파일에서 **정확히 1번만 등장하는 라인**만 앵커로 사용
- **Histogram**: 모든 라인의 **출현 빈도 히스토그램**을 만들고, 빈도가 가장 낮은 라인부터 우선 매칭. 유니크 라인이 없어도 "가장 드문 라인"을 앵커로 사용 가능

### 코드 리뷰 가독성 비교

| 시나리오 | Patience | Histogram | 비고 |
|---|---|---|---|
| 함수 추가/삭제 | ✅ 우수 | ✅ 우수 | 함수 시그니처가 유니크하므로 동등 |
| 반복 패턴 코드 (테스트 등) | ⚠️ 앵커 부족 시 Myers 폴백 | ✅ 우수 | 빈도 기반이라 반복에도 안정적 |
| 설정 파일 (YAML/JSON) | ❌ 유니크 라인 거의 없음 | ✅ 우수 | Histogram의 가장 큰 강점 |
| 대규모 리팩토링 | ✅ 우수 | ✅ 약간 더 우수 | 리팩토링 중 유니크 라인 감소 시 차이 |

### 구체적 예시: 반복 패턴에서의 차이

```go
// 테스트 코드에서 중간에 새 테스트 추가 시
func TestAdd(t *testing.T) {
    result := Add(1, 2)        // 이 패턴이 모든 테스트에 반복
    assert.Equal(t, 3, result) // 이 패턴도 반복
}

func TestMultiply(t *testing.T) { // ← 새로 추가
    result := Multiply(4, 3)
    assert.Equal(t, 12, result)
}

func TestSubtract(t *testing.T) {
    result := Subtract(5, 3)
    assert.Equal(t, 2, result)
}
```

- **Patience**: `assert.Equal` 등이 유니크하지 않아 앵커 부족 → Myers 폴백 가능
- **Histogram**: 각 `TestXxx` 시그니처의 빈도가 낮으므로 자동으로 우선 매칭 → 안정적

### 성능 비교

Histogram이 일반적으로 더 빠르다. Patience는 유니크 라인 탐색 → patience sorting(LIS) → 재귀 과정을 거치지만, Histogram은 빈도 카운팅 기반으로 한 번의 패스에서 매칭 후보를 효율적으로 좁힌다.

### 학술 근거

논문 "How Different Are Different diff Algorithms in Git?" (Empirical Software Engineering, 2019)에서 Histogram 알고리즘을 Git 저장소 마이닝 시 강력히 권장. 코드 변경을 더 정확하게 표현하기 때문.

### 결론

**Histogram을 최종 목표로 설정.** 단, Patience가 구현이 단순하므로 먼저 구현한 후 Histogram으로 확장하는 단계적 전략을 채택.

---

## 소스 코드 참조

### JetBrains Diff 엔진 (포팅 대상)

- **저장소**: https://github.com/JetBrains/intellij-community
- **핵심 경로**: `platform/util/diff/src/com/intellij/diff/comparison/`
- **언어**: Kotlin/Java
- **라이선스**: Apache License 2.0

### 기존 포팅 선례

- **jbdiff (PHP 포팅)**: https://github.com/123inkt/jbdiff
  - JetBrains diff를 PHP로 포팅한 라이브러리
  - Apache 2.0 라이선스 유지
  - 참고할 수 있는 구조적 모범 사례

### Patience / Histogram Diff 참조 구현체

| 구현체 | 언어 | 라이선스 | 참조 가능 여부 |
|---|---|---|---|
| `peter-evans/patience` | **Go** | MIT | ✅ 적극 권장 |
| `ruby_patience_diff` | Ruby | MIT | ✅ 참조 가능 |
| Bazaar 원본 (Bram Cohen) | Python | **GPL v2+** | ⚠️ 참조 시 GPL 감염 |
| Git xdiff | C | **GPL v2** | ⚠️ 참조 시 GPL 감염 |
| `janestreet/patdiff` | OCaml | MIT | ✅ 참조 가능 |

---

## 라이선스 전략

### 원칙

- JetBrains 코드 포팅: **Apache 2.0 의무사항 준수**
- Patience / Histogram Diff: **MIT 구현체만 참조** (GPL 구현체 참조 금지)
- Apache 2.0 + MIT는 완전히 호환되므로 결합에 문제 없음

### Apache 2.0 의무사항 (반드시 준수)

1. **저작권 고지 유지**: `Copyright 2000-2024 JetBrains s.r.o. and contributors`
2. **NOTICE 파일 전달**: 원본 NOTICE 파일 내용 포함
3. **변경 사항 명시**: 원본에서 수정/포팅한 사실을 표시
4. **라이선스 텍스트 포함**: Apache 2.0 전문을 함께 배포

### MIT 의무사항

1. 저작권 고지 유지 (참조한 구현체의 copyright)
2. MIT 라이선스 텍스트 포함

### 배포 시 라이선스 구성

```
LICENSE                  # 이 프로젝트 자체의 라이선스 (Apache 2.0 또는 MIT)
NOTICE                   # JetBrains 원본 저작권 고지 + 변경 사항 명시
THIRD_PARTY_LICENSES.md  # 참조한 모든 구현체의 라이선스 고지
```

---

## 구현 계획

### 아키텍처 원칙: 라인 매칭 알고리즘 교체 가능 설계

```go
// 라인 매칭 알고리즘을 인터페이스로 추상화
type LineMatcher interface {
    Match(left, right []string) []LineMatch
}

type MyersMatcher struct{}      // 폴백용 / JetBrains 원본 호환
type PatienceMatcher struct{}   // Phase 2 - 단순 구현, 유니크 라인 기반
type HistogramMatcher struct{}  // Phase 3 - 최종 목표, 빈도 기반
```

### Phase 1: JetBrains Diff 핵심 로직 포팅

- `ComparisonManagerImpl` — 비교 진입점
- Line-level 비교 로직 (Myers 기반, `LineMatcher` 인터페이스로 추상화)
- Word-level 비교 로직
- Character-level 비교 로직
- ComparisonPolicy (DEFAULT, TRIM_WHITESPACES, IGNORE_WHITESPACES)

### Phase 2: Patience Diff 통합

- `PatienceMatcher` 구현 (MIT 구현체 참조)
- 유니크 라인 탐색 → patience sorting(LIS) → 재귀 분할
- `LineMatcher` 인터페이스를 통해 기본 매처를 Patience로 교체
- 고유 라인이 없는 구간에서의 Myers 폴백 유지

### Phase 3: Histogram Diff 확장

- `HistogramMatcher` 구현 (Patience 위에 빈도 카운팅 레이어 추가)
- 유니크 라인 제한 제거 → 최소 빈도 라인 기반 매칭으로 확장
- 반복 패턴 코드, 설정 파일(YAML/JSON) 등에서 개선 효과 검증
- 성능 벤치마크: Myers vs Patience vs Histogram

### Phase 4: 출력 포맷

- Unified diff 출력
- Side-by-side diff 출력
- Word/character 수준의 변경 정보를 포함하는 구조화된 출력 (JSON 등)

### 대상 언어

- **Go**: 1차 구현 목표
- **Swift**: 2차 구현 목표 (Go 구현 안정화 후)

---

## 참고 자료

- Myers 알고리즘 원논문: Eugene W. Myers, "An O(ND) Difference Algorithm and Its Variations" (1986)
- Bram Cohen의 Patience Diff 설명: https://bramcohen.livejournal.com/73318.html
- Patience Diff 상세 해설: https://blog.jcoglan.com/2017/09/19/the-patience-diff-algorithm/
- Histogram vs Myers 비교 (Adam Johnson): https://adamj.eu/tech/2024/01/18/git-improve-diff-histogram/
- 학술 논문 — diff 알고리즘 비교: https://link.springer.com/article/10.1007/s10664-019-09772-z
- JetBrains diff 소스: https://github.com/JetBrains/intellij-community/tree/master/platform/util/diff/src/com/intellij/diff
- PHP 포팅 선례 (jbdiff): https://github.com/123inkt/jbdiff
- Go Patience Diff (MIT): https://github.com/peter-evans/patience
