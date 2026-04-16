# Multi-Language Library Migration Guide (Go & Swift)

이 문서는 하나의 레포지토리 내에서 동일한 기능을 수행하는 Go 라이브러리와 Swift 라이브러리를 공존시키고 관리하기 위한 아키텍처 설계를 설명합니다.

## 1. 프로젝트 구조 (Directory Structure)

전략: **Root Configuration + Sub-folder Source Isolation**
레포지토리의 루트에는 각 언어의 패키지 매니저 설정 파일을 두고, 실제 소스 코드는 각각 `go/`, `swift/` 폴더에서 독립적으로 관리합니다.

```text
.
├── go.mod                  # Go 모듈 정의 (Root)
├── Package.swift           # Swift Package 정의 (Root)
├── Makefile                # 빌드 및 테스트 자동화
├── README.md               # 프로젝트 통합 문서
├── go/                     # Go 라이브러리 소스 및 패키지
│   └── pkg/
│       └── processor/
│           └── processor.go
└── swift/                  # Swift 라이브러리 소스 및 테스트
    ├── Sources/
    │   └── MyLibrary/      # Swift 모듈 소스
    │       └── Processor.swift
    └── Tests/
        └── MyLibraryTests/ # Swift 유닛 테스트
```

## 2. 언어별 설정 상세
### 2.1 Swift (Swift Package Manager)
루트의 Package.swift에서 path 파라미터를 사용하여 하위 디렉토리의 소스를 참조합니다.

```swift
// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "MyCrossPlatformLib",
    products: [
        .library(name: "MyLibrary", targets: ["MyLibrary"]),
    ],
    targets: [
        .target(
            name: "MyLibrary",
            path: "swift/Sources"
        ),
        .testTarget(
            name: "MyLibraryTests",
            dependencies: ["MyLibrary"],
            path: "swift/Tests"
        ),
    ]
)
```

### 2.2 Go (Go Modules)
루트에 go.mod를 배치하여 표준적인 모듈 구조를 유지하되, 소스 경로는 go/ 프리픽스를 포함하게 됩니다.

```go
// go.mod
module [github.com/username/my-cross-platform-project](https://github.com/username/my-cross-platform-project)

go 1.21
```

## 3. 라이브러리 사용법
### 3.1 Swift 프로젝트 (Xcode)

1. Xcode에서 `File > Add Package Dependencies...` 를 선택합니다.
2. 레포지토리 URL을 입력합니다: `https://github.com/username/my-cross-platform-project`
3. Xcode가 루트의 `Package.swift` 를 인식하여 자동으로 `swift/Sources` 의 코드를 라이브러리로 구성합니다.
4. 코드에서 사용: `import MyLibrary`

### 3.2 Go 프로젝트
1. 모듈을 설치합니다: `go get github.com/username/my-cross-platform-project`
2. 코드에서 임포트합니다: `import "github.com/username/my-cross-platform-project/go/pkg/processor"`

## 4. 유지보수 및 버전 관리
- Single Tagging: git tag v1.0.0과 같이 하나의 태그를 생성하면 Go와 Swift 라이브러리에 동일한 버전이 부여됩니다.
- CI/CD: GitHub Actions 등을 통해 go/ 폴더와 swift/ 폴더의 변경사항을 감지하여 독립적인 테스트 파이프라인을 실행합니다.
- Access Control: 
  - Swift: 외부 앱에서 접근해야 하는 인터페이스는 반드시 public 키워드를 사용해야 합니다.
  - Go: 외부에서 접근해야 하는 식별자는 대문자로 시작해야 합니다.

## 5. 설계 결정 사유 (Design Decision)
- 유지보수성: 로직 변경 시 한 레포지토리 내에서 두 언어의 동기화를 즉각적으로 확인할 수 있습니다.
- 표준 준수: 각 언어 생태계(SPM, Go Mod)의 표준 임포트 방식을 해치지 않으면서 모노레포를 구성했습니다.
- 효율성: Go 컴파일러는 빌드 시 .swift 파일을 무시하므로 빌드 성능 저하가 발생하지 않습니다. """
