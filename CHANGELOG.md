# Changelog

All notable changes to this extension will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## Unreleased

## [v0.8.2](https://github.com/stackus/goht/compare/v0.8.1...v0.8.2) - 2025-07-23

### Fixed

- Accessing slotted templates were inaccessible inside a `@render` command.

### Added

- Slotted templates from a parent template will not be passed down into the child templates. Slotted templates with their own set of slotted templates will take precedence over the parent template's slotted templates.

## [v0.8.1](https://github.com/stackus/goht/compare/v0.8.0...v0.8.1) - 2025-05-06

### Fixed

- Addressed a stack overflow caused by slotted templates incorrectly calling `Render` on themselves and not the original template.

## [v0.8.0](https://github.com/stackus/goht/compare/v0.7.0...v0.8.0) - 2025-04-30

**BREAKING CHANGE**: The signature of the generated Go code has changed. Regenerate all templates prior to this version.

### New Template Composition Feature: Named Slots

Named slots are a new feature that allows you to define places in your templates that you want to populate with any template content. This allows you to create more complex templates that can be reused in different contexts.

```haml
@haml HamlSlots() {
  .basic
    =@slot basic
  .with-default-content
    =@slot defaults
      %span Displayed when nothing is passed in for "defaults"
}
```

```slim
@slim SlimSlots() {
  .basic
    =@slot basic
  .with-default-content
    =@slot defaults
      span Displayed when nothing is passed in for "defaults"
}
```

```html
@ego EgoSlots() {
  <div>
    <%@slot basic %>
  </div>
  <div>
    <%@slot defaults { %>
      <span>Displayed when nothing is passed in for "defaults"</span>
    <% } %>
  </div>
}
```

In the above examples, the slot for "basic" will only be rendered if content is passed in for it.
The slot for "defaults" will fall back to the default content if no content is passed in for it.

#### Using slots in your program
Making use of slots in your program is a simple process:

```go
err := SomeTemplate().Render(ctx, w, 
  OtherTemplate().Slot("basic"),
)
```

Pass in one or more templates into the optional third parameter of the `Render` method.
Then instead of calling `Render` on the slotted template, call `Slot` with the name of the slot you want it to fill.

The slotted templates can be any template, and any template can be used as slotted content, including templates that have their own slots.

```go
err := Layout().Render(ctx, w,
  Sidebar().Slot("sidebar"),
  Header(headerProps).Slot("header"),
  UserDetailsPage(userProps).Slot("main",
    LastActionResults(resultsProps).Slot("notifications"),
  ),
  Footer().Slot("footer"),
)
```

Templates can use slots, `@slot <name>` and the internally rendered templates, `@render SomeTemplate()`
and `@children`, to create templates with incredible levels of reuse and composition.


## [v0.7.0](https://github.com/stackus/goht/compare/v0.6.0...v0.7.0) - 2025-04-29

### New Template: EGO

EGO is a new supported template style that is not whitespace based like the Haml and Slim templates.
It has a syntax similar to [EJS](https://ejs.co/) or [ERB](https://docs.ruby-lang.org/en/2.3.0/ERB.html) template engines.

```html
@ego ExampleEgo(name string) {
  <div>
    <p>Hello <%= name %></p>
    <% for i := 0; i < 10; i++ { %>
      <p>Item <%=%d i %></p>
    <% } -%>
    <%@render ExampleChild() { %>
      <p>Content rendered by child template</p>
    <% } %>
  </div>
}
```

The opening tags that are supported are:
- `<%` - Start of a Go code block
  - Examples: `<% for k, v := range list { %>`, `<% foo := "bar" %>`, `<% if foo == "bar" { %>`
- `<%-` - Start of a Go code block with whitespace stripping
  - Examples: `<%- for k, v := range list { %>`, `<%- foo := "bar" %>`, `<%- if foo == "bar" { %>`
- `<%=` - Start of a Go output block; supports the formatting directives like `%d`, `%v`, etc.
  - Examples: `<%= unsafeHTML %>`, `<%= %t someBool %>`, `<%= props.Value %>`
- `<%!` - Start of a Go unescaped output block; supports the formatting directives like `%d`, `%v`, etc.
  - Examples: `<%! safeHTML %>`, `<%! %t someBool %>`, `<%! props.Value %>`
- `<%@` - Start of a command block; Either `@render` or `@children`
  - Examples: `<%@ render ExampleChild(props ChildProps) { %>`, `<%@ children %>`
- `<%#` - Start of a comment; the content will be ignored
  - Examples: `<%# This is a comment %>`

The closing tags that are supported are:
- `%>` - Normal closing tag
  - Examples: `<% foo := "bar" %>`, `<%= foo %>`
- `-%>` - Closing tag with whitespace stripping
  - Examples: `<% foo := "bar" -%>`, `<%= foo -%>`
- `$%>` - Closing tag with newline stripping (one newline)
  - Examples: `<% foo := "bar" $%>`, `<%= foo $%>`

> Note: Opening and closing braces will be required for the Go code blocks when using EGO. Automatic brace insertion is not intentionally supported, and any support that might remain (all template engines use the same parser) may not continue to work in the future.

### Added

- Added a new template directive `@ego`

### Fixed

- Fixed automatic closing of Go code in the generated code when two unrelated code blocks were siblings in the indented template engines.

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
