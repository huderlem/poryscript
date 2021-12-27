# Poryscript

[![Actions Status](https://github.com/huderlem/poryscript/workflows/Go/badge.svg)](https://github.com/huderlem/poryscript/actions) [![codecov](https://codecov.io/gh/huderlem/poryscript/branch/master/graph/badge.svg)](https://codecov.io/gh/huderlem/poryscript)


Use the online [Poryscript Playground](http://www.huderlem.com/poryscript-playground/) to test it out.

Poryscript is a higher-level scripting language that compiles into the scripting language used in [pokeemerald](https://github.com/pret/pokeemerald), [pokefirered](https://github.com/pret/pokefirered), and [pokeruby](https://github.com/pret/pokeruby). It makes scripting faster and easier. Some advantages of using Poryscript are:
1. Branching control flow with `if`, `elif`, `else`, `while`, `do...while`, and `switch` statements.
2. Inline text
3. Auto-formatting text to fit within the in-game text box
4. Better map script organization

View the [Changelog](https://github.com/huderlem/poryscript/blob/master/CHANGELOG.md) to see what's new, and download the latest version from the [Releases](https://github.com/huderlem/poryscript/releases).

**Table of Contents**
- [Usage](#usage)
- [Poryscript Syntax (How to Write Scripts)](#poryscript-syntax-how-to-write-scripts)
  * [`script` Statement](#script-statement)
    + [Boolean Expressions](#boolean-expressions)
    + [`while` and `do...while` Loops](#while-and-dowhile-loops)
    + [Conditional Operators](#conditional-operators)
    + [Regular Commands](#regular-commands)
    + [Early-Exiting a Script](#early-exiting-a-script)
    + [`switch` Statement](#switch-statement)
  * [`text` Statement](#text-statement)
    + [Automatic Text Formatting](#automatic-text-formatting)
    + [Custom Text Encoding](#custom-text-encoding)
  * [`movement` Statement](#movement-statement)
  * [`mapscripts` Statement](#mapscripts-statement)
  * [`raw` Statement](#raw-statement)
  * [Comments](#comments)
  * [Constants](#constants)
  * [Scope Modifiers](#scope-modifiers)
  * [Compile-Time Switches](#compile-time-switches)
  * [Optimization](#optimization)
- [Local Development](#local-development)
  * [Building from Source](#building-from-source)
  * [Running the tests](#running-the-tests)
- [Versioning](#versioning)
- [License](#license)
- [Acknowledgments](#acknowledgments)


# Usage
Poryscript is a command-line program.  It reads an input script and outputs the resulting compiled bytecode script. You can either feed it the input script from a file or from `stdin`.  Similarly, Poryscript can output a file or to `stdout`.

```
> ./poryscript -h
Usage of poryscript:
  -f string
        set default font id (leave empty to use default defined in font widths config file)
  -fw string
        font widths config JSON file (default "font_widths.json")
  -h    show poryscript help information
  -i string
        input poryscript file (leave empty to read from standard input)
  -l int
        set default line length in pixels for formatted text (default 208)
  -o string
        output script file (leave empty to write to standard output)
  -optimize
        optimize compiled script size (To disable, use '-optimize=false') (default true)
  -s value
        set a compile-time switch. Multiple -s options can be set. Example: -s VERSION=RUBY -s LANGUAGE=GERMAN
  -v    show version of poryscript
```

Convert a `.pory` script to a compiled `.inc` script, which can be directly included in a decompilation project:
```
./poryscript -i data/scripts/myscript.pory -o data/scripts/myscript.inc
```

To automatically convert your Poryscript scripts when compiling a decomp project, perform these two steps:
1. Create a new `tools/poryscript/` directory, and add the `poryscript` command-line executable tool to it. Also copy `font_widths.json` to the same location.
```
# For example, on Windows, place the files here.
pokeemerald/tools/poryscript/poryscript.exe
pokeemerald/tools/poryscript/font_widths.json
```
It's also a good idea to add `tools/poryscript` to your `.gitignore` before your next commit.

2. Update the Makefile with these changes (Note, don't add the `+` symbol at the start of the lines. That's just to show the line is being added.):
```diff
+ SCRIPT := tools/poryscript/poryscript$(EXE)
```
```diff
mostlyclean: tidy
	rm -f sound/direct_sound_samples/*.bin
	rm -f $(MID_SUBDIR)/*.s
	find . \( -iname '*.1bpp' -o -iname '*.4bpp' -o -iname '*.8bpp' -o -iname '*.gbapal' -o -iname '*.lz' -o -iname '*.latfont' -o -iname '*.hwjpnfont' -o -iname '*.fwjpnfont' \) -exec rm {} +
	rm -f $(AUTO_GEN_TARGETS)
+	rm -f $(patsubst %.pory,%.inc,$(shell find data/ -type f -name '*.pory'))
```
```diff
%.s: ;
%.png: ;
%.pal: ;
%.aif: ;
+ %.pory: ;
```
```diff
sound/%.bin: sound/%.aif ; $(AIF) $< $@
+ data/%.inc: data/%.pory; $(SCRIPT) -i $< -o $@ -fw tools/poryscript/font_widths.json
```
```diff
-TOOLDIRS := $(filter-out tools/agbcc tools/binutils,$(wildcard tools/*))
+TOOLDIRS := $(filter-out tools/agbcc tools/binutils tools/poryscript,$(wildcard tools/*))
```

3. Update `make_tools.mk` with the same change:
```diff
-TOOLDIRS := $(filter-out tools/agbcc tools/binutils,$(wildcard tools/*))
+TOOLDIRS := $(filter-out tools/agbcc tools/binutils tools/poryscript,$(wildcard tools/*))
```

# Poryscript Syntax (How to Write Scripts)

A single `.pory` file is composed of many top-level statements. The valid top-level statements are `script`, `text`, `movement`, `mapscripts`, and `raw`.
```
mapscripts MyMap_MapScripts {
    ...
}

script MyScript {
    ...
}

text MyText {
    "Hi, I'm some text.\n"
    "I'm global and can be accessed in C code."
}

movement MyMovement {
    walk_left
    walk_right * 3
}

raw `
MyLocalText:
    .string "I'm directly included.$"
`
```

## `script` Statement
The `script` statement creates a global script containing script commands and control flow logic.  Here is an example:
```
script MyScript {
    # Show a different message, depending on the badges the player owns.
    lock
    faceplayer
    if (flag(FLAG_RECEIVED_TOP_PRIZE)) {
        msgbox("You received the best prize!")
    } elif (flag(FLAG_RECEIVED_WORST_PRIZE)) {
        msgbox("Ouch, you received the worst prize.")
    } else {
        msgbox("Hmm, you didn't receive anything.")
    }
    release
    end
}
```

As you can see, using `if` statements greatly simplifies writing scripts because it does not require the author to manually define new sub-labels with `goto` statements everywhere.

`if` statements can be nested inside each other, as you would expect.
```
    if (flag(FLAG_TEMP) == true) {
        if (var(VAR_BADGES) < 8) {
            ...
        } else {
            ...
        }
    }
```

Note the special keyword `elif`.  This is just the way Poryscript specifies an "else if". Many `elif` statements can be chained together.

### Boolean Expressions
Compound boolean expressions are also supported. This means you can use the AND (`&&`) and OR (`||`) logical operators to combine expressions. For example:
```
    # Basic AND of two conditions.
    if (!defeated(TRAINER_MISTY) && var(VAR_TIME) != DAY) {
        msgbox("The Cerulean Gym's doors don't\n"
               "open until morning.")
    }
    ...
    # Group nested conditions together with another set of parentheses.
    if (flag(FLAG_IS_CHAMPION) && !(flag(FLAG_SYS_TOWER_GOLD) || flag(FLAG_SYS_DOME_GOLD))) {
        msgbox("You should try to beat the\n"
               "Battle Tower or Battle Dome!")
    }
```

### `while` and `do...while` Loops
`while` statements are used to do loops.  They can be nested inside each or inside `if` statements, as one would expect.
```
    # Force player to answer "Yes" to NPC question.
    msgbox("Do you agree to the quest?", MSGBOX_YESNO)
    while (var(VAR_RESULT) != 1) {
        msgbox("...How about now?", MSGBOX_YESNO)
    }
    setvar(VAR_QUEST_ACCEPTED, 1)
```

The `while` statement can also be written as an infinite loop by omitting the boolean expression. This would be equivalent to `while(true)` in typical programming languages. (Of course, you'll want to `break` out of the infinite loop, or hard-stop the script.)
```
    while {
        msgbox("Want to see this message again?", MSGBOX_YESNO")
        if (var(VAR_RESULT) != 1) {
            break
        }
    }
```

`do...while` statements are very similar to `while` statements.  The only difference is that they always execute their body once before checking the condition.
```
    # Force player to answer "Yes" to NPC question.
    do {
        msgbox("Can you help me solve the puzzle?", MSGBOX_YESNO)
    } while (var(VAR_RESULT) == 0)
```

`break` can be used to break out of a loop, like many programming languages. Similary, `continue` returns to the start of the loop.

### Conditional Operators
The condition operators have strict rules about what conditions they accept. The operand on the left side of the condition must be a `flag()`, `var()`, or `defeated()` check. They each have a different set of valid comparison operators, described below.

| Type | Valid Operators |
| ---- | --------------- |
| `flag` | `==` |
| `var` | `==`, `!=`, `>`, `>=`, `<`, `<=` |
| `defeated` | `==` |

All operators support implicit truthiness, which means you don't have to specify any of the above operators in a condition. Below are some examples of equivalent conditions:
```
# Check if the flag is set.
if (flag(FLAG_1))
if (flag(FLAG_1) == true)

# Check if the flag is cleared.
if (!flag(FLAG_1))
if (flag(FLAG_1) == false)

# Check if the var is not equal to 0.
if (var(VAR_1))
if (var(VAR_1) != 0)

#Check if the var is equal to 0.
if (!var(VAR_1))
if (var(VAR_1) == 0)

# Check if the trainer has been defeated.
if (defeated(TRAINER_GARY))
if (defeated(TRAINER_GARY) == true)

# Check if the trainer hasn't been defeated.
if (!defeated(TRAINER_GARY))
if (defeated(TRAINER_GARY) == false)
```

When not using implicit truthiness, like in the above examples, they each have different valid comparison values on the right-hand side of the condition.

| Type | Valid Comparison Values |
| ---- | --------------- |
| `flag` | `TRUE`, `true`, `FALSE`, `false` |
| `var` | any value (e.g. `5`, `VAR_TEMP_1`, `VAR_FOO + BASE_OFFSET`) |
| `defeated` | `TRUE`, `true`, `FALSE`, `false` |

One quirk of the Gen 3 decomp scripting engine is that using the `compare` scripting command with a value in the range `0x4000 <= x <= 0x40FF` or `0x8000 <= x <= 0x8015` will result in comparing agains a `var`, rather than the raw value. To force the comparison against a raw value, like `0x4000`, use the `value()` operator.  For example:

```
if (var(VAR_DAMAGE_DEALT) >= value(0x4000))
```

The resulting script be use the `compare_var_to_value` command, rather than the usual `compare` command.

### Regular Commands
Regular non-branching commands that take arguments, such as `msgbox`, must wrap their arguments in parentheses. For example:
```
    lock
    faceplayer
    addvar(VAR_TALKED_COUNT, 1)
    msgbox("Hello.")
    release
    end
```

### Early-Exiting a Script
Use `end` or `return` to early-exit out of a script.
```
script MyScript {
    if (flag(FLAG_WON) == true) {
        end
    }
    ...
}
```

### `switch` Statement
A `switch` statement is an easy way to separate different logic for a set of concrete values. Poryscript `switch` statements behave similarly to other languages. However, the cases `break` implicitly. It is not possible to "fall through" to the next case by omitting a `break` at the end of a case, like in C. You *can* use `break` to break out of a case, though--it's just not required. Multiple cases can be designated by listing them immediately after another without a body. Finally, an optional `default` case will take over if none of the provided `case` values are met.  A `switch` statement's comparison value *must always be a `var()` operator*.  Of course, `switch` statements can appear anywhere in the script's logic, such as inside `while` loops, or even other `switch` statements.

```
    switch (var(VAR_NUM_THINGS)) {
        case 0:
            msgbox("You have 0 things.")
        case 1:
        case 2:
            msgbox("You have 1 or 2 things.")
        default:
            msgbox("You have at least 3 things.")
    }
```

## `text` Statement
Use `text` to include text that's intended to be shared between multiple scripts or in C code. The `text` statement is just a convenient way to write chunks of text, and it exports the text globally, so it is accessible in C code. Currently, there isn't much of a reason to use `text`, but it will be more useful in future updates of Poryscript.
```
script MyScript {
    msgbox(MyText)
}

text MyText {
    "Hello, there.\p"
    "You can refer to me in scripts or C code."
}
```
A small quality-of-life feature is that Poryscript automatically adds the `$` terminator character to text, so the user doesn't need to manually type it all the time.

### Automatic Text Formatting
Text auto-formatting is also supported by Poryscript. The `format()` function can be wrapped around any text, either inline or `text`, and Poryscript will automatically fit the text to the size of the in-game text window by inserting automatic line breaks. A simple example:
```
msgbox(format("Hello, this is some long text that I want Poryscript to automatically format for me."))
```
Becomes:
```
.string "Hello, this is some long text that I\n"
.string "want Poryscript to automatically\l"
.string "format for me.$"
```
Like other text, formatted text can span multiple lines if you use a new set of quotes for each line. You can also manually add your own line breaks (`\p`, `\n`, `\l`), and it will still work as expected.
```
text MyText {
    format("Hello, are you the real-live legendary {PLAYER} that everyone talks about?\p"
           "Amazing!\pSo glad to meet you!")
}
```
Becomes:
```
.string "Hello, are you the real-live legendary\n"
.string "{PLAYER} that everyone talks about?\p"
.string "Amazing!\p"
.string "So glad to meet you!$"
```
The font id can optionally be specified as the second parameter to `format()`.
```
text MyText {
    format("Hello, are you the real-live legendary {PLAYER} that everyone talks about?\pAmazing!\pSo glad to meet you!", "1_latin")
}
```
Becomes:
```
.string "Hello, are you the real-live legendary\n"
.string "{PLAYER} that everyone talks about?\p"
.string "Amazing!\p"
.string "So glad to meet you!$"
```
The font widths configuration JSON file informs Poryscript how many pixels wide each character in the message is. Different fonts have different character widths. For convenience, Poryscript comes with `font_widths.json`, which contains the configuration for pokeemerald's `1_latin` font. More fonts can easily be added to this file by the user by creating anothing font id node under the `fonts` key in `font_widths.json`.

The length of a line can optionally be specified as the third parameter to `format()` if a font id was specified as the second parameter.

```
text MyText {
    format("Hello, are you the real-live legendary {PLAYER} that everyone talks about?\pAmazing!\pSo glad to meet you!", "1_latin", 100)
}
```
Becomes:
```
.string "Hello, are you the\n"
.string "real-live\l"
.string "legendary\l"
.string "{PLAYER} that\l"
.string "everyone talks\l"
.string "about?\p"
.string "Amazing!\p"
.string "So glad to meet\n"
.string "you!$"
```

### Custom Text Encoding
When Poryscript compiles text, the resulting text content is rendered using the `.string` assembler directive. The decomp projects' build process then processes those `.string` directives and substituted the string characters with the game-specific text representation. It can be useful to specify different types of strings, though. For example, implementing print-debugging commands might make use of ASCII text. Poryscript allows you to specify which assembler directive to use for text. Simply add the directive as a prefix to the string content like this:
```
ascii"My ASCII string."
custom"My Custom string."

// compiles to...
.ascii "My ASCII string.\0"
.custom "My Custom string."
```

Note that Poryscript will automatically add the `\0` suffix character to ASCII strings. It will **not** add suffix to any other directives.

## `movement` Statement
Use `movement` statements to conveniently define movement data that is typically used with the `applymovement` command. `*` can be used as a shortcut to repeat a single command many times. Data defined with `movement` is created with local scope, not global.
```
script MyScript {
    lock
    applymovement(2, MyMovement)
    waitmovement(0)
    release
}
movement MyMovement {
    walk_left
    walk_up * 5
    face_down
}
```
Becomes:
```
MyScript::
	lock
	applymovement 2, MyMovement
	waitmovement 0
	release
	return

MyMovement:
	walk_left
	walk_up
	walk_up
	walk_up
	walk_up
	walk_up
	face_down
	step_end
```

## `mapscripts` Statement
Use `mapscripts` to define a set of map script definitions. Scripts can be inlined for convenience, or a label to another script can simply be specified. Some map script types, like `MAP_SCRIPT_ON_FRAME_TABLE`, require a list of comparison variables and scripts to execute when the variable's value is equal to some value. In these cases, you use brackets `[]` to specify that list of scripts. Below is a full example showing map script definitions for a new map called `MyNewCity`:
```
mapscripts MyNewCity_MapScripts {
    MAP_SCRIPT_ON_RESUME: MyNewCity_OnResume
    MAP_SCRIPT_ON_TRANSITION {
        random(2)
        switch (var(VAR_RESULT)) {
            case 0: setweather(WEATHER_ASH)
            case 1: setweather(WEATHER_RAIN_HEAVY)
        }
    }
    MAP_SCRIPT_ON_FRAME_TABLE [
        VAR_TEMP_0, 0: MyNewCity_OnFrame_0
        VAR_TEMP_0, 1 {
            lock
            msgbox("This script is inlined.")
            setvar(VAR_TEMP_0, 2)
            release
        }
    ]
}

script MyNewCity_OnResume {
    ...
}

script MyNewCity_OnFrame_0 {
    ...
}
```

For maps with no map scripts, simply make an empty `mapscripts` statement:
```
mapscripts MyNewCity_MapScripts {}
```

## `raw` Statement
Use `raw` to include raw bytecode script. Anything in a `raw` statement will be directly included into the compiled script. This is useful for defining custom data, or data types not supported in regular Poryscript.
```
raw `
TestMap_MapScripts::
	.byte 0
`

script MyScript {
    lock
    faceplayer
    # Text can span multiple lines. Use a new set of quotes for each line.
    msgbox("This is shorter text,\n"
           "but we can still put it\l"
           "on multiple lines.")
    applymovement(OBJ_EVENT_ID_PLAYER, MyScript_Movement)
    waitmovement(0)
    msgbox(MyScript_LongText)
    release
    end
}

raw `
MyScript_Movement:
    walk_left
    walk_down
    step_end

MyScript_LongText:
    .string "Hi, there.\p"
    .string "This text is too long\n"
    .string "to inline above.\p"
    .string "We'll put it down here\n"
    .string "instead, so it's out of\l"
    .string "the way.$"
`
```

## Comments
Use single-line comments with `#` or `//`. Everything after the `#` or `//` will be ignored. Comments cannot be placed in a `raw` statement. (Users who wish to run the C preprocessor on Poryscript files should use `//` comments to avoid conflict with C preprocessor directives that use the `#` character.)
```
# This script does some cool things.
script MyScript {
    // This is also a valid comment.
    ...
}
```

## Constants
Use `const` to define constants that can be used in the current script. This is especially useful for giving human-friendly names to event object ids, or temporary flags. Constants must be defined before they are used. Constants can also be composed of previously-defined constants.
```
const PROF_BIRCH_ID = 3
const ASSISTANT_ID = PROF_BIRCH_ID + 1
const FLAG_GREETED_BIRCH = FLAG_TEMP_2

script ProfBirchScript {
    applymovement(PROF_BIRCH_ID, BirchMovementData)
    showobject(ASSISTANT_ID)
    setflag(FLAG_GREETED_BIRCH)
}
```

Note that these constants are **not** a general macro system. They can only be used in certain places in Poryscript syntax. Below is an example of all possible places where constants can be substituted into the script:
```
const CONSTANT = 1

mapscripts MyMapScripts {
    MAP_SCRIPT_ON_FRAME_TABLE [
        // The operand and comparison values can both use constants in a
        // table-based map script.
        CONSTANT, CONSTANT: MyOnFrameScript_0
    ]
}

script MyScript {
    // Any parameter of any command can use constants.
    somecommand(CONSTANT)

    // Any comparison operator can use constants, as well as their comparison values.
    if (flag(CONSTANT)) {}
    if (var(CONSTANT) == CONSTANT) {}
    if (defeated(CONSTANT)) {}

    // A switch var value can be a constant, as well as the individual cases.
    switch (var(CONSTANT)) {
        case CONSTANT: break
    }
}
```

## Scope Modifiers
To control whether a script should be global or local, a scope modifier can be specified. This is supported for `script`, `text`, `movement`, and `mapscripts`. In this context, "global" means that the label will be defined with two colons `::`.  Local scopes means one colon `:`.
```
script(global) MyGlobalScript {
    ...
}
script(local) MyLocalScript {
    ...
}
```
Becomes:
```
MyGlobalScript::
    ...

MyLocalScript:
    ...
```

The top-level statements have different default scopes. They are as follows:

| Type | Default Scope |
| ---- | --------------- |
| `script` | Global |
| `text` | Global |
| `movement` | Local |
| `mapscripts` | Global |

## Compile-Time Switches
Use the `poryswitch` statement to change compiler behavior depending on custom switches. This makes it easy to make scripts behave different depending on, say, the `GAME_VERSION` or `LANGUAGE`. Any content that does not match the compile-time switch will not be included in the final output. To define custom switches, use the `-s` option when running `poryscript`.  You can specify multiple switches, and each key/value pair must be separated by an equals sign. For example:

```
./poryscript -i script.pory -o script.inc -s GAME_VERSION=RUBY -s LANGUAGE=GERMAN
```

The `poryswitch` statement can be embedded into any script section, including `text` and `movement` statements. The underscore `_` case is used as the fallback, if none of the other cases match. Cases that only contain a single statement or command can be started with a colon `:`.  Otherwise, use curly braces to define the case's block.

Here are some examples of compile-time switches. This assumes that two compile-time switches are defined, `GAME_VERSION` and `LANGUAGE`.

```
script MyScript {
    lock
    faceplayer
    poryswitch(GAME_VERSION) {
        RUBY {
            msgbox("Here, take this Ruby Orb.")
            giveitem(ITEM_RUBY_ORB)
        }
        SAPPHIRE {
            msgbox("Here, take this Sapphire Orb.")
            giveitem(ITEM_SAPPHIRE_ORB)
        }
        _: msgbox(format("This case is used when GAME_VERSION doesn't match either of the above."))
    }
    release
}

text MyText {
    poryswitch(LANGUAGE) {
        GERMAN:  "Hallo. Ich spreche Deutsch."
        ENGLISH: "Hello. I speak English."
    }
}

movement MyMovement {
    face_player
    walk_down
    poryswitch(GAME_VERSION) {
        RUBY: walk_left * 2
        SAPPHIRE {
            walk_right * 2
            walk_left * 4
        }
    }
}
```

Note, `poryswitch` can also be embedded inside inlined `mapscripts` scripts.

## Optimization
By default, Poryscript produces optimized output. It attempts to minimize the number of `goto` commands and unnecessary script labels. To disable optimizations, pass the `-optimize=false` option to `poryscript`.

# Local Development

These instructions will get you setup and working with Poryscript's code. You can either build the Poryscript tool from source, or simply download the latest release from the Releases tab on GitHub.

## Building from Source

First, install [Go](http://golang.org).  Poryscript has no additional dependencies.  It uses Go modules, so you shouldn't need to be located in a Go workspace.

Navigate to the Poryscript working directory, and build it:
```
cd your/path/to/poryscript
go build
```

This will create a `poryscript` executable binary in the same directory. Then you can simply install it into your project by running `./install.sh ../yourprojectname` instead of manually copying the files over, similarly to how agbcc is installed into projects.

## Running the tests

Poryscript has automated tests for its `emitter`, `parser`, and `lexer` packages. To run all of the tests from the base directory:
```
> go test ./...
?       github.com/huderlem/poryscript  [no test files]
?       github.com/huderlem/poryscript/ast      [no test files]
ok      github.com/huderlem/poryscript/emitter  0.523s
ok      github.com/huderlem/poryscript/lexer    0.273s
ok      github.com/huderlem/poryscript/parser   0.779s
?       github.com/huderlem/poryscript/token    [no test files]
```


# Versioning

Poryscript uses [Semantic Versioning](http://semver.org/). For the available versions, see the [tags on this repository](https://github.com/huderlem/poryscript/tags).

# License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

# Acknowledgments

* Thorsten Ball's *Writing An Interpreter In Go* helped bootstrap the lexer, AST, and parser for this project. A chunk of that code was derived and/or copied from that book, as I had never written something of this nature before.
