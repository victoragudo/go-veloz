package compile

import (
	"fmt"
	"strconv"

	"github.com/victoragudo/go-veloz/internal/runtime"
)

const maxNestingDepth = 500

type parser struct {
	toks   []Token
	pos    int
	blocks map[string]*BlockNode
	parent string
	depth  int
}

func (p *parser) enterNesting() error {
	p.depth++
	if p.depth > maxNestingDepth {
		return p.errf("template is nested too deeply (limit %d)", maxNestingDepth)
	}
	return nil
}

func (p *parser) leaveNesting() { p.depth-- }

func Parse(src string) (*Template, error) {
	toks, err := Lex(src)
	if err != nil {
		return nil, err
	}
	p := &parser{toks: toks, blocks: map[string]*BlockNode{}}
	nodes, term, err := p.parseBody(nil)
	if err != nil {
		return nil, err
	}
	if term != "" {
		return nil, p.errf("unexpected end tag %q", term)
	}
	return &Template{Nodes: nodes, Blocks: p.blocks, Parent: p.parent}, nil
}

func (p *parser) cur() Token { return p.toks[p.pos] }

func (p *parser) peek(n int) Token {
	i := p.pos + n
	if i < len(p.toks) {
		return p.toks[i]
	}
	return p.toks[len(p.toks)-1]
}

func (p *parser) advance() Token {
	t := p.toks[p.pos]
	if p.pos < len(p.toks)-1 {
		p.pos++
	}
	return t
}

func (p *parser) expect(tt TokenType) (Token, error) {
	if p.cur().Type != tt {
		return Token{}, p.errf("expected %s, got %s", tt, p.cur().Type)
	}
	return p.advance(), nil
}

func (p *parser) errf(format string, args ...any) error {
	return fmt.Errorf("line %d: %s", p.cur().Line, fmt.Sprintf(format, args...))
}

func (p *parser) parseBody(stops map[string]bool) ([]Node, string, error) {
	if err := p.enterNesting(); err != nil {
		return nil, "", err
	}
	defer p.leaveNesting()
	var nodes []Node
	for {
		t := p.cur()
		switch t.Type {
		case TEOF:
			return nodes, "", nil
		case TText:
			nodes = append(nodes, &TextNode{Text: t.Val})
			p.advance()
		case TVarOpen:
			p.advance()
			e, err := p.parseExpr()
			if err != nil {
				return nil, "", err
			}
			if _, err := p.expect(TVarClose); err != nil {
				return nil, "", err
			}
			nodes = append(nodes, &PrintNode{Expr: e})
		case TBlockOpen:
			if stops != nil && p.peek(1).Type == TIdent && stops[p.peek(1).Val] {
				term := p.peek(1).Val
				p.advance()
				p.advance()
				return nodes, term, nil
			}
			node, err := p.parseStatement()
			if err != nil {
				return nil, "", err
			}
			if node != nil {
				nodes = append(nodes, node)
			}
		default:
			return nil, "", p.errf("unexpected %s", t.Type)
		}
	}
}

func (p *parser) parseStatement() (Node, error) {
	p.advance()
	if p.cur().Type != TIdent {
		return nil, p.errf("expected tag name, got %s", p.cur().Type)
	}
	switch p.cur().Val {
	case "if":
		return p.parseIf()
	case "for":
		return p.parseFor()
	case "set":
		return p.parseSet()
	case "block":
		return p.parseBlock()
	case "include":
		return p.parseInclude()
	case "extends":
		return p.parseExtends()
	default:
		return nil, p.errf("unknown tag %q", p.cur().Val)
	}
}

func (p *parser) parseIf() (Node, error) {
	p.advance()
	node := &IfNode{}
	stops := map[string]bool{"elseif": true, "else": true, "endif": true}
	for {
		cond, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TBlockClose); err != nil {
			return nil, err
		}
		body, term, err := p.parseBody(stops)
		if err != nil {
			return nil, err
		}
		node.Conds = append(node.Conds, cond)
		node.Blocks = append(node.Blocks, body)
		switch term {
		case "elseif":
			continue
		case "else":
			if _, err := p.expect(TBlockClose); err != nil {
				return nil, err
			}
			elseBody, term2, err := p.parseBody(map[string]bool{"endif": true})
			if err != nil {
				return nil, err
			}
			if term2 != "endif" {
				return nil, p.errf("expected endif")
			}
			if _, err := p.expect(TBlockClose); err != nil {
				return nil, err
			}
			node.Else = elseBody
			return node, nil
		case "endif":
			if _, err := p.expect(TBlockClose); err != nil {
				return nil, err
			}
			return node, nil
		default:
			return nil, p.errf("unclosed if")
		}
	}
}

func (p *parser) parseFor() (Node, error) {
	p.advance()
	first, err := p.expect(TIdent)
	if err != nil {
		return nil, err
	}
	node := &ForNode{ValVar: first.Val}
	if p.cur().Type == TComma {
		p.advance()
		second, err := p.expect(TIdent)
		if err != nil {
			return nil, err
		}
		node.KeyVar = first.Val
		node.ValVar = second.Val
	}
	if _, err := p.expect(TIn); err != nil {
		return nil, err
	}
	seq, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	node.Seq = seq
	if _, err := p.expect(TBlockClose); err != nil {
		return nil, err
	}
	body, term, err := p.parseBody(map[string]bool{"else": true, "endfor": true})
	if err != nil {
		return nil, err
	}
	node.Body = body
	switch term {
	case "else":
		if _, err := p.expect(TBlockClose); err != nil {
			return nil, err
		}
		elseBody, term2, err := p.parseBody(map[string]bool{"endfor": true})
		if err != nil {
			return nil, err
		}
		if term2 != "endfor" {
			return nil, p.errf("expected endfor")
		}
		if _, err := p.expect(TBlockClose); err != nil {
			return nil, err
		}
		node.Else = elseBody
	case "endfor":
		if _, err := p.expect(TBlockClose); err != nil {
			return nil, err
		}
	default:
		return nil, p.errf("unclosed for")
	}
	return node, nil
}

func (p *parser) parseSet() (Node, error) {
	p.advance()
	name, err := p.expect(TIdent)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TAssign); err != nil {
		return nil, err
	}
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TBlockClose); err != nil {
		return nil, err
	}
	return &SetNode{Name: name.Val, Value: val}, nil
}

func (p *parser) parseBlock() (Node, error) {
	p.advance()
	name, err := p.expect(TIdent)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TBlockClose); err != nil {
		return nil, err
	}
	body, term, err := p.parseBody(map[string]bool{"endblock": true})
	if err != nil {
		return nil, err
	}
	if term != "endblock" {
		return nil, p.errf("expected endblock")
	}
	if p.cur().Type == TIdent {
		p.advance()
	}
	if _, err := p.expect(TBlockClose); err != nil {
		return nil, err
	}
	node := &BlockNode{Name: name.Val, Body: body}
	p.blocks[name.Val] = node
	return node, nil
}

func (p *parser) parseInclude() (Node, error) {
	p.advance()
	name, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TBlockClose); err != nil {
		return nil, err
	}
	return &IncludeNode{Name: name}, nil
}

func (p *parser) parseExtends() (Node, error) {
	p.advance()
	name, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	lit, ok := name.(*LiteralExpr)
	if !ok || lit.Val.Kind() != runtime.KindString {
		return nil, p.errf("extends requires a string literal")
	}
	if _, err := p.expect(TBlockClose); err != nil {
		return nil, err
	}
	p.parent = lit.Val.String()
	return nil, nil
}

func (p *parser) parseExpr() (Expr, error) {
	if err := p.enterNesting(); err != nil {
		return nil, err
	}
	defer p.leaveNesting()
	return p.parseTernary()
}

func (p *parser) parseTernary() (Expr, error) {
	cond, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.cur().Type != TQuestion {
		return cond, nil
	}
	p.advance()
	if p.cur().Type == TColon {
		p.advance()
		elseE, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &TernaryExpr{Cond: cond, Then: nil, Else: elseE}, nil
	}
	thenE, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TColon); err != nil {
		return nil, err
	}
	elseE, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &TernaryExpr{Cond: cond, Then: thenE, Else: elseE}, nil
}

func (p *parser) parseOr() (Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.cur().Type == TOr {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &LogicalExpr{Op: TOr, L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (Expr, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	for p.cur().Type == TAnd {
		p.advance()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &LogicalExpr{Op: TAnd, L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseNot() (Expr, error) {
	if p.cur().Type == TNot {
		p.advance()
		x, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: TNot, X: x}, nil
	}
	return p.parseComparison()
}

func (p *parser) parseComparison() (Expr, error) {
	left, err := p.parseConcat()
	if err != nil {
		return nil, err
	}
	switch p.cur().Type {
	case TEq, TNeq, TLt, TGt, TLte, TGte:
		op := p.advance().Type
		right, err := p.parseConcat()
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Op: op, L: left, R: right}, nil
	case TIn:
		p.advance()
		right, err := p.parseConcat()
		if err != nil {
			return nil, err
		}
		return &InExpr{X: left, Seq: right, Not: false}, nil
	case TNot:
		if p.peek(1).Type == TIn {
			p.advance()
			p.advance()
			right, err := p.parseConcat()
			if err != nil {
				return nil, err
			}
			return &InExpr{X: left, Seq: right, Not: true}, nil
		}
	}
	return left, nil
}

func (p *parser) parseConcat() (Expr, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}
	for p.cur().Type == TTilde {
		p.advance()
		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: TTilde, L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseAdditive() (Expr, error) {
	left, err := p.parseMul()
	if err != nil {
		return nil, err
	}
	for p.cur().Type == TPlus || p.cur().Type == TMinus {
		op := p.advance().Type
		right, err := p.parseMul()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseMul() (Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.cur().Type == TStar || p.cur().Type == TSlash || p.cur().Type == TPercent {
		op := p.advance().Type
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseUnary() (Expr, error) {
	if p.cur().Type == TMinus {
		p.advance()
		x, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: TMinus, X: x}, nil
	}
	if p.cur().Type == TPlus {
		p.advance()
		return p.parseUnary()
	}
	return p.parsePow()
}

func (p *parser) parsePow() (Expr, error) {
	left, err := p.parsePostfix()
	if err != nil {
		return nil, err
	}
	if p.cur().Type == TPow {
		p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Op: TPow, L: left, R: right}, nil
	}
	return left, nil
}

func (p *parser) parsePostfix() (Expr, error) {
	node, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		switch p.cur().Type {
		case TDot:
			p.advance()
			name, err := p.expect(TIdent)
			if err != nil {
				return nil, err
			}
			node = &AttrExpr{Target: node, Name: name.Val}
		case TLBracket:
			p.advance()
			idx, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(TRBracket); err != nil {
				return nil, err
			}
			node = &IndexExpr{Target: node, Index: idx}
		case TLParen:
			p.advance()
			args, err := p.parseArgs()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(TRParen); err != nil {
				return nil, err
			}
			node = &CallExpr{Target: node, Args: args}
		case TPipe:
			p.advance()
			fname, err := p.expect(TIdent)
			if err != nil {
				return nil, err
			}
			var fargs []Expr
			if p.cur().Type == TLParen {
				p.advance()
				fargs, err = p.parseArgs()
				if err != nil {
					return nil, err
				}
				if _, err := p.expect(TRParen); err != nil {
					return nil, err
				}
			}
			node = &FilterExpr{X: node, Name: fname.Val, Args: fargs}
		default:
			return node, nil
		}
	}
}

func (p *parser) parseArgs() ([]Expr, error) {
	var args []Expr
	if p.cur().Type == TRParen {
		return args, nil
	}
	for {
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, e)
		if p.cur().Type == TComma {
			p.advance()
			continue
		}
		break
	}
	return args, nil
}

func (p *parser) parsePrimary() (Expr, error) {
	t := p.cur()
	switch t.Type {
	case TInt:
		p.advance()
		n, err := strconv.ParseInt(t.Val, 10, 64)
		if err != nil {
			return nil, p.errf("invalid integer %q", t.Val)
		}
		return &LiteralExpr{Val: runtime.Int(n)}, nil
	case TFloat:
		p.advance()
		f, err := strconv.ParseFloat(t.Val, 64)
		if err != nil {
			return nil, p.errf("invalid float %q", t.Val)
		}
		return &LiteralExpr{Val: runtime.Float(f)}, nil
	case TString:
		p.advance()
		return &LiteralExpr{Val: runtime.Str(t.Val)}, nil
	case TTrue:
		p.advance()
		return &LiteralExpr{Val: runtime.Bool(true)}, nil
	case TFalse:
		p.advance()
		return &LiteralExpr{Val: runtime.Bool(false)}, nil
	case TNull:
		p.advance()
		return &LiteralExpr{Val: runtime.Nil()}, nil
	case TIdent:
		p.advance()
		return &IdentExpr{Name: t.Val}, nil
	case TLParen:
		p.advance()
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TRParen); err != nil {
			return nil, err
		}
		return e, nil
	case TLBracket:
		return p.parseArray()
	case TLBrace:
		return p.parseMap()
	}
	return nil, p.errf("unexpected %s in expression", t.Type)
}

func (p *parser) parseArray() (Expr, error) {
	p.advance()
	node := &ArrayExpr{}
	if p.cur().Type == TRBracket {
		p.advance()
		return node, nil
	}
	for {
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		node.Elems = append(node.Elems, e)
		if p.cur().Type == TComma {
			p.advance()
			if p.cur().Type == TRBracket {
				break
			}
			continue
		}
		break
	}
	if _, err := p.expect(TRBracket); err != nil {
		return nil, err
	}
	return node, nil
}

func (p *parser) parseMap() (Expr, error) {
	p.advance()
	node := &MapExpr{}
	if p.cur().Type == TRBrace {
		p.advance()
		return node, nil
	}
	for {
		var key Expr
		if p.cur().Type == TIdent {
			key = &LiteralExpr{Val: runtime.Str(p.advance().Val)}
		} else {
			k, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			key = k
		}
		if _, err := p.expect(TColon); err != nil {
			return nil, err
		}
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		node.Keys = append(node.Keys, key)
		node.Vals = append(node.Vals, val)
		if p.cur().Type == TComma {
			p.advance()
			if p.cur().Type == TRBrace {
				break
			}
			continue
		}
		break
	}
	if _, err := p.expect(TRBrace); err != nil {
		return nil, err
	}
	return node, nil
}
