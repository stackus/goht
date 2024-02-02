# Contributing to Hamlet

We are thrilled that you are interested in contributing to Hamlet! Hamlet is a Haml engine for Go, focused on compiling Haml into type-safe Go code. This document provides guidelines for contributing to various parts of the project, including the CLI, compiler, and runtime.

## Table of Contents

- [Getting Started](#getting-started)
- [Contributing to Different Sections](#contributing-to-different-sections)
	- [CLI](#cli)
	- [Compiler](#compiler)
	- [Runtime](#runtime)
  - [Documentation](#documentation)
- [Bug Submissions](#bug-submissions)
- [Pull Requests](#pull-requests)
- [Adhering to the Haml Spec](#adhering-to-the-haml-spec)
- [Coding Standards](#coding-standards)

## Getting Started

Before you begin, please ensure you have a GitHub account and are familiar with the basics of making a pull request. If you are new to Git or GitHub, we recommend reviewing [GitHub's documentation](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests).

## Contributing to Different Sections

### CLI
- Share your ideas for improving the CLI experience.
- Contribute to the development or debugging of CLI features.

### Compiler
- Help enhance the compiler’s efficiency and reliability.
- Work on feature additions or bug fixes related to the compiler.
- Improve code coverage of the compiler, covering corner and edge cases.

### Runtime
- Participate in optimizing runtime performance.
- Contribute to making the runtime more robust and fault-tolerant.

### Documentation
- Help write and improve the documentation.
- Contribute to the [examples](examples) directory.

## IDE Extensions
- Contribute to the development of extensions and plugins for Hamlet in various editors and IDEs that will use the built-in Language Server Protocol (LSP) support.

## Bug Submissions

We welcome bug reports! If you've found a bug in Hamlet, please submit it as an issue in our GitHub repository. Include as much detail as possible, such as:

- A clear and concise description of the bug.
- Steps to reproduce the bug.
- Expected and actual behavior.
- Screenshots or code snippets, if applicable.

## Pull Requests

Contributions to fix bugs or add features are made through pull requests (PRs). Here's how you can submit a PR:

1. Fork the repository and create your branch from `master`.
2. Make your changes, ensuring they adhere to the project's coding standards.
3. Write tests for your changes and ensure that all tests pass.
4. Submit a pull request with a clear description of your changes.

## Adhering to the Haml Spec

It is crucial for Hamlet to stick as closely as possible to the Haml specification. However, due to differences in syntax and structure between Go and Ruby, some deviations are inevitable. When contributing, consider the following:

- Strive for consistency with the Haml spec.
- Document any necessary deviations due to language differences.

## Coding Standards

- Write clean, readable, and well-documented code.
- Follow Go's standard coding conventions.
- Include tests for new features or bug fixes.

Thank you for contributing to Hamlet! Your efforts help make this project better for everyone.
