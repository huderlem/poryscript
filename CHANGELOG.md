# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
Nothing, yet.

## [2.12.0] - 2021-12-27
### Added
- Add `value()` operator, which can be used on the right-hand side of a `var()` comparison. It will force a `compare_var_to_value` command to be output. This makes it possible to compare values that occupy the same range as vars (`0x4000 <= x <= 0x40FF` and `0x8000 <= x <= 0x8015`).
- Add ability to author inifinite loops using the `while` statement without any boolean expression.

## [2.11.0] - 2021-10-23
### Added
- Add -l command-line option to define default line length for formatted text.
- Add -f command-line option to define default font id from `font_widths.json` for formatted text.

## [2.10.0] - 2021-04-03
### Added
- Add ability to specify custom directives for text. (e.g. `ascii"My ASCII text"` will result in `.ascii "My ASCII text\0"`)

## [2.9.0] - 2020-09-07
### Added
- Add optional maximum line length parameter to `format()` operator.

## [2.8.1] - 2020-05-06
### Fixed
- Fix bug where `switch` statement `default` case didn't work properly when combined with other cases.

## [2.8.0] - 2020-03-25
### Added
- Add ability to use the NOT (`!`) operator in front of nested boolean expressions. Example: `if (flag(A) && !(flag(B) || flag(C)))`

## [2.7.2] - 2019-11-16
### Fixed
- Fix bug where implicit text labels weren't properly inserted into command arguments.

## [2.7.1] - 2019-11-13
### Fixed
- Fix bug where control codes with spaces in them (e.g. `{COLOR BLUE}`) were not handled properly in `format()`.

## [2.7.0] - 2019-11-05
### Added
- Add support for compile-time switches using the `poryswitch` statement. This helps with language differences or game-version differences, for example.
- Add support for user-defined constants with `const` keyword. This helps with things like defining event object ids to refer to throughout the script.

## [2.6.0] - 2019-10-26
### Added
- Add support for scope modifiers `global` and `local` for `script`, `text`, `movement`, and `mapscripts` statements. This will force labels for be generated with `::` (global) or `:` (local) in the compiled output script.

## [2.5.0] - 2019-10-16
### Added
- Comments can now be used with `//`, in addition to the existing '#' style. This is to support users who want to process Poryscript with the C preprocessor.
- Add `movement` statement, which is used to define movement data. Use `*` as a shortcut for repeating a movement command many times. `step_end` terminator is automatically added to the end of the data.
- Add `mapscripts` statement, which is used to define map scripts. Scripts can be inlined, or simply specified with a label.

### Fixed
- Fix harmless bug where `format()` could result in empty `.string ""` lines in the compiled out.
- Fix bug where `end` command was incorrectly being replaced with a `return`.
- Fix bug where negative numbers were not parsed correctly.

## [2.4.0] - 2019-10-13
### Added
- Add support for text auto-formatting with the `format()` operator. Font widths are loaded from a config JSON file. Specify config file with `-fw <config filepath>`. If `-fw` is omitted, Poryscript will try to load `font_widths.json` by default.

### Changed
- Text is now automatically terminated with a `$` character, so the user doesn't have to manually type it for all pieces of text. Of course, this does not apply to text within `raw` statements.

## [2.3.0] - 2019-10-12
## Added
- Add `defeated()` operator, which is used to check if a trainer has been defeated. Without this new `defeated()` operator, it was impossible to write scripts that checked trainer flags without using `raw`.
- Add `text` statements.

## [2.2.0] - 2019-10-07
### Changed
- Identical implicit texts are now combined into a single text output.

### Fixed
- Fix some potential infinite loops when parsing certain invalid scripts.

## [2.1.1] - 2019-09-14
### Fixed
- Fix bug where hexadecimal numbers were not tokenized correctly, resulting in a space after the `0x` prefix.

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

[Unreleased]: https://github.com/huderlem/poryscript/compare/2.12.0...HEAD
[2.12.0]: https://github.com/huderlem/poryscript/compare/2.11.0...2.12.0
[2.11.0]: https://github.com/huderlem/poryscript/compare/2.10.0...2.11.0
[2.10.0]: https://github.com/huderlem/poryscript/compare/2.9.0...2.10.0
[2.9.0]: https://github.com/huderlem/poryscript/compare/2.8.1...2.9.0
[2.8.1]: https://github.com/huderlem/poryscript/compare/2.8.0...2.8.1
[2.8.0]: https://github.com/huderlem/poryscript/compare/2.7.2...2.8.0
[2.7.2]: https://github.com/huderlem/poryscript/compare/2.7.1...2.7.2
[2.7.1]: https://github.com/huderlem/poryscript/compare/2.7.0...2.7.1
[2.7.0]: https://github.com/huderlem/poryscript/compare/2.6.0...2.7.0
[2.6.0]: https://github.com/huderlem/poryscript/compare/2.5.0...2.6.0
[2.5.0]: https://github.com/huderlem/poryscript/compare/2.4.0...2.5.0
[2.4.0]: https://github.com/huderlem/poryscript/compare/2.3.0...2.4.0
[2.3.0]: https://github.com/huderlem/poryscript/compare/2.2.0...2.3.0
[2.2.0]: https://github.com/huderlem/poryscript/compare/2.1.1...2.2.0
[2.1.1]: https://github.com/huderlem/poryscript/compare/2.1.0...2.1.1
[2.1.0]: https://github.com/huderlem/poryscript/compare/2.0.0...2.1.0
[2.0.0]: https://github.com/huderlem/poryscript/compare/1.0.0...2.0.0
[1.0.0]: https://github.com/huderlem/poryscript/tree/1.0.0
