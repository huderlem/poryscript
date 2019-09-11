# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
Nothing, yet.

## [2.1.0] - 2019-09-11
### Added
- Add implicit truthiness checks for `var()` and `flag()` operators.
- Add NOT (`!`) prefix operator for `var()` and `flag()` operators.

### Changed
- Errors are now prefixed with `PORYSCRIPT`, and they are written to `stderr`, instead of `stdout`.
- The program will no longer panic when handled errors occur.

### Fixed
- Fix parser errors that were not showing the line number of the error.

## [2.0.0] - 2019-09-02
### Added
- Add single-line comments with the `#` character.
- Add `go.mod` file so the project can be built outside of the Go workspace.
- Add `while` loops.
- Add `do...while` loops.
- Add `break` and `continue` statements.
- Add compound boolean expressions.
- Add output optimization which significantly simplifies and shrinks the resulting compiled scripts. Turn off optimization by specifying `-optimize=false`.
- Add `switch` statements.

### Changed
- `raw` no longer takes a label name.
- Removed `raw_global`, since there is no longer a concept of being global or local for `raw`.

### Fixed
- Inline texts are now generated with labels that are prefixed to their parent script's name. Otherwise, they would easily clash with external scripts because they were all simply named `Text_<num>`.

## [1.0.0] - 2019-08-27
Initial Release

[Unreleased]: https://github.com/huderlem/poryscript/compare/2.1.0...HEAD
[2.1.0]: https://github.com/huderlem/poryscript/compare/2.0.0...2.1.0
[2.0.0]: https://github.com/huderlem/poryscript/compare/1.0.0...2.0.0
[1.0.0]: https://github.com/huderlem/poryscript/tree/1.0.0
