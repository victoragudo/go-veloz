package compile

import (
	"fmt"
	"strings"
)

type lexer struct {
	src       string
	pos       int
	line      int
	lineStart int
	toks      []Token
	trimNext  bool
}

func Lex(src string) ([]Token, error) {
	lx := &lexer{src: src, line: 1}
	if err := lx.run(); err != nil {
		return nil, err
	}
	lx.emitAt(TEOF, "", lx.pos)
	return lx.toks, nil
}

func (lx *lexer) col(offset int) int { return offset - lx.lineStart + 1 }

func (lx *lexer) emitAt(t TokenType, v string, start int) {
	lx.toks = append(lx.toks, Token{Type: t, Val: v, Line: lx.line, Col: lx.col(start)})
}

func (lx *lexer) advanceLines(s string, base int) {
	if i := strings.LastIndexByte(s, '\n'); i >= 0 {
		lx.line += strings.Count(s, "\n")
		lx.lineStart = base + i + 1
	}
}

func (lx *lexer) run() error {
	for lx.pos < len(lx.src) {
		next := lx.findOpen()
		if next < 0 {
			lx.pushText(lx.src[lx.pos:])
			lx.pos = len(lx.src)
			break
		}
		if next > lx.pos {
			lx.pushText(lx.src[lx.pos:next])
			lx.pos = next
		}
		if err := lx.lexTag(); err != nil {
			return err
		}
	}
	return nil
}

func (lx *lexer) findOpen() int {
	i := lx.pos
	for i < len(lx.src)-1 {
		if lx.src[i] == '{' {
			switch lx.src[i+1] {
			case '{', '%', '#':
				return i
			}
		}
		i++
	}
	return -1
}

func (lx *lexer) pushText(s string) {
	base := lx.pos
	val := s
	if lx.trimNext {
		val = strings.TrimLeft(s, " \t\r\n")
		lx.trimNext = false
	}
	prefix := s[:len(s)-len(val)]
	lx.advanceLines(prefix, base)
	start := base + len(prefix)
	if val != "" {
		lx.emitAt(TText, val, start)
	}
	lx.advanceLines(val, start)
}

func (lx *lexer) lexTag() error {
	start := lx.pos
	kind := lx.src[lx.pos+1]
	trimLeft := false
	adv := 2
	if lx.pos+2 < len(lx.src) && lx.src[lx.pos+2] == '-' {
		trimLeft = true
		adv = 3
	}
	if trimLeft && len(lx.toks) > 0 && lx.toks[len(lx.toks)-1].Type == TText {
		last := &lx.toks[len(lx.toks)-1]
		last.Val = strings.TrimRight(last.Val, " \t\r\n")
	}
	lx.pos += adv

	switch kind {
	case '#':
		return lx.skipComment()
	case '{':
		lx.emitAt(TVarOpen, "", start)
		return lx.lexExpr(TVarClose)
	case '%':
		lx.emitAt(TBlockOpen, "", start)
		return lx.lexExpr(TBlockClose)
	}
	return nil
}

func (lx *lexer) skipComment() error {
	end := strings.Index(lx.src[lx.pos:], "#}")
	if end < 0 {
		return fmt.Errorf("%d:%d: unclosed comment", lx.line, lx.col(lx.pos))
	}
	body := lx.src[lx.pos : lx.pos+end]
	lx.advanceLines(body, lx.pos)
	lx.pos += end + 2
	if strings.HasSuffix(body, "-") {
		lx.trimNext = true
	}
	return nil
}

func (lx *lexer) lexExpr(closer TokenType) error {
	closeStr := "}}"
	if closer == TBlockClose {
		closeStr = "%}"
	}
	for {
		lx.skipSpace()
		if lx.pos >= len(lx.src) {
			return fmt.Errorf("%d:%d: unclosed tag", lx.line, lx.col(lx.pos))
		}
		if lx.src[lx.pos] == '-' && strings.HasPrefix(lx.src[lx.pos+1:], closeStr) {
			start := lx.pos
			lx.trimNext = true
			lx.pos += 1 + len(closeStr)
			lx.emitAt(closer, "", start)
			return nil
		}
		if strings.HasPrefix(lx.src[lx.pos:], closeStr) {
			start := lx.pos
			lx.pos += len(closeStr)
			lx.emitAt(closer, "", start)
			return nil
		}
		if err := lx.lexToken(); err != nil {
			return err
		}
	}
}

func (lx *lexer) skipSpace() {
	for lx.pos < len(lx.src) {
		c := lx.src[lx.pos]
		if c == '\n' {
			lx.line++
			lx.lineStart = lx.pos + 1
			lx.pos++
		} else if c == ' ' || c == '\t' || c == '\r' {
			lx.pos++
		} else {
			break
		}
	}
}

func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func (lx *lexer) lexToken() error {
	c := lx.src[lx.pos]
	switch {
	case isIdentStart(c):
		return lx.lexIdent()
	case isDigit(c):
		return lx.lexNumber()
	case c == '\'' || c == '"':
		return lx.lexString(c)
	}
	return lx.lexOperator()
}

func (lx *lexer) lexIdent() error {
	start := lx.pos
	for lx.pos < len(lx.src) && isIdentPart(lx.src[lx.pos]) {
		lx.pos++
	}
	word := lx.src[start:lx.pos]
	if kw, ok := keywords[word]; ok {
		lx.emitAt(kw, word, start)
		return nil
	}
	lx.emitAt(TIdent, word, start)
	return nil
}

func (lx *lexer) lexNumber() error {
	start := lx.pos
	isFloat := false
	for lx.pos < len(lx.src) && isDigit(lx.src[lx.pos]) {
		lx.pos++
	}
	if lx.pos < len(lx.src) && lx.src[lx.pos] == '.' && lx.pos+1 < len(lx.src) && isDigit(lx.src[lx.pos+1]) {
		isFloat = true
		lx.pos++
		for lx.pos < len(lx.src) && isDigit(lx.src[lx.pos]) {
			lx.pos++
		}
	}
	if lx.pos < len(lx.src) && (lx.src[lx.pos] == 'e' || lx.src[lx.pos] == 'E') {
		save := lx.pos
		lx.pos++
		if lx.pos < len(lx.src) && (lx.src[lx.pos] == '+' || lx.src[lx.pos] == '-') {
			lx.pos++
		}
		if lx.pos < len(lx.src) && isDigit(lx.src[lx.pos]) {
			isFloat = true
			for lx.pos < len(lx.src) && isDigit(lx.src[lx.pos]) {
				lx.pos++
			}
		} else {
			lx.pos = save
		}
	}
	if isFloat {
		lx.emitAt(TFloat, lx.src[start:lx.pos], start)
	} else {
		lx.emitAt(TInt, lx.src[start:lx.pos], start)
	}
	return nil
}

func (lx *lexer) lexString(quote byte) error {
	startLine, startCol := lx.line, lx.col(lx.pos)
	lx.pos++
	var sb strings.Builder
	for lx.pos < len(lx.src) {
		c := lx.src[lx.pos]
		if c == quote {
			lx.pos++
			lx.toks = append(lx.toks, Token{Type: TString, Val: sb.String(), Line: startLine, Col: startCol})
			return nil
		}
		if c == '\\' && lx.pos+1 < len(lx.src) {
			lx.pos++
			esc := lx.src[lx.pos]
			switch esc {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '\'':
				sb.WriteByte('\'')
			case '"':
				sb.WriteByte('"')
			default:
				sb.WriteByte(esc)
			}
			lx.pos++
			continue
		}
		if c == '\n' {
			lx.line++
			lx.lineStart = lx.pos + 1
		}
		sb.WriteByte(c)
		lx.pos++
	}
	return fmt.Errorf("%d:%d: unterminated string", startLine, startCol)
}

var twoCharOps = map[string]TokenType{
	"==": TEq,
	"!=": TNeq,
	"<=": TLte,
	">=": TGte,
	"**": TPow,
	"&&": TAnd,
	"||": TOr,
}

var oneCharOps = map[byte]TokenType{
	'+': TPlus,
	'-': TMinus,
	'*': TStar,
	'/': TSlash,
	'%': TPercent,
	'~': TTilde,
	'|': TPipe,
	'.': TDot,
	',': TComma,
	':': TColon,
	'(': TLParen,
	')': TRParen,
	'[': TLBracket,
	']': TRBracket,
	'{': TLBrace,
	'}': TRBrace,
	'=': TAssign,
	'<': TLt,
	'>': TGt,
	'!': TNot,
	'?': TQuestion,
}

func (lx *lexer) lexOperator() error {
	start := lx.pos
	if start+1 < len(lx.src) {
		two := lx.src[start : start+2]
		if tt, ok := twoCharOps[two]; ok {
			lx.pos += 2
			lx.emitAt(tt, two, start)
			return nil
		}
	}
	c := lx.src[start]
	if tt, ok := oneCharOps[c]; ok {
		lx.pos++
		lx.emitAt(tt, string(c), start)
		return nil
	}
	return fmt.Errorf("%d:%d: unexpected character %q", lx.line, lx.col(start), string(c))
}
