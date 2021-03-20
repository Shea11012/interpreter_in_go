package evaluator

import (
	"github.com/Shea11012/interpreter_in_go/builtin"
	"github.com/Shea11012/interpreter_in_go/object"
)

var builtins = make(map[string]*object.Builtin)

func init() {
	builtins["len"] = builtin.GetBuiltinByName("len")
	builtins["first"] = builtin.GetBuiltinByName("first")
	builtins["last"] = builtin.GetBuiltinByName("last")
	builtins["rest"] = builtin.GetBuiltinByName("rest")
	builtins["push"] = builtin.GetBuiltinByName("push")
	builtins["puts"] = builtin.GetBuiltinByName("puts")

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
}
