package runtime

import (
	"fmt"
	"strings"
)

func (op Op) String() string {
	switch op {
	case OpText:
		return "TEXT"
	case OpConst:
		return "CONST"
	case OpNil:
		return "NIL"
	case OpTrue:
		return "TRUE"
	case OpFalse:
		return "FALSE"
	case OpLoadVar:
		return "LOAD_VAR"
	case OpLoadLocal:
		return "LOAD_LOCAL"
	case OpStoreLocal:
		return "STORE_LOCAL"
	case OpAttr:
		return "ATTR"
	case OpIndex:
		return "INDEX"
	case OpCall:
		return "CALL"
	case OpCallMethod:
		return "CALL_METHOD"
	case OpMakeList:
		return "MAKE_LIST"
	case OpMakeMap:
		return "MAKE_MAP"
	case OpAdd:
		return "ADD"
	case OpSub:
		return "SUB"
	case OpMul:
		return "MUL"
	case OpDiv:
		return "DIV"
	case OpMod:
		return "MOD"
	case OpPow:
		return "POW"
	case OpNeg:
		return "NEG"
	case OpConcat:
		return "CONCAT"
	case OpEq:
		return "EQ"
	case OpNeq:
		return "NEQ"
	case OpLt:
		return "LT"
	case OpGt:
		return "GT"
	case OpLte:
		return "LTE"
	case OpGte:
		return "GTE"
	case OpIn:
		return "IN"
	case OpNotIn:
		return "NOT_IN"
	case OpNot:
		return "NOT"
	case OpEcho:
		return "ECHO"
	case OpPop:
		return "POP"
	case OpJump:
		return "JUMP"
	case OpJumpFalse:
		return "JUMP_FALSE"
	case OpJumpIfFalseOrPop:
		return "JUMP_IF_FALSE_OR_POP"
	case OpJumpIfTrueOrPop:
		return "JUMP_IF_TRUE_OR_POP"
	case OpIterInit:
		return "ITER_INIT"
	case OpIterNext:
		return "ITER_NEXT"
	case OpRenderBlock:
		return "RENDER_BLOCK"
	case OpInclude:
		return "INCLUDE"
	}
	return fmt.Sprintf("OP(%d)", int(op))
}

func Disassemble(p *Program) string {
	var b strings.Builder
	fmt.Fprintf(&b, "program %q (locals=%d)\n", p.Name, p.NumLocals)
	for i, in := range p.Instrs {
		fmt.Fprintf(&b, "%4d  %-20s", i, in.Op)
		switch in.Op {
		case OpText, OpConst, OpLoadVar, OpAttr:
			fmt.Fprintf(&b, " %d ; %q", in.Arg, p.Consts[in.Arg].String())
		case OpCall, OpCallMethod:
			idx, argc := unpackCall(in.Arg)
			fmt.Fprintf(&b, " idx=%d argc=%d", idx, argc)
		case OpJump, OpJumpFalse, OpJumpIfFalseOrPop, OpJumpIfTrueOrPop:
			fmt.Fprintf(&b, " -> %d", in.Arg)
		default:
			if in.Arg != 0 {
				fmt.Fprintf(&b, " %d", in.Arg)
			}
		}
		b.WriteByte('\n')
	}
	for name, blk := range p.Blocks {
		fmt.Fprintf(&b, "\nblock %q:\n%s", name, Disassemble(blk))
	}
	return b.String()
}
