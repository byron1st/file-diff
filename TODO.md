# TODO

## SwiftLint 알고리즘 함수 disable 리팩토링 검토

Step 4 (Comparison core 포팅) 에서 Go 원본과 1:1 시그니처 유지를 위해 다음 함수들에 대해 SwiftLint 규칙을 블록 단위로 disable 했다. Step 5 이후 Swift 관용에 맞춰 리팩토링 여부를 재검토한다.

### 비활성화된 규칙

#### `function_body_length`
함수 본문(주석/공백 제외) 길이 제한.

- `.swiftlint.yml`: `warning: 50`, `error: 80`
- 의도: 한 함수의 책임을 제한하고 가독성/재사용성을 확보.

**Disable 대상**:
- `processLastRanges` (`swift/Sources/FileDiff/Internal/ChunkOptimizer.swift`, ~68줄)
- `correctChangesSecondStep` (`swift/Sources/FileDiff/ByLine.swift`, ~51줄)

이들은 상태 변수(`sample`, `last1/last2`, `builder.index1/index2`) 를 공유하는 단일 루프라 쪼개려면 래퍼 클래스 또는 인자 추가가 필요하다.

#### `function_parameter_count`
함수 파라미터 개수 제한.

- 기본값: `warning: 5`, `error: 8`
- 의도: 파라미터 과다는 구조체로 묶을 신호.

**Disable 대상**:
- `expandRangeLines`, `expandForwardLines`, `expandBackwardLines` — 각 6개 (`lines1, lines2, start1, start2, end1, end2`)
- `getLineShift` — 7개 (`lines1, lines2, touchSideIsLeft, eqFwd, eqBwd, r1, r2`)

`start*/end*` 네 인자를 `DiffRange` 로 묶으면 5개 이하로 줄일 수 있으나 Go 원본 시그니처와 괴리된다.

### 리팩토링 옵션
- `DiffRange` 로 `start1/start2/end1/end2` 를 묶어 `expand*Lines` 를 5개 이하 파라미터로 재설계.
- `correctChangesSecondStep` 의 루프 상태를 `SecondStepState` 구조체로 추출.
- `processLastRanges` 의 분기(forward merge / backward merge / shift)를 private helper 로 분리.

### 판단 기준
- Go 원본 알고리즘이 변경될 때 Swift 쪽 대응 비용.
- Step 6/7 (Patience/Histogram) 에서 동일 유틸(`expand*Lines`) 의 재사용 패턴.
- 추후 Swift-native 개선(`DiffRange` 메서드화 등) 도입 시 자연스러운 이행 경로인지.
