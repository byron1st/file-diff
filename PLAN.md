# Project Implementation Plan: Enhanced Diff 엔진 (Go)

이 문서는 `CONTEXT.md` 및 JetBrains 참조 소스 코드(Java/Kotlin)를 바탕으로, JetBrains IDE의 다단계 Diff 엔진 성능과 Histogram 라인 매칭 방식을 결합한 새로운 Diff 엔진을 개발하기 위한 단계별 구현 계획입니다. 각 단계는 `계획 -> 구현 -> 검증(테스트)` 사이클이 가능한 독립적 단위로 구성됩니다. 최초 구현 대상 언어는 **Go**입니다.

---

## 🚀 Phase 1: 기본 도메인 설계 (핵심 인터페이스 및 정책)

**계획 (Plan)**
* 프로젝트의 모듈화된 아키텍처를 설계하고 핵심 비교 인터페이스를 도출합니다.
* JetBrains 원본의 여러 `ComparisonPolicy` 및 `LineMatcher` 등의 추상을 정의하여 교체 가능한 구조를 만듭니다.

**구현 (Implement)**
1. Go 프로젝트 초기화 (`go mod init`) 및 패키지 구조(`diff`, `comparison`, `util`) 설정
2. `ComparisonPolicy` 정의 (`DEFAULT`, `TRIM_WHITESPACES`, `IGNORE_WHITESPACES`)
3. `LineMatcher` 인터페이스 정의 (입력: `left`, `right` 라인 배열 / 출력: 매칭 블록 배열)
4. 저작권 고지를 포함하는 기본 LICENSE/NOTICE 구조 추가 (Apache 2.0 및 MIT)

**검증 (Test)**
* 정의된 Go 인터페이스 및 Enum 구문이 정상 컴파일되는지 확인
* 기본 타입들과 인터페이스가 상호작용하는지 아무 동작하지 않는 Dummy 구현체로 단위 테스트 작성

---

## 🚀 Phase 2: Myers 기반 Line-level 매처 (JetBrains 로직 포팅)

**계획 (Plan)**
* JetBrains의 `ByLineRt.kt` 핵심 원리와 `iterables`의 Chunk 최적화 로직 등 Line-level 비교 엔진을 Go로 가져옵니다.

**구현 (Implement)**
1. `ByLineRt.kt` 및 Myers 기반 최장 공통 부분 수열(LCS) 탐색 로직(`iterables` 하위 구조 등) 이식
2. `MyersMatcher` 구현체를 `LineMatcher` 인터페이스에 연결
3. 공백 처리 등 `ComparisonPolicy`에 따른 라인 데이터 트리밍 처리(`TrimUtil.kt` 이식 포함) 

**검증 (Test)**
* 라인 추가/수정/삭제 등 기본 `git diff`와 동일한 방식으로 블록이 식별되는지 단위 테스트 수행
* 파일 내 공백이 다른 라인이 `IGNORE_WHITESPACES` 정책에서 같다고 처리되는지 테스트

---

## 🚀 Phase 3: 세부 비교 로직 (Word & Char-level) 이식

**계획 (Plan)**
* 라인 단위로 잘려진 변경 블록 내에서, 단어(Word) 및 문자(Char) 단위의 미세 수정 위치를 정확하게 하이라이팅하는 JetBrains의 장점을 구현합니다.

**구현 (Implement)**
1. 텍스트 단어 단위 바운더리를 찾는 `CharacterUtils.kt` 포팅
2. `ByWordRt.kt` 포팅: 한 라인 안에서 수정된 단어 매칭 알고리즘 이식
3. `ByCharRt.kt` 포팅: 단어보다 더 미세한 레벨의 탐색 알고리즘 이식
4. (선택) `ChangeCorrector`와 같은 의미적 블록 병합 유틸리티 도입

**검증 (Test)**
* 라인 단위 테스트를 통과한 결과 블록 내에서, 짧은 변수명 하나의 스펠링 변경 등(Word-level diff)이 올바르게 추출되는지 확인
* 괄호 여닫기, 띄어쓰기 등 미세한 차이에 대한 Char-level 테스트 케이스 작성 및 통과 확인

---

## 🚀 Phase 4: Patience Diff 엔진 결합

**계획 (Plan)**
* 단순한 라인 변경이 아니라 **함수의 이동이나 대규모 리팩토링**에서 우수한 가독성을 확보하기 위해, 고양된 기준선(앵커)을 잡는 능력을 추가합니다.

**구현 (Implement)**
1. MIT 라이선스의 `peter-evans/patience` 구현을 참고하여 `PatienceMatcher` 작성
2. 파일 양측에서 정확히 **한 번씩만 등장하는 유니크 라인**을 필터링 및 추출
3. 추출된 유니크 라인들의 인덱스로 가장 긴 증가하는 부분 수열(LIS) 탐색 후 Patience Sort 진행하여 앵커 기반으로 재귀적 매칭 수행
4. 유니크 라인이 없는 고립 구간에 대해서는 구조적 처리가 어려우므로 `Phase 2`에 만든 `MyersMatcher`로 Fallback 처리.

**검증 (Test)**
* Myers로 감지할 때 메서드의 중괄호(`{`, `}`)간의 엇갈림 식별 에러가 발생하게 하는 리팩토링 테스트 시나리오 작성
* 해당 시나리오를 `PatienceMatcher`가 함수 시그니처 앵커링을 통해 정확히 이동/삭제로 잡아내는지 평가

---

## 🚀 Phase 5: 최종 엔진, Histogram Diff 구현 (최우선 순위 목표)

**계획 (Plan)**
* Patience Diff가 취약한 상황(예: 완전히 동일한 매개변수가 수없이 떨어져 있는 경우, JSON/YAML 등 유니크 라인이 없는 상황)을 극복하고, 빈도 기반의 Histogram을 이용해 이 프로젝트의 **최종 목표 알고리즘**을 달성합니다.

**구현 (Implement)**
1. `PatienceMatcher` 아키텍처 위에 빈도 발생수 카운팅 해시테이블을 덧붙인 `HistogramMatcher` 개발
2. 유니크한 라인이 없을 시, 가장 "적게 등장하는 라인"들을 탐색하고 이를 우선 매칭(Anchor)으로 삼는 최적화 로직 추가
3. 프로젝트의 디폴트 `LineMatcher`를 `HistogramMatcher`로 교체 선언

**검증 (Test)**
* 반복 구조가 많은 `JSON` 혹은 `YAML` 설정 파일에서 중간 노드가 삭제 및 추가되었을 때, Myers나 Patience보다 오차 없이 가장 정확한 줄을 잘린 것으로 나타내는지 벤치마크 테스트
* Histogram Diff와 JetBrains 세부 단위 비교(Char/Word)의 복합 파이프라인 결합 시 최종 산출 트리(AST 혹은 변경 Node 트리) 무결정 입증

---

## 🚀 Phase 6: 다중 포맷 출력 모듈 및 사용자 인터페이스

**계획 (Plan)**
* 파이프라인 처리가 끝난 Diff 데이터를 사용자가 시각적으로, 혹은 도구들이 파싱 가능하게끔 출력단을 정비합니다.

**구현 (Implement)**
1. 두 텍스트와 산출된 매칭 블록들을 받아서 관리하는 통합 `ComparisonManager` 계층 구성
2. 터미널 및 버전 관리 도구에 익숙한 `Unified Diff` (+/-) 출력기 구현
3. 터미널 환경에서 구조적 비교를 돕는 측면형 `Side-by-Side Diff` 렌더링 도입
4. 써드파티 병합 도구가 읽어들일 수 있는 `JSON` 기반 계층적/구조화 포맷 출력 도입

**검증 (Test)**
* End-to-End 테스트 환경 구축: 서로 다른 코드 텍스트 2개 주입 시 의도한 모든 포맷으로 일치하는 값이 출력되는지 확인.
* Jetbrains 라이브러리의 테스트 팩과 결과물 비교(JSON 포맷 스키마 검증).
