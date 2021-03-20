package repl

import (
	"bufio"
	"fmt"
	"github.com/Shea11012/interpreter_in_go/builtin"
	"github.com/Shea11012/interpreter_in_go/compiler"
	"github.com/Shea11012/interpreter_in_go/lexer"
	"github.com/Shea11012/interpreter_in_go/object"
	"github.com/Shea11012/interpreter_in_go/parser"
	"github.com/Shea11012/interpreter_in_go/vm"
	"io"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	constants := []object.Object{}
	globals := make([]object.Object,vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()

	for i, v := range builtin.BuiltinFns {
		symbolTable.DefineBuiltin(i,v.Name)
	}

	for {
		_, _ = fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		comp := compiler.NewWithState(symbolTable,constants)
		err := comp.Compile(program)
		if err != nil {
			_, _ = fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		code := comp.Bytecode()
		constants = code.Constants
		machine := vm.NewWithGlobalsStore(code,globals)
		err = machine.Run()
		if err != nil {
			_, _ = fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		stackTop := machine.LastPoppedStackElem()
		_, _ = io.WriteString(out, stackTop.Inspect())
		_, _ = io.WriteString(out, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	_, _ = io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		_, _ = io.WriteString(out, "\t"+msg+"\n")
	}
}
