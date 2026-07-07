package runtime

type Op uint8

const (
	OpText Op = iota
	OpConst
	OpNil
	OpTrue
	OpFalse
	OpLoadVar
	OpLoadLocal
	OpStoreLocal
	OpAttr
	OpIndex
	OpCall
	OpCallMethod
	OpMakeList
	OpMakeMap
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpPow
	OpNeg
	OpConcat
	OpEq
	OpNeq
	OpLt
	OpGt
	OpLte
	OpGte
	OpIn
	OpNotIn
	OpNot
	OpEcho
	OpPop
	OpJump
	OpJumpFalse
	OpJumpIfFalseOrPop
	OpJumpIfTrueOrPop
	OpIterInit
	OpIterNext
	OpRenderBlock
	OpInclude
)

type Instr struct {
	Op  Op
	Arg int
}

type LoopSpec struct {
	ValSlot    int
	KeySlot    int
	LoopSlot   int
	ElseTarget int
	EndTarget  int
}

type Callable func(args []Value) (Value, error)

type Program struct {
	Name       string
	Instrs     []Instr
	Consts     []Value
	Callables  []Callable
	Loops      []LoopSpec
	NumLocals  int
	Blocks     map[string]*Program
	Parent     string
	Autoescape bool
}

type Loader interface {
	Load(name string) (*Program, bool)
}

func packCall(index, argc int) int { return index<<8 | argc&0xff }

func unpackCall(arg int) (index, argc int) { return arg >> 8, arg & 0xff }
