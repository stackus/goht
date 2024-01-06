# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

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
