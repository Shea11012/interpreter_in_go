package evaluator

import (
	"fmt"
	"github.com/Shea11012/interpreter_in_go/object"
)

var builtins = make(map[string]*object.Builtin)

func init() {
	builtins["len"] = &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError("wrong number of arguments got=%d,want=1", len(args))
		}

		switch arg := args[0].(type) {
		case *object.String:
			return &object.Integer{Value: int64(len(arg.Value))}
		case *object.Array:
			return &object.Integer{Value: int64(len(arg.Elements))}
		default:
			return newError("argument to `len` not supported, got %s", arg.Type())
		}
	}}

	builtins["first"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments got=%d,want=1", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `first` must be array,got=%s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return NULL
		},
	}

	builtins["last"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments got=%d,want=1", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `last` must be array,got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}

			return NULL
		},
	}

	builtins["rest"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments got=%d,want=1", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `rest` must be array,got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]object.Object, length-1, length-1)
				copy(newElements, arr.Elements[1:length])

				return &object.Array{Elements: newElements}
			}

			return NULL
		},
	}

	builtins["push"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments got=%d,want=2", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be array,got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			newElements := make([]object.Object, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]

			return &object.Array{Elements: newElements}
		},
	}

	builtins["map"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments got=%d,want=2", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `map` must be array,got %s", args[0].Type())
			}

			if args[1].Type() != object.FUNCTION_OBJ {
				return newError("second argument to `map` must be function")
			}

			arr := args[0].(*object.Array)
			fn := args[1].(*object.Function)

			newElements := make([]object.Object, 0, len(arr.Elements))
			for _, el := range arr.Elements {
				params := []object.Object{el}
				newElements = append(newElements, applyFunction(fn, params))
			}

			return &object.Array{Elements: newElements}
		},
	}

	builtins["reduce"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments got=%d,want=3", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("first argument to `reduce` must be array,got %s", args[0].Type())
			}

			if args[1].Type() != object.INTEGER_OBJ {
				return newError("second argument to `reduce` must be integer,got %s", args[1].Type())
			}

			if args[2].Type() != object.FUNCTION_OBJ {
				return newError("third argument to `reduce` must be function,got %s", args[2].Type())
			}

			arr := args[0].(*object.Array)
			initial := args[1]
			fn := args[2].(*object.Function)
			var result = initial

			for _, el := range arr.Elements {
				params := []object.Object{result, el}
				result = applyFunction(fn, params)
			}

			resultInt := result.(*object.Integer)
			return resultInt
		},
	}

	builtins["puts"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return NULL
		},
	}
}
