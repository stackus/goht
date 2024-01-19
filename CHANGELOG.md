# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [v0.2.1](https://github.com/stackus/hamlet/compare/v0.2.0...v0.2.1) - 2024-01-19

### Fixed
- Fixed an issue with using spaces as the choice of indentation in Haml templates. Previously, the lexer would consume spaces of the first indent within a template.

### Changed
- Improved the logging of the Hamlet CLI

### Added
- Added the version to the Hamlet CLI
- The `generate` command now accepts a `--watch` flag to watch for changes to the input file and automatically regenerate the output file.
- A WIP TextMate grammer for Hamlet has been added to the `/bundle` directory. This is a work in progress and still has some issues. This can be imported into JetBrains IDEs such as GoLand and IntelliJ IDEA.
  - Open the IDE's Settings
  - Navigate to `Editor > TextMate Bundles`
  - Click the `+` button to add a new bundle by selecting the `/bundle` directory.

### Other
- An extension for VSCode which provides syntax highlighting has also been released: [Hamlet (Go+Haml)](https://marketplace.visualstudio.com/items?itemName=stackus.hamlet-go-vscode)

## [v0.2.0](https://github.com/stackus/hamlet/compare/v0.1.0...v0.2.0) - 2024-01-05

### Changed
- Improved Go code syntax in Haml templates: In scenarios where Haml is wrapped in loops or conditions using Go, the requirement for opening and closing braces has been removed. This makes the syntax more streamlined and similar to the original Ruby-based Haml, enhancing readability and ease of use. Existing Go code with braces will continue to work without modifications.

Before:
```haml
- for _, item := range items {
  %li= item
- }
```

After:
```haml
- for _, item := range items
  %li= item
```

## [v0.1.0] - 2024-01-01
The initial release.
