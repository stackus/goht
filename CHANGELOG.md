# Changelog

All notable changes to this extension will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## Unreleased

## [v0.4.0](https://github.com/stackus/goht/compare/v0.3.0...v0.4.0) - 2024-02-03

![GoHT](docs/goht_header.png)

### Changed
- **Project Rename**: Renamed the project from `Hamlet` to `GoHT`. This was done to give the project a unique name and to avoid any potential confusion with other projects named Hamlet. Renaming the project while it's still early means that the impact of the change is minimal. The CLI and package names have been updated to reflect this change.
  - The major version number will not be bumped for this change because we are still in a development release. [Semver Rule #4](https://semver.org/#spec-item-4)
  - The fallout of this renaming will be if it evers gets traction with others they'll be left wondering if the project is pronounced like "goat" or "got". My answer is it's like "goat".
- The VSCode extension has been updated to reflect the new project name and logo.
  - The latest version will be using the GoHT CLI

## [v0.3.0](https://github.com/stackus/hamlet/compare/v0.2.1...v0.3.0) - 2024-02-01

### Added
- Added a `lsp` command to the Hamlet CLI. See `hamlet help lsp` for more information. This will enable development of extensions and plugins for Hamlet in various editors and IDEs. 

### Fixed
- Fixed a parsing issue when there are multiple imports and either all or some of the imports are named.
- Fixed a parsing issue when the template file contained comments that came before the package declaration.
  - The comments will not remain at the top of the generated file and will be placed after the package and import declarations.

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
