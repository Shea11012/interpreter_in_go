package vm

import (
	"fmt"
	"github.com/Shea11012/interpreter_in_go/ast"
	"github.com/Shea11012/interpreter_in_go/compiler"
	"github.com/Shea11012/interpreter_in_go/lexer"
	"github.com/Shea11012/interpreter_in_go/object"
	"github.com/Shea11012/interpreter_in_go/parser"
	"strconv"
	"testing"
)

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not integer got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value got=%d,want=%d", result.Value, expected)
	}

	return nil
}

type vmTestCase struct {
	input    string
	expected interface{}
	name     string
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {

			program := parse(tt.input)
			comp := compiler.New()
			err := comp.Compile(program)
			if err != nil {
				t.Fatalf("compiler error: %s", err)
			}

			for i,constant := range comp.Bytecode().Constants {
				fmt.Printf("constant %d %p (%T):\n",i,constant,constant)

				switch constant := constant.(type) {
				case *object.CompiledFunction:
					fmt.Printf(" Instructions:\n%s",constant.Instructions)
				case *object.Integer:
					fmt.Printf(" Value: %d\n",constant.Value)
				}
			}

			vm := New(comp.Bytecode())
			err = vm.Run()
			if err != nil {
				t.Fatalf("vm error: %s", err)
			}

			stackElem := vm.LastPoppedStackElem()

			testExpectedObject(t, tt.expected, stackElem)
		})
	}

}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case string:
		err := testStringObject(expected, actual)
		if err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(expected, actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	case *object.Null:
		if actual != Null {
			t.Errorf("object is not null: %T (%+v)", actual, actual)
		}

	case *object.Error:
		obj, ok := actual.(*object.Error)
		if !ok {
			t.Errorf("object is not Error: %T (%+v)", actual, actual)
			return
		}

		if obj.Message != expected.Message {
			t.Errorf("wrong error message. expected=%q, got=%q", expected.Message, obj.Message)
		}

	case []int:
		array, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("object not array: %T (%+v)", actual, actual)
			return
		}

		if len(array.Elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d,got=%d", len(expected), len(array.Elements))
			return
		}

		for i, expectedElem := range expected {
			err := testIntegerObject(int64(expectedElem), array.Elements[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}

	case map[object.HashKey]int64:
		hash, ok := actual.(*object.Hash)
		if !ok {
			t.Errorf("object is not Hash. got=%T (%+v)", actual, actual)
			return
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("hash has wrong number of Pairs. want=%d,got=%d", len(expected), len(hash.Pairs))
			return
		}

		for expectedKey, expectedValue := range expected {
			pair, ok := hash.Pairs[expectedKey]
			if !ok {
				t.Errorf("no pair for given key in pairs")
			}

			err := testIntegerObject(expectedValue, pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	}
}

func testStringObject(expected string, obj object.Object) error {
	result, ok := obj.(*object.String)
	if !ok {
		return fmt.Errorf("object is not string. got=%T (%+v)", obj, obj)
	}

	if result.Value != expected {
		return fmt.Errorf("obj has rong value. got=%s,want=%s", result.Value, expected)
	}

	return nil
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not boolean got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value got=%t,want=%t", result.Value, expected)
	}

	return nil
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{input: "1", expected: 1},
		{input: "2", expected: 2},
		{input: "1 + 2", expected: 3},
		{input: "1-2", expected: -1},
		{input: "1*2", expected: 2},
		{input: "4/2", expected: 2},
		{input: "50/2*2+10-5", expected: 55},
		{input: "5+5+5+5-10", expected: 10},
		{input: "2*2*2*2*2", expected: 32},
		{input: "5*2+10", expected: 20},
		{input: "5+2*10", expected: 25},
		{input: "5*(2+10)", expected: 60},
		{input: "-5", expected: -5},
		{input: "-10", expected: -10},
		{input: "-50 + 100 + -50", expected: 0},
		{input: "(5 + 10 * 2 + 15 / 3)", expected: 30},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{input: "true", expected: true},
		{input: "false", expected: false},
		{input: "1<2", expected: true},
		{input: "1>2", expected: false},
		{input: "1<1", expected: false},
		{input: "1>1", expected: false},
		{input: "1 == 1", expected: true},
		{input: "1 != 1", expected: false},
		{input: "1 == 2", expected: false},
		{input: "1 != 2", expected: true},
		{input: "true == true", expected: true},
		{input: "false == false", expected: true},
		{input: "true == false", expected: false},
		{input: "true != false", expected: true},
		{input: "false != true", expected: true},
		{input: "(1<2) == true", expected: true},
		{input: "(1<2) == false", expected: false},
		{input: "(1>2) == true", expected: false},
		{input: "(1>2) == false", expected: true},
		{input: "!true", expected: false},
		{input: "!false", expected: true},
		{input: "!5", expected: false},
		{input: "!!true", expected: true},
		{input: "!!false", expected: false},
		{input: "!!5", expected: true},
		{input: "!(if(false){5;})", expected: true},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{input: "if(false){10}else{20}", expected: 20},
		{input: "if(true){10}", expected: 10},
		{input: "if(true){10} else {20}", expected: 10},
		{input: "if(1){10}", expected: 10},
		{input: "if(1 < 2){10}", expected: 10},
		{input: "if(1 < 2){10} else {20}", expected: 10},
		{input: "if(1 > 2){10} else {20}", expected: 20},
		{input: "if(1 > 2){10}", expected: Null},
		{input: "if(false){10}", expected: Null},
		{input: "if((if(false){10})){10} else {20}", expected: 20},
	}

	runVmTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{input: "let one=1;one", expected: 1},
		{input: "let one=1; let two=2; one+two", expected: 3},
		{input: "let one=1; let two=one + one; one+two", expected: 3},
	}

	runVmTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{input: `"monkey"`, expected: "monkey"},
		{input: `"mon" + "key"`, expected: "monkey"},
		{input: `"mon" + "key" + "banana"`, expected: "monkeybanana"},
	}

	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{input: "[]", expected: []int{}},
		{input: "[1,2,3]", expected: []int{1, 2, 3}},
		{input: "[1+2,3*4,5+6]", expected: []int{3, 12, 11}},
	}

	runVmTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "{}",
			expected: map[object.HashKey]int64{},
		},
		{
			input: "{1:2,2:3}",
			expected: map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 2,
				(&object.Integer{Value: 2}).HashKey(): 3,
			},
		},
		{
			input: "{1+1:2*2,3+3:4*4}",
			expected: map[object.HashKey]int64{
				(&object.Integer{Value: 2}).HashKey(): 4,
				(&object.Integer{Value: 6}).HashKey(): 16,
			},
		},
	}

	runVmTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{input: "[1,2,3][1]", expected: 2},
		{input: "[1,2,3][0+2]", expected: 3},
		{input: "[[1,1,1]][0][0]", expected: 1},
		{input: "[][0]", expected: Null},
		{input: "[1,2,3][99]", expected: Null},
		{input: "[1][-1]", expected: Null},
		{input: "{1:1,2:2}[1]", expected: 1},
		{input: "{1:1,2:2}[2]", expected: 2},
		{input: "{1:1}[0]", expected: Null},
		{input: "{}[0]", expected: Null},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let fivePlusTen = fn() { 5 + 10; };
			fivePlusTen();
`,
			expected: 15,
		},
		{
			input: `
			let one = fn(){1;};
			let two = fn(){2;};
			one() + two();
`,
			expected: 3,
		},
		{
			input: `
			let a = fn(){1};
			let b = fn(){a() + 1};
			let c = fn(){b() + 1};
			c();
`,
			expected: 3,
		},
	}

	runVmTests(t, tests)
}

func TestFunctionsWithReturnStatement(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let earlyExit = fn() { return 99; 100;};
			earlyExit();
`,
			expected: 99,
		},
		{
			input: `
			let earlyExit = fn() { return 99; return 100;};
			earlyExit();
`,
			expected: 99,
		},
	}

	runVmTests(t, tests)
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let noReturn = fn(){};
			noReturn();
`,
			expected: Null,
		},
		{
			input: `
			let noReturn = fn(){};
			let noReturnTwo = fn(){ noReturn(); };
			noReturn();
			noReturnTwo();
`,
			expected: Null,
		},
	}

	runVmTests(t, tests)
}

func TestFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let returnsOne = fn(){1;};
			let returnsOneReturner = fn(){ returnsOne;};
			returnsOneReturner()();
`,
			expected: 1,
		},
		{
			input: `
			let returnsOneReturner = fn(){ 
				let returnsOne = fn(){1;};
				returnsOne;
			};
			returnsOneReturner()();
`,
			expected: 1,
		},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let one = fn() { let one = 1;one};
			one();
`,
			expected: 1,
		},
		{
			input: `
			let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
			oneAndTwo();
`,
			expected: 3,
		},
		{
			input: `
				let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
				let threeAndFour = fn() { let three = 3; let four = 4; three + four; };
				oneAndTwo() + threeAndFour();
`,
			expected: 10,
		},
		{
			input: `
			let firstFoobar = fn() { let foobar = 50; foobar;};
			let secondFoobar = fn() { let foobar = 100; foobar;};
			firstFoobar() + secondFoobar();
`,
			expected: 150,
		},
		{
			input: `
			let globalSeed = 50;
			let minusOne = fn() {
				let num = 1;
				globalSeed - num;
			}
			let minusTwo = fn() {
				let num = 2;
				globalSeed - num;
			}
			minusOne() + minusTwo();
`,
			expected: 97,
			name:     "5",
		},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithArgumentsAndBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let identify = fn(a) {a;};
			identify(4);
`,
			expected: 4,
		},
		{
			input: `
			let sum = fn(a,b) {	a + b;};
			sum(1,2);
`,
			expected: 3,
		},
		{
			input: `
			let sum = fn(a,b) {	let c = a + b;c;};
			sum(1,2);
`,
			expected: 3,
		},
		{
			input: `
			let sum = fn(a,b) {	let c = a + b;c;};
			sum(1,2) + sum(3,4);
`,
			expected: 10,
		},
		{
			input: `
			let sum = fn(a,b) {let c = a + b;c;};
			let outer = fn() { sum(1,2) + sum(3,4); };
			outer();
`,
			expected: 10,
		},
		{
			input: `
			let globalNum = 10;
			let sum = fn(a,b) {
				let c = a +b;
				c + globalNum;
			};

			let outer = fn() {
				sum(1,2) + sum(3,4) + globalNum;
			};
			outer() + globalNum;
`,
			expected: 50,
		},
	}
	runVmTests(t, tests)
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			fn(){1;}(1);
`,
			expected: `wrong number of arguments: want=0, got=1`,
		},
		{
			input: `
			fn(a){a;}();
`,
			expected: `wrong number of arguments: want=1, got=0`,
		},
		{
			input: `
			fn(a,b){a+b;}(1);
`,
			expected: `wrong number of arguments: want=2, got=1`,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			program := parse(tt.input)
			comp := compiler.New()
			err := comp.Compile(program)
			if err != nil {
				t.Fatalf("compiler error: %s", err)
			}

			vm := New(comp.Bytecode())
			err = vm.Run()
			if err == nil {
				t.Fatalf("expected VM error but resulted in none")
			}

			if err.Error() != tt.expected {
				t.Fatalf("wrong VM erroor: want=%q, got=%q", tt.expected, err)
			}
		})
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		/*{
			input:    `map([1,2,3],fn(a) { a * 2});`,
			expected: []int{2, 4, 6},
		},*/
		{
			input:    `len("")`,
			expected: 0,
		},
		{
			input:    `len("four")`,
			expected: 4,
		},
		{
			input:    `len("hello world")`,
			expected: 11,
		},
		{
			input: `len(1)`,
			expected: &object.Error{
				Message: "argument to `len` not supported, got INTEGER",
			},
		},
		{
			input:    `len("one","two")`,
			expected: &object.Error{Message: "wrong number of arguments. got=2, want=1"},
		},
		{
			input:    `len([1,2,3])`,
			expected: 3,
		},
		{
			input:    `len([])`,
			expected: 0,
		},
		{
			input:    `puts("hello","world!")`,
			expected: Null,
		},
		{
			input:    `first([1,2,3])`,
			expected: 1,
		},
		{
			input:    `first([])`,
			expected: Null,
		},
		{
			input:    `first(1)`,
			expected: &object.Error{Message: "argument to `first` must be ARRAY, got INTEGER"},
		},
		{
			input:    `last([1,2,3])`,
			expected: 3,
		},
		{
			input:    `last([])`,
			expected: Null,
		},
		{
			input:    `last(1)`,
			expected: &object.Error{Message: "argument to `last` must be ARRAY, got INTEGER"},
		},
		{
			input:    `rest([1,2,3])`,
			expected: []int{2, 3},
		},
		{
			input:    `rest([])`,
			expected: Null,
		},
		{
			input:    `push([],1)`,
			expected: []int{1},
		},
		{
			input:    `push(1,1)`,
			expected: &object.Error{Message: "argument to `push` must be ARRAY, got INTEGER"},
		},
	}

	runVmTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let newClosure = fn(a) {
				fn() { a; };
			};
			let closure = newClosure(99);
			closure();
`,
			expected: 99,
		},
		{
			input: `
			let newAdder = fn(a,b) {
				fn(c) { a + b + c };
			}
			let adder = newAdder(1,2);
			adder(8);
`,
			expected: 11,
		},
		{
			input: `
			let newAdder = fn(a,b) {
				let c = a + b;
				fn(d) { c + d };
			}
			let adder = newAdder(1,2);
			adder(8);
`,
			expected: 11,
		},
		{
			input: `
			let newAdderOuter = fn(a,b) {
				let c = a + b;
				fn(d) { 
					let e = c + d;
					fn(f) { e+f; };
				};
			}
			let adderInner = newAdderOuter(1,2);
			let adder = adderInner(3);
			adder(8);
`,
			expected: 14,
		},
		{
			input: `
			let a = 1;
			let newAdderOuter = fn(b) {
				fn(c) {
					fn(d) { a + b + c + d };
				};
			};
			let newAdderInner = newAdderOuter(2);
			let adder = newAdderInner(3);
			adder(8);
`,
			expected: 14,
		},
		{
			input: `
			let newClosure = fn(a,b) {
				let one = fn() { a; };
				let two = fn() { b; };
				fn() { one() + two(); };
			};
			let closure = newClosure(9,90);
			closure();
`,
			expected: 99,
		},
	}

	runVmTests(t, tests)
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let countDown = fn(x) {
				if (x == 0) {
					return 0;
				} else {
					countDown(x - 1);
				}
			};
			countDown(1);
`,
			expected: 0,
		},
		{
			input: `
			let countDown = fn(x) {
				if (x == 0) {
					return 0;
				} else {
					countDown(x - 1);
				}
			};
			let wrapper = fn() {
				countDown(1);
			};
			wrapper();
`,
			expected: 0,
		},
		{
			input: `
			let wrapper = fn() {
				let countDown = fn(x) {
					if (x == 0) {
						return 0;
					} else {
						countDown(x-1);
					}
				};
				countDown(1);
			};
			wrapper();
`,
			expected: 0,
		},
	}

	runVmTests(t, tests)
}
