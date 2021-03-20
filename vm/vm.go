package vm

import (
	"fmt"
	"github.com/Shea11012/interpreter_in_go/builtin"
	"github.com/Shea11012/interpreter_in_go/code"
	"github.com/Shea11012/interpreter_in_go/compiler"
	"github.com/Shea11012/interpreter_in_go/object"
)

const StackSize = 2048
const GlobalsSize = 65536

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Null{}
)

type VM struct {
	constants []object.Object

	stack []object.Object
	sp    int // 指向栈中下一个值，栈顶stack[sp - 1]

	globals []object.Object

	frames      []*Frame
	framesIndex int
}

const MaxFrames = 1024

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:   bytecode.Constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     make([]object.Object, GlobalsSize),
		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, globals []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = globals
	return vm
}

// StackTop 获取栈顶值
func (v *VM) StackTop() object.Object {
	if v.sp == 0 {
		return nil
	}

	return v.stack[v.sp-1]
}

// Run 运行指令
func (v *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode
	// 从0开始读取
	for v.currentFrame().ip < len(v.currentFrame().Instructions())-1 {
		v.currentFrame().ip++

		ip = v.currentFrame().ip
		ins = v.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpPop:
			v.pop()

		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			v.currentFrame().ip += 2 // 对应每个op可以操作的字节数
			err := v.push(v.constants[constIndex])
			if err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := v.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := v.executeComparison(op)
			if err != nil {
				return err
			}

		case code.OpTrue:
			err := v.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := v.push(False)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := v.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := v.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			v.currentFrame().ip = pos - 1 // 跳过一个指令
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			v.currentFrame().ip += 2 // 跳过两个或1个指令
			condition := v.pop()
			if !isTruthy(condition) {
				v.currentFrame().ip = pos - 1
			}

		case code.OpNull:
			err := v.push(Null)
			if err != nil {
				return err
			}

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:]) // 读取两个字节
			v.currentFrame().ip += 2                   // 索引后移 2 位
			err := v.push(v.globals[globalIndex])
			if err != nil {
				return err
			}

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:]) // 读取两个字节
			v.currentFrame().ip += 2                   // 索引后移 2 位
			v.globals[globalIndex] = v.pop()

		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			v.currentFrame().ip += 1

			frame := v.currentFrame()
			v.stack[frame.basePointer+int(localIndex)] = v.pop()

		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			v.currentFrame().ip += 1

			frame := v.currentFrame()
			err := v.push(v.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}

		case code.OpArray:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			v.currentFrame().ip += 2

			// 根据 OpArray 找出数组元素数量， 根据数量推算出栈中元素，数组的起始索引，构建数组
			array := v.buildArray(v.sp-numElements, v.sp)
			// 修改栈索引
			v.sp = v.sp - numElements

			err := v.push(array)
			if err != nil {
				return err
			}

		case code.OpHash:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			v.currentFrame().ip += 2

			hash, err := v.buildHash(v.sp-numElements, v.sp)
			if err != nil {
				return err
			}
			v.sp = v.sp - numElements

			err = v.push(hash)
			if err != nil {
				return err
			}

		case code.OpIndex:
			index := v.pop()
			left := v.pop()

			err := v.executeIndexExpression(left, index)
			if err != nil {
				return err
			}

		case code.OpClosure:
			constIndex := code.ReadUint16(ins[ip+1:]) // 获取闭包函数索引
			numFree := code.ReadUint8(ins[ip+3:])	// 获取闭包函数所需变量数量
			v.currentFrame().ip += 3

			err := v.pushClosure(int(constIndex),int(numFree))
			if err != nil {
				return err
			}

		case code.OpCurrentClosure:
			currentClosure := v.currentFrame().cl
			err := v.push(currentClosure)
			if err != nil {
				return err
			}

		case code.OpGetBuiltin:
			builtinIndex := code.ReadUint8(ins[ip+1:])
			v.currentFrame().ip += 1

			definition := builtin.BuiltinFns[builtinIndex]
			err := v.push(definition.Builtin)
			if err != nil {
				return err
			}

		case code.OpGetFree:
			freeIndex := code.ReadUint8(ins[ip+1:])
			v.currentFrame().ip += 1

			currentClosure := v.currentFrame().cl
			err := v.push(currentClosure.Free[freeIndex])
			if err != nil {
				return err
			}

		case code.OpCall:
			numArgs := int(code.ReadUint8(ins[ip+1:]))
			v.currentFrame().ip += 1

			err := v.executeCall(numArgs)
			if err != nil {
				return err
			}

		case code.OpReturnValue:
			returnValue := v.pop()
			frame := v.popFrame()
			v.sp = frame.basePointer - 1
			err := v.push(returnValue)
			if err != nil {
				return err
			}

		case code.OpReturn:
			frame := v.popFrame()
			v.sp = frame.basePointer - 1
			err := v.push(Null)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

// LastPoppedStackElem 取栈顶值
func (v *VM) LastPoppedStackElem() object.Object {
	return v.stack[v.sp]
}

// push 将值推入栈中，sp指向栈中下一个值
func (v *VM) push(o object.Object) error {
	if v.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	v.stack[v.sp] = o
	v.sp++
	return nil
}

// pop 弹出栈顶值
func (v *VM) pop() object.Object {
	o := v.stack[v.sp-1]
	v.sp--
	return o
}

func (v *VM) executeBinaryOperation(op code.Opcode) error {
	right := v.pop()
	left := v.pop()
	leftType := left.Type()
	rightType := right.Type()
	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return v.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return v.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unspported types for binary operation: %s %s", leftType, rightType)
	}
}

func (v *VM) executeBinaryIntegerOperation(op code.Opcode, left object.Object, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return v.push(&object.Integer{Value: result})
}

func (v *VM) executeComparison(op code.Opcode) error {
	right := v.pop()
	left := v.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return v.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return v.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return v.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (v *VM) executeIntegerComparison(op code.Opcode, left object.Object, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return v.push(nativeBoolToBooleanObject(leftValue == rightValue))
	case code.OpNotEqual:
		return v.push(nativeBoolToBooleanObject(leftValue != rightValue))
	case code.OpGreaterThan:
		return v.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (v *VM) executeBangOperator() error {
	operand := v.pop()

	switch operand {
	case True:
		return v.push(False)
	case False:
		return v.push(True)
	case Null:
		return v.push(True)
	default:
		return v.push(False)
	}
}

func (v *VM) executeMinusOperator() error {
	operand := v.pop()
	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return v.push(&object.Integer{Value: -value})
}

func (v *VM) executeBinaryStringOperation(op code.Opcode, left object.Object, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unspported string operation: %d", op)
	}
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	return v.push(&object.String{Value: leftValue + rightValue})
}

func (v *VM) buildArray(startIndex int, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = v.stack[i]
	}

	return &object.Array{Elements: elements}
}

func (v *VM) buildHash(startIndex int, endIndex int) (object.Object, error) {
	hashPairs := make(map[object.HashKey]object.HashPair, endIndex-startIndex)
	for i := startIndex; i < endIndex; i += 2 {
		key := v.stack[i]
		value := v.stack[i+1]
		pair := object.HashPair{Key: key, Value: value}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashPairs[hashKey.HashKey()] = pair
	}
	return &object.Hash{Pairs: hashPairs}, nil
}

func (v *VM) executeIndexExpression(left object.Object, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return v.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return v.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supporated: %s", left.Type())
	}
}

func (v *VM) executeArrayIndex(array object.Object, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > max {
		return v.push(Null)
	}

	return v.push(arrayObject.Elements[i])
}

func (v *VM) executeHashIndex(hash object.Object, index object.Object) error {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return v.push(Null)
	}

	return v.push(pair.Value)
}

func (v *VM) currentFrame() *Frame {
	return v.frames[v.framesIndex-1]
}

func (v *VM) pushFrame(f *Frame) {
	v.frames[v.framesIndex] = f
	v.framesIndex++
}

func (v *VM) popFrame() *Frame {
	v.framesIndex--
	return v.frames[v.framesIndex]
}

func (v *VM) executeCall(args int) error {
	callee := v.stack[v.sp-1-args]

	switch callee := callee.(type) {
	case *object.Closure:
		return v.callClosure(callee,args)
	case *object.Builtin:
		return v.callBuiltin(callee,args)
	default:
		return fmt.Errorf("calling non-function and non-builtin-in")
	}
}

func (v *VM) callClosure(cl *object.Closure,numArgs int) error {
	if numArgs != cl.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",cl.Fn.NumParameters,numArgs)
	}

	frame := NewFrame(cl, v.sp-numArgs)
	v.pushFrame(frame)
	// 跳过函数地址和函数变量地址
	v.sp = frame.basePointer + cl.Fn.NumLocals

	return nil
}

func (v *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
	args := v.stack[v.sp-numArgs:v.sp]
	result := builtin.Fn(args...)
	v.sp = v.sp - numArgs - 1

	var err error
	if result != nil {
		err = v.push(result)
	} else {
		err = v.push(Null)
	}

	return err
}

func (v *VM) pushClosure(constIndex int,numFree int) error {
	constant := v.constants[constIndex]	// 获取闭包函数
	function,ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v",constant)
	}

	free := make([]object.Object,numFree)
	for i := 0; i < numFree; i++ {
		free[i] = v.stack[v.sp-numFree+i]	// sp 永远指向栈顶，所以减去 numFree 就是自由变量所处的位置
	}

	v.sp -= numFree

	closure := &object.Closure{Fn: function,Free: free}

	return v.push(closure)
}

func nativeBoolToBooleanObject(b bool) *object.Boolean {
	if b {
		return True
	}
	return False
}
