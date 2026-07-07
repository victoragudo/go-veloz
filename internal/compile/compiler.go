package compile

import "github.com/victoragudo/go-veloz/internal/runtime"

type Resolver interface {
	ResolveCallable(name string, filter bool) (runtime.Callable, bool)
}

type scope struct {
	vars map[string]int
}

type compiler struct {
	res        Resolver
	autoescape bool

	instrs      []runtime.Instr
	consts      []runtime.Value
	constIdx    map[string]int
	callables   []runtime.Callable
	callableIdx map[string]int
	loops       []runtime.LoopSpec
	blocks      map[string]*runtime.Program

	scopes    []scope
	nextSlot  int
	numLocals int

	err error
}

func Compile(tmpl *Template, res Resolver, autoescape bool) (*runtime.Program, error) {
	blocks := map[string]*runtime.Program{}
	c := newCompiler(res, autoescape, blocks)
	c.enterScope()
	for _, n := range tmpl.Nodes {
		c.node(n)
		if c.err != nil {
			return nil, c.err
		}
	}
	return &runtime.Program{
		Instrs:     c.instrs,
		Consts:     c.consts,
		Callables:  c.callables,
		Loops:      c.loops,
		NumLocals:  c.numLocals,
		Blocks:     blocks,
		Parent:     tmpl.Parent,
		Autoescape: autoescape,
	}, nil
}

func newCompiler(res Resolver, autoescape bool, blocks map[string]*runtime.Program) *compiler {
	return &compiler{
		res:         res,
		autoescape:  autoescape,
		constIdx:    map[string]int{},
		callableIdx: map[string]int{},
		blocks:      blocks,
	}
}

func (c *compiler) enterScope() {
	c.scopes = append(c.scopes, scope{vars: map[string]int{}})
}

func (c *compiler) exitScope() {
	top := c.scopes[len(c.scopes)-1]
	c.nextSlot -= len(top.vars)
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *compiler) declareLocal(name string) int {
	slot := c.nextSlot
	c.nextSlot++
	if c.nextSlot > c.numLocals {
		c.numLocals = c.nextSlot
	}
	c.scopes[len(c.scopes)-1].vars[name] = slot
	return slot
}

func (c *compiler) resolveLocal(name string) (int, bool) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if slot, ok := c.scopes[i].vars[name]; ok {
			return slot, true
		}
	}
	return 0, false
}

func (c *compiler) emit(op runtime.Op, arg int) int {
	c.instrs = append(c.instrs, runtime.Instr{Op: op, Arg: arg})
	return len(c.instrs) - 1
}

func (c *compiler) patch(i, target int) {
	c.instrs[i].Arg = target
}

func (c *compiler) here() int { return len(c.instrs) }

func (c *compiler) addConst(v runtime.Value) int {
	key := constKey(v)
	if key != "" {
		if idx, ok := c.constIdx[key]; ok {
			return idx
		}
	}
	idx := len(c.consts)
	c.consts = append(c.consts, v)
	if key != "" {
		c.constIdx[key] = idx
	}
	return idx
}

func constKey(v runtime.Value) string {
	switch v.Kind() {
	case runtime.KindString:
		return "s:" + v.String()
	case runtime.KindInt:
		return "i:" + v.String()
	case runtime.KindFloat:
		return "f:" + v.String()
	case runtime.KindBool:
		return "b:" + v.String()
	case runtime.KindNil:
		return "n:"
	}
	return ""
}

func (c *compiler) resolveCallable(name string, filter bool) (int, bool) {
	prefix := "fn:"
	if filter {
		prefix = "ft:"
	}
	key := prefix + name
	if idx, ok := c.callableIdx[key]; ok {
		return idx, true
	}
	fn, ok := c.res.ResolveCallable(name, filter)
	if !ok {
		return 0, false
	}
	idx := len(c.callables)
	c.callables = append(c.callables, fn)
	c.callableIdx[key] = idx
	return idx, true
}

func (c *compiler) fail(format string, args ...any) {
	if c.err == nil {
		c.err = &CompileError{Message: sprintf(format, args...)}
	}
}

func (c *compiler) node(n Node) {
	switch v := n.(type) {
	case *TextNode:
		if v.Text == "" {
			return
		}
		c.emit(runtime.OpText, c.addConst(runtime.Str(v.Text)))
	case *PrintNode:
		c.expr(v.Expr)
		c.emit(runtime.OpEcho, 0)
	case *SetNode:
		c.expr(v.Value)
		slot, ok := c.resolveLocal(v.Name)
		if !ok {
			slot = c.declareLocal(v.Name)
		}
		c.emit(runtime.OpStoreLocal, slot)
	case *IfNode:
		c.compileIf(v)
	case *ForNode:
		c.compileFor(v)
	case *BlockNode:
		c.compileBlock(v)
	case *IncludeNode:
		c.compileInclude(v)
	}
}

func (c *compiler) compileIf(n *IfNode) {
	var endJumps []int
	for i, cond := range n.Conds {
		c.expr(cond)
		jf := c.emit(runtime.OpJumpFalse, 0)
		for _, bn := range n.Blocks[i] {
			c.node(bn)
		}
		endJumps = append(endJumps, c.emit(runtime.OpJump, 0))
		c.patch(jf, c.here())
	}
	for _, en := range n.Else {
		c.node(en)
	}
	end := c.here()
	for _, j := range endJumps {
		c.patch(j, end)
	}
}

func (c *compiler) compileFor(n *ForNode) {
	c.expr(n.Seq)

	loopIdx := len(c.loops)
	c.loops = append(c.loops, runtime.LoopSpec{})

	c.emit(runtime.OpIterInit, loopIdx)

	c.enterScope()
	valSlot := c.declareLocal(n.ValVar)
	keySlot := -1
	if n.KeyVar != "" {
		keySlot = c.declareLocal(n.KeyVar)
	}
	loopSlot := -1
	if n.usesLoopVar() {
		loopSlot = c.declareLocal(loopVarName)
	}

	loopTop := c.here()
	c.emit(runtime.OpIterNext, loopIdx)
	for _, bn := range n.Body {
		c.node(bn)
	}
	c.emit(runtime.OpJump, loopTop)
	c.exitScope()

	elseTarget := c.here()
	for _, en := range n.Else {
		c.node(en)
	}
	endTarget := c.here()

	c.loops[loopIdx] = runtime.LoopSpec{
		ValSlot:    valSlot,
		KeySlot:    keySlot,
		LoopSlot:   loopSlot,
		ElseTarget: elseTarget,
		EndTarget:  endTarget,
	}
}

func (c *compiler) compileBlock(n *BlockNode) {
	nameIdx := c.addConst(runtime.Str(n.Name))
	c.emit(runtime.OpRenderBlock, nameIdx)

	sub := newCompiler(c.res, c.autoescape, c.blocks)
	sub.enterScope()
	for _, bn := range n.Body {
		sub.node(bn)
	}
	if sub.err != nil {
		c.err = sub.err
		return
	}
	c.blocks[n.Name] = &runtime.Program{
		Name:       n.Name,
		Instrs:     sub.instrs,
		Consts:     sub.consts,
		Callables:  sub.callables,
		Loops:      sub.loops,
		NumLocals:  sub.numLocals,
		Autoescape: c.autoescape,
	}
}

func (c *compiler) compileInclude(n *IncludeNode) {
	lit, ok := n.Name.(*LiteralExpr)
	if !ok || lit.Val.Kind() != runtime.KindString {
		c.fail("include requires a string literal template name")
		return
	}
	c.emit(runtime.OpInclude, c.addConst(lit.Val))
}

func (c *compiler) expr(e Expr) {
	switch v := e.(type) {
	case *LiteralExpr:
		c.literal(v.Val)
	case *IdentExpr:
		if slot, ok := c.resolveLocal(v.Name); ok {
			c.emit(runtime.OpLoadLocal, slot)
		} else {
			c.emit(runtime.OpLoadVar, c.addConst(runtime.Str(v.Name)))
		}
	case *AttrExpr:
		c.expr(v.Target)
		c.emit(runtime.OpAttr, c.addConst(runtime.Str(v.Name)))
	case *IndexExpr:
		c.expr(v.Target)
		c.expr(v.Index)
		c.emit(runtime.OpIndex, 0)
	case *UnaryExpr:
		c.expr(v.X)
		if v.Op == TMinus {
			c.emit(runtime.OpNeg, 0)
		} else {
			c.emit(runtime.OpNot, 0)
		}
	case *BinaryExpr:
		c.expr(v.L)
		c.expr(v.R)
		c.emit(binaryOp(v.Op), 0)
	case *LogicalExpr:
		c.compileLogical(v)
	case *InExpr:
		c.expr(v.X)
		c.expr(v.Seq)
		if v.Not {
			c.emit(runtime.OpNotIn, 0)
		} else {
			c.emit(runtime.OpIn, 0)
		}
	case *TernaryExpr:
		c.compileTernary(v)
	case *ArrayExpr:
		for _, el := range v.Elems {
			c.expr(el)
		}
		c.emit(runtime.OpMakeList, len(v.Elems))
	case *MapExpr:
		for i := range v.Keys {
			c.expr(v.Keys[i])
			c.expr(v.Vals[i])
		}
		c.emit(runtime.OpMakeMap, len(v.Keys))
	case *CallExpr:
		c.compileCall(v)
	case *FilterExpr:
		c.compileFilter(v)
	}
}

func (c *compiler) literal(val runtime.Value) {
	switch val.Kind() {
	case runtime.KindNil:
		c.emit(runtime.OpNil, 0)
	case runtime.KindBool:
		if val.AsBool() {
			c.emit(runtime.OpTrue, 0)
		} else {
			c.emit(runtime.OpFalse, 0)
		}
	default:
		c.emit(runtime.OpConst, c.addConst(val))
	}
}

func (c *compiler) compileLogical(n *LogicalExpr) {
	c.expr(n.L)
	var jmp int
	if n.Op == TAnd {
		jmp = c.emit(runtime.OpJumpIfFalseOrPop, 0)
	} else {
		jmp = c.emit(runtime.OpJumpIfTrueOrPop, 0)
	}
	c.expr(n.R)
	c.patch(jmp, c.here())
}

func (c *compiler) compileTernary(n *TernaryExpr) {
	if n.Then == nil {
		c.expr(n.Cond)
		jmp := c.emit(runtime.OpJumpIfTrueOrPop, 0)
		c.expr(n.Else)
		c.patch(jmp, c.here())
		return
	}
	c.expr(n.Cond)
	jf := c.emit(runtime.OpJumpFalse, 0)
	c.expr(n.Then)
	jend := c.emit(runtime.OpJump, 0)
	c.patch(jf, c.here())
	c.expr(n.Else)
	c.patch(jend, c.here())
}

func (c *compiler) compileCall(n *CallExpr) {
	switch target := n.Target.(type) {
	case *IdentExpr:
		idx, ok := c.resolveCallable(target.Name, false)
		if !ok {
			c.fail("unknown function %q", target.Name)
			return
		}
		for _, a := range n.Args {
			c.expr(a)
		}
		c.emit(runtime.OpCall, packCall(idx, len(n.Args)))
	case *AttrExpr:
		c.expr(target.Target)
		for _, a := range n.Args {
			c.expr(a)
		}
		c.emit(runtime.OpCallMethod, packCall(c.addConst(runtime.Str(target.Name)), len(n.Args)))
	default:
		c.fail("expression is not callable")
	}
}

func (c *compiler) compileFilter(n *FilterExpr) {
	idx, ok := c.resolveCallable(n.Name, true)
	if !ok {
		c.fail("unknown filter %q", n.Name)
		return
	}
	c.expr(n.X)
	for _, a := range n.Args {
		c.expr(a)
	}
	c.emit(runtime.OpCall, packCall(idx, len(n.Args)+1))
}

func binaryOp(t TokenType) runtime.Op {
	switch t {
	case TPlus:
		return runtime.OpAdd
	case TMinus:
		return runtime.OpSub
	case TStar:
		return runtime.OpMul
	case TSlash:
		return runtime.OpDiv
	case TPercent:
		return runtime.OpMod
	case TPow:
		return runtime.OpPow
	case TTilde:
		return runtime.OpConcat
	case TEq:
		return runtime.OpEq
	case TNeq:
		return runtime.OpNeq
	case TLt:
		return runtime.OpLt
	case TGt:
		return runtime.OpGt
	case TLte:
		return runtime.OpLte
	case TGte:
		return runtime.OpGte
	}
	return runtime.OpNil
}

func packCall(index, argc int) int { return index<<8 | argc&0xff }
