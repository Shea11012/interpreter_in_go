package vm

import (
	"github.com/Shea11012/interpreter_in_go/code"
	"github.com/Shea11012/interpreter_in_go/object"
)

type Frame struct {
	cl          *object.Closure
	ip          int
	basePointer int	// 指向当前 frame 的栈底
}

func NewFrame(cl *object.Closure,basePointer int) *Frame {
	return &Frame{cl: cl, ip: -1,basePointer: basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
