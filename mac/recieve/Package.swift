// swift-tools-version:5.8
import PackageDescription

let package = Package(
    name: "recieve",
    platforms: [
        .macOS(.v10_14)
    ],
    dependencies: [],
    targets: [
        .executableTarget(
            name: "recieve",
            dependencies: []
        ),
    ]
)