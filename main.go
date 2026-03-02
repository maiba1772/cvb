package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"cvb-lang/evaluator"
	"cvb-lang/lexer"
	"cvb-lang/object"
	"cvb-lang/parser"
)

const PROMPT = "cvb> "

func main() {
	if len(os.Args) > 1 {
		runFile(os.Args[1])
	} else {
		startREPL()
	}
}

func runFile(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "Parse error: %s\n", err)
		}
		os.Exit(1)
	}

	env := object.NewEnvironment()
	result := evaluator.Eval(program, env)

	if result != nil && result.Type() == object.ERROR_OBJ {
		fmt.Fprintf(os.Stderr, "Runtime error: %s\n", result.Inspect())
		os.Exit(1)
	}
}

func startREPL() {
	reader := bufio.NewReader(os.Stdin)
	env := object.NewEnvironment()

	fmt.Println("CVB Language REPL - Type 'exit' to quit")
	fmt.Print(PROMPT)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "exit" || line == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		if line == "" {
			fmt.Print(PROMPT)
			continue
		}

		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			for _, err := range p.Errors() {
				fmt.Fprintf(os.Stderr, "Parse error: %s\n", err)
			}
			fmt.Print(PROMPT)
			continue
		}

		result := evaluator.Eval(program, env)

		if result != nil {
			if result.Type() == object.ERROR_OBJ {
				fmt.Fprintf(os.Stderr, "Runtime error: %s\n", result.Inspect())
			} else if result.Type() != object.NULL_OBJ {
				fmt.Println(result.Inspect())
			}
		}

		fmt.Print(PROMPT)
	}
}
