# Poryscript

[![Actions Status](https://github.com/huderlem/poryscript/workflows/Go/badge.svg)](https://github.com/huderlem/poryscript/actions) [![codecov](https://codecov.io/gh/huderlem/poryscript/branch/master/graph/badge.svg)](https://codecov.io/gh/huderlem/poryscript)


Use the online [Poryscript Playground](http://www.huderlem.com/poryscript-playground/) to test it out.

Poryscript is a higher-level scripting language that compiles into the scripting language used in [pokeemerald](https://github.com/pret/pokeemerald), [pokefirered](https://github.com/pret/pokefirered), and [pokeruby](https://github.com/pret/pokeruby). It makes scripting faster and easier. The main advantages to using Poryscript are:
1. Branching control flow with `if`, `elif`, `else`, `while`, `do...while`, and `switch` statements.
2. Inline text

View the [Changelog](https://github.com/huderlem/poryscript/blob/master/CHANGELOG.md) to see what's new, and download the latest version from the [Releases](https://github.com/huderlem/poryscript/releases).


# Usage
Poryscript is a command-line program.  It reads an input script and outputs the resulting compiled bytecode script. You can either feed it the input script from a file or from `stdin`.  Similarly, Poryscript can output a file or to `stdout`.

```
> ./poryscript -h
Usage of poryscript:
  -h    show poryscript help information
  -i string
        input poryscript file (leave empty to read from standard input)
  -o string
        output script file (leave empty to write to standard output)
  -optimize
        optimize compiled script size (To disable, use '-optimize=false') (default true)
  -v    show version of poryscript
```

Convert a `.pory` script to a compiled `.inc` script, which can be directly included in a decompilation project:
```
./poryscript -i data/scripts/myscript.pory -o data/scripts/myscript.inc
```

To automatically convert your Poryscript scripts when compiling a decomp project, perform these two steps:
1. Add the `poryscript` command-line executable tool to the `tools` directory:
```
# For example, on Windows, place the file here.
pokeemerald/tools/poryscript/poryscript.exe
```
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
+ data/%.inc: data/%.pory; $(SCRIPT) -i $< -o $@
```

# Poryscript Syntax (How to Write Scripts)

A single `.pory` file is composed of many top-level statements. The valid top-level statements are `script`, `text`, and `raw`.
```
script MyScript {
    ...
}

text MyText {
    "Hi, I'm some text.\n"
    "I'm global and can be accessed in C code.$"
}

raw `
MyLocalText:
    .string "I'm directly included."
`
```

The `script` statement creates a global script containing script commands and control flow logic.  Here is an example:
```
script MyScript {
    # Show a different message, depending on the badges the player owns.
    lock
    faceplayer
    if (flag(FLAG_RECEIVED_TOP_PRIZE)) {
        msgbox("You received the best prize!$")
    } elif (flag(FLAG_RECEIVED_WORST_PRIZE)) {
        msgbox("Ouch, you received the worst prize.$")
    } else {
        msgbox("Hmm, you didn't receive anything.$")
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

Compound boolean expressions are also supported. This means you can use the AND (`&&`) and OR (`||`) logical operators to combine expressions. For example:
```
    # Basic AND of two conditions.
    if (!defeated(TRAINER_MISTY) && var(VAR_TIME) != DAY) {
        msgbox("The Cerulean Gym's doors don't\n"
               "open until morning.$")
    }
    ...
    # Group nested conditions together with another set of parentheses.
    if (flag(FLAG_IS_CHAMPION) && (flag(FLAG_SYS_TOWER_GOLD) || flag(FLAG_SYS_DOME_GOLD))) {
        msgbox("Wow, you're a Battle Frontier pro!$)
    }
```

`while` statements are used to do loops.  They can be nested inside each or inside `if` statements, as one would expect.
```
    # Force player to answer "Yes" to NPC question.
    msgbox("Do you agree to the quest?$", MSGBOX_YESNO)
    while (var(VAR_RESULT) != 1) {
        msgbox("...How about now?$", MSGBOX_YESNO)
    }
    setvar(VAR_QUEST_ACCEPTED, 1)
```

`do...while` statements are very similar to `while` statements.  The only difference is that they always execute their body once before checking the condition.
```
    # Force player to answer "Yes" to NPC question.
    do {
        msgbox("Can you help me solve the puzzle?$", MSGBOX_YESNO)
    } while (var(VAR_RESULT) == 0)
```

`break` can be used to break out of a loop, like many programming languages. Similary, `continue` returns to the start of the loop.

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

# The NOT operator (!) can only be used directly before an
# operator, unlike regular programming languages.
# This code would be an error, because the NOT operator is not
# applying directly to a single flag() operator.
if (!(flag(FLAG_1) && flag(FLAG_2)))
```

When not using implicit truthiness, like in the above examples, they each have different valid comparison values on the right-hand side of the condition.

| Type | Valid Comparison Values |
| ---- | --------------- |
| `flag` | `TRUE`, `true`, `FALSE`, `false` |
| `var` | any value (e.g. `5`, `VAR_TEMP_1`, `VAR_FOO + BASE_OFFSET`) |
| `defeated` | `TRUE`, `true`, `FALSE`, `false` |

Regular non-branching commands that take arguments, such as `msgbox`, must wrap their arguments in parentheses. For example:
```
    lock
    faceplayer
    addvar(VAR_TALKED_COUNT, 1)
    msgbox("Hello.$")
    release
    end
```

Use `end` or `return` to early-exit out of a script.
```
script MyScript {
    if (flag(FLAG_WON) == true) {
        return
    }
    ...
}
```

A `switch` statement is an easy way to separate different logic for a set of concrete values. Poryscript `switch` statements behave similarly to other languages. However, the cases `break` implicitly. It is not possible to "fall through" to the next case by omitting a `break` at the end of a case, like in C. You *can* use `break` to break out of a case, though--it's just not required. Multiple cases can be designated by listing them immediately after another without a body. Finally, an optional `default` case will take over if none of the provided `case` values are met.  A `switch` statement's comparison value *must always be a `var()` operator*.  Of course, `switch` statements can appear anywhere in the script's logic, such as inside `while` loops, or even other `switch` statements.

```
    switch (var(VAR_NUM_THINGS)) {
        case 0:
            msgbox("You have 0 things.$")
        case 1:
        case 2:
            msgbox("You have 1 or 2 things.$")
        default:
            msgbox("You have at least 3 things.$")
    }
```

Use `text` to include text that's intended to be shared between multiple scripts or in C code. The `text` statement is just a convenient way to write chunks of text, and it exports the text globally, so it is accessible in C code. Currently, there isn't much of a reason to use `text`, but it will be more useful in future updates of Poryscript.
```
script MyScript {
    msgbox(MyText)
}

text MyText {
    "Hello, there.\p"
    "You can refer to me in scripts or C code.$"
}
```

Use `raw` to include raw bytecode script. Anything in a `raw` statement will be directly included into the compiled script. This is useful for defining data or long text.
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
           "on multiple lines.\p$")
    applymovement(EVENT_OBJ_ID_PLAYER, MyScript_Movement)
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

Use single-line comments with `#`. Everything after the `#` will be ignored. Comments cannot be placed in a `raw` statement.
```
# This script does some cool things.
script MyScript {
    # I'm a comment
    ...
}
```

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

This will create a `poryscript` executable binary in the same directory.

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
