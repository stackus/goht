# Changelog

All notable changes to this extension will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## Unreleased

## [v0.6.3](https://github.com/stackus/goht/compare/v0.6.0...v0.6.3) - 2025-04-27

### Added

- Added support to split long code lines across multiple lines using `\` or `,`. to both Haml and Slim parsers.
	- Works for all code tags (`- code`, `= code`, `=@render code`)

### Fixed

- Numerous Slim parsing bugs
- Fixes some issues with the parsing of attributes with Haml and its whitespace handling

## [v0.6.0](https://github.com/stackus/goht/compare/v0.5.0...v0.6.0) - 2025-04-26

### Added

- Added support for [Slim](https://slim-lang.com) templates.

## [v0.5.0](https://github.com/stackus/goht/compare/v0.4.5...v0.5.0) - 2024-06-26

### Changed
- Require tabs for indentation in GoHT templates. This aligns the indentation style with Go's standard formatting.
- Require templates to be indented at a minimum of one level. The initial indent must be one level.

## [v0.4.5](https://github.com/stackus/goht/compare/v0.4.4...v0.4.5) - 2024-02-23

### Fixed
- Parsing `@goht` directives that had `interface{ Method() }` parameter types would fail to parse the function declaration completely.


## [v0.4.4](https://github.com/stackus/goht/compare/v0.4.3...v0.4.4) - 2024-02-23

### Changed
- GoHT templates can be defined with receivers

```haml
@goht (u User) Details() {
.name Name: #{u.Name}
.age Age: #{u.Age}
}
```

### Added
- A new example demonstrating the use of receivers with GoHT templates.

## [v0.4.3](https://github.com/stackus/goht/compare/v0.4.2...v0.4.3) - 2024-02-23

### Changed
- Reduced the amount of code generated for a `@render` command that has no children.

### Fixed
- Log what is available from the client. JetBrains clients are currently not fully identified in the LSP server.

### Added
- Helpers:
	- `goht.If(condition bool, trueResult, falseResult string) string`


## [v0.4.2](https://github.com/stackus/goht/compare/v0.4.1...v0.4.2) - 2024-02-07

### Breaking Changes
- GoHT will now require all templates to start at column 1. Previously, a template could start at any column, and that would be considered to be the base indentation level.
  - This change will allow the detection of inconsistent indentation, swapping between tabs and spaces, and indenting more than one level at a time.

### Fixed
- Removed the extra backslashes around the dynamic attribute values.
- Fixed an issue with parsing attributes that ended with a boolean attribute.
- Removed the rendering of a newline from elements that had no other children than a newline.
- Added missing newlines after HTML comments.
- Added missing newline after :preserve filter output.
- Additional leading whitespace inside the filters will now be kept
- Removed an extra newline after the rendered children output

### Added
- Template tests to check the correctness of the generated Go code.
- Render tests to check the correctness of the rendered HTML output.

## [v0.4.1](https://github.com/stackus/goht/compare/v0.4.0...v0.4.1) - 2024-02-04

### Fixed
- Fixed rendering issues with the three text filters, `:plain`, `:escaped`, and `:preserve`.
  - The `:plain` filter will now correctly unescape the interpolated content.
  - The `:escaped` filter will now correctly escape all content.
  - The `:preserve` filter will now correctly unescape the interpolated content and preserve the newlines.

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
