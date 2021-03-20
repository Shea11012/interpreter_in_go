package builtin

import (
	"fmt"
	"github.com/Shea11012/interpreter_in_go/object"
)

type BuiltinFn struct {
	Name    string
	Builtin *object.Builtin
}

var BuiltinFns []BuiltinFn

type method func() BuiltinFn

func init() {
	m := []method{
		lenMethod,
		putsMethod,
		firstMethod,
		lastMethod,
		restMethod,
		pushMethod,
		mapMethod,
	}
	registerMethod(m...)
}

func registerMethod(methods ...method) {
	for _, m := range methods {
		BuiltinFns = append(BuiltinFns, m())
	}
}

func GetBuiltinByName(name string) *object.Builtin {
	for _, def := range BuiltinFns {
		if def.Name == name {
			return def.Builtin
		}
	}

	return nil
}

func lenMethod() BuiltinFn {
	return BuiltinFn{
		Name: "len",
		Builtin: &object.Builtin{Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			default:
				return newError("argument to `len` not supported, got %s", arg.Type())
			}
		}},
	}
}

func putsMethod() BuiltinFn {
	return BuiltinFn{
		Name: "puts",
		Builtin: &object.Builtin{Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return nil
		}},
	}
}

func firstMethod() BuiltinFn {
	return BuiltinFn{
		Name: "first",
		Builtin: &object.Builtin{Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return nil
		}},
	}
}

func lastMethod() BuiltinFn {
	return BuiltinFn{
		Name: "last",
		Builtin: &object.Builtin{Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}

			return nil
		}},
	}
}

func restMethod() BuiltinFn {
	return BuiltinFn{
		Name: "rest",
		Builtin: &object.Builtin{Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `rest` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]object.Object, length-1, length-1)
				copy(newElements, arr.Elements[1:length])

				return &object.Array{Elements: newElements}
			}

			return nil
		}},
	}
}

func pushMethod() BuiltinFn {
	return BuiltinFn{
		Name: "push",
		Builtin: &object.Builtin{Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			newElements := make([]object.Object, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]

			return &object.Array{Elements: newElements}
		}},
	}
}

// todo
func mapMethod() BuiltinFn {
	return BuiltinFn{
		Name: "map",
		Builtin: &object.Builtin{Fn: func(args ...object.Object) object.Object {
			fmt.Println(len(args))
			fmt.Printf("%+v\n",args[0])
			fmt.Printf("%+v\n",args[1])
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `map` must be array, got %s", args[0].Type())
			}

			if args[1].Type() != object.COMPILE_FUNCTION_OBJ {
				return newError("second argument to `map` must be compiled function")
			}

			/*arr := args[0].(*object.Array)
			fmt.Printf("%T %+v",arr,arr)
			fn := args[1].(*object.CompiledFunction)
			fmt.Printf("%T %+v",fn,fn)*/
			/*fnLiteral := &ast.FunctionLiteral{
				Token: token.Token{
					Type: token.FUNCTION,
				},
				Parameters: fn.Parameters,
				Body:       fn.Body,
			}

			callExpression := &ast.CallExpression{
				Token:     token.Token{},
				Function:  fnLiteral,
				Arguments: nil,
			}

			comp := compiler.New()

			newElements := make([]object.Object, 0, len(arr.Elements))
			for _, el := range arr.Elements {
				integer := el.(*object.Integer)
				arguments := []ast.Expression{
					&ast.IntegerLiteral{
						Token: token.Token{
							Type: token.INT,
						},
						Value: integer.Value,
					},
				}
				callExpression.Arguments = arguments

				expressionStmt := &ast.ExpressionStatement{
					Token:      token.Token{},
					Expression: callExpression,
				}

				program := &ast.Program{Statements: []ast.Statement{expressionStmt}}
				err := comp.Compile(program)
				if err != nil {
					return newError("map compile err: %s", err)
				}
				v := vm.New(comp.Bytecode())
				err = v.Run()
				if err != nil {
					return newError("map run err: %s", err)
				}

				obj := v.LastPoppedStackElem()
				comp.Reset()
				newElements = append(newElements, obj)
			}*/

			// return &object.Array{Elements: newElements}
			return nil
		}},
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}
