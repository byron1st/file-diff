// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "FileDiff",
    platforms: [
        .macOS(.v13)
    ],
    products: [
        .library(name: "FileDiff", targets: ["FileDiff"])
    ],
    targets: [
        .target(
            name: "FileDiff",
            path: "swift/Sources/FileDiff"
        ),
        .testTarget(
            name: "FileDiffTests",
            dependencies: ["FileDiff"],
            path: "swift/Tests/FileDiffTests"
        )
    ],
    swiftLanguageModes: [.v6]
)
