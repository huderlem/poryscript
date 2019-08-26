package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/huderlem/poryscript/emitter"
	"github.com/huderlem/poryscript/lexer"
	"github.com/huderlem/poryscript/parser"
)

type options struct {
	inputFilepath  string
	outputFilepath string
}

func parseOptions() options {
	helpPtr := flag.Bool("h", false, "show poryscript help information")
	inputPtr := flag.String("i", "", "input poryscript file (leave empty to read from standard input)")
	outputPtr := flag.String("o", "", "output script file (leave empty to write to standard output)")
	flag.Parse()

	if *helpPtr == true {
		flag.Usage()
		os.Exit(0)
	}

	return options{
		inputFilepath:  *inputPtr,
		outputFilepath: *outputPtr,
	}
}

func getInput(filepath string) string {
	var bytes []byte
	var err error
	if filepath == "" {
		bytes, err = ioutil.ReadAll(os.Stdin)
	} else {
		bytes, err = ioutil.ReadFile(filepath)
	}

	if err != nil {
		panic(fmt.Sprintf("Error reading poryscript input: %s", err.Error()))
	}
	return string(bytes)
}

func writeOutput(output string, filepath string) {
	if filepath == "" {
		fmt.Print(output)
	} else {
		f, err := os.Create(filepath)
		if err != nil {
			panic(err)
		}

		_, err = io.WriteString(f, output)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	options := parseOptions()
	input := getInput(options.inputFilepath)

	parser := parser.New(lexer.New(input))
	program := parser.ParseProgram()
	if program == nil {
		os.Exit(1)
	}

	emitter := emitter.New(program)
	result := emitter.Emit()
	writeOutput(result, options.outputFilepath)
}
