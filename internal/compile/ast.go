package compile

import "veloz/internal/runtime"

type Node interface{ node() }

type TextNode struct{ Text string }

type PrintNode struct{ Expr Expr }

type IfNode struct {
	Conds  []Expr
	Blocks [][]Node
	Else   []Node
}

type ForNode struct {
	KeyVar string
	ValVar string
	Seq    Expr
	Body   []Node
	Else   []Node
}

type SetNode struct {
	Name  string
	Value Expr
}

type BlockNode struct {
	Name string
	Body []Node
}

type IncludeNode struct{ Name Expr }

func (*TextNode) node()    {}
func (*PrintNode) node()   {}
func (*IfNode) node()      {}
func (*ForNode) node()     {}
func (*SetNode) node()     {}
func (*BlockNode) node()   {}
func (*IncludeNode) node() {}

const loopVarName = "loop"

func (n *ForNode) usesLoopVar() bool {
	return anyNodeUsesLoop(n.Body)
}

func anyNodeUsesLoop(nodes []Node) bool {
	for _, n := range nodes {
		if nodeUsesLoop(n) {
			return true
		}
	}
	return false
}

func nodeUsesLoop(n Node) bool {
	switch v := n.(type) {
	case *PrintNode:
		return exprUsesLoop(v.Expr)
	case *IfNode:
		for _, cond := range v.Conds {
			if exprUsesLoop(cond) {
				return true
			}
		}
		for _, blk := range v.Blocks {
			if anyNodeUsesLoop(blk) {
				return true
			}
		}
		return anyNodeUsesLoop(v.Else)
	case *ForNode:
		return exprUsesLoop(v.Seq) || anyNodeUsesLoop(v.Body) || anyNodeUsesLoop(v.Else)
	case *SetNode:
		return exprUsesLoop(v.Value)
	case *IncludeNode:
		return exprUsesLoop(v.Name)
	}
	return false
}

func exprUsesLoop(e Expr) bool {
	switch v := e.(type) {
	case *IdentExpr:
		return v.Name == loopVarName
	case *AttrExpr:
		return exprUsesLoop(v.Target)
	case *IndexExpr:
		return exprUsesLoop(v.Target) || exprUsesLoop(v.Index)
	case *UnaryExpr:
		return exprUsesLoop(v.X)
	case *BinaryExpr:
		return exprUsesLoop(v.L) || exprUsesLoop(v.R)
	case *LogicalExpr:
		return exprUsesLoop(v.L) || exprUsesLoop(v.R)
	case *CallExpr:
		if exprUsesLoop(v.Target) {
			return true
		}
		return anyExprUsesLoop(v.Args)
	case *FilterExpr:
		return exprUsesLoop(v.X) || anyExprUsesLoop(v.Args)
	case *ArrayExpr:
		return anyExprUsesLoop(v.Elems)
	case *MapExpr:
		return anyExprUsesLoop(v.Keys) || anyExprUsesLoop(v.Vals)
	case *TernaryExpr:
		if exprUsesLoop(v.Cond) || exprUsesLoop(v.Else) {
			return true
		}
		return v.Then != nil && exprUsesLoop(v.Then)
	case *InExpr:
		return exprUsesLoop(v.X) || exprUsesLoop(v.Seq)
	}
	return false
}

func anyExprUsesLoop(exprs []Expr) bool {
	for _, e := range exprs {
		if exprUsesLoop(e) {
			return true
		}
	}
	return false
}

type Expr interface{ expr() }

type LiteralExpr struct{ Val runtime.Value }

type IdentExpr struct{ Name string }

type AttrExpr struct {
	Target Expr
	Name   string
}

type IndexExpr struct {
	Target Expr
	Index  Expr
}

type UnaryExpr struct {
	Op TokenType
	X  Expr
}

type BinaryExpr struct {
	Op TokenType
	L  Expr
	R  Expr
}

type LogicalExpr struct {
	Op TokenType
	L  Expr
	R  Expr
}

type CallExpr struct {
	Target Expr
	Args   []Expr
}

type FilterExpr struct {
	X    Expr
	Name string
	Args []Expr
}

type ArrayExpr struct{ Elems []Expr }

type MapExpr struct {
	Keys []Expr
	Vals []Expr
}

type TernaryExpr struct {
	Cond Expr
	Then Expr
	Else Expr
}

type InExpr struct {
	X   Expr
	Seq Expr
	Not bool
}

func (*LiteralExpr) expr() {}
func (*IdentExpr) expr()   {}
func (*AttrExpr) expr()    {}
func (*IndexExpr) expr()   {}
func (*UnaryExpr) expr()   {}
func (*BinaryExpr) expr()  {}
func (*LogicalExpr) expr() {}
func (*CallExpr) expr()    {}
func (*FilterExpr) expr()  {}
func (*ArrayExpr) expr()   {}
func (*MapExpr) expr()     {}
func (*TernaryExpr) expr() {}
func (*InExpr) expr()      {}

type Template struct {
	Nodes  []Node
	Blocks map[string]*BlockNode
	Parent string
}
