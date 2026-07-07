package compile

import (
	"fmt"
	"strings"
)

type lexer struct {
	src      string
	pos      int
	line     int
	toks     []Token
	trimNext bool
}

func Lex(src string) ([]Token, error) {
	lx := &lexer{src: src, line: 1}
	if err := lx.run(); err != nil {
		return nil, err
	}
	lx.emit(TEOF, "")
	return lx.toks, nil
}

func (lx *lexer) emit(t TokenType, v string) {
	lx.toks = append(lx.toks, Token{Type: t, Val: v, Line: lx.line})
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
	if lx.trimNext {
		s = strings.TrimLeft(s, " \t\r\n")
		lx.trimNext = false
	}
	lx.countLines(s)
	if s == "" {
		return
	}
	lx.emit(TText, s)
}

func (lx *lexer) countLines(s string) {
	lx.line += strings.Count(s, "\n")
}

func (lx *lexer) lexTag() error {
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
		lx.emit(TVarOpen, "")
		return lx.lexExpr(TVarClose)
	case '%':
		lx.emit(TBlockOpen, "")
		return lx.lexExpr(TBlockClose)
	}
	return nil
}

func (lx *lexer) skipComment() error {
	end := strings.Index(lx.src[lx.pos:], "#}")
	if end < 0 {
		return fmt.Errorf("line %d: unclosed comment", lx.line)
	}
	body := lx.src[lx.pos : lx.pos+end]
	lx.countLines(body)
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
			return fmt.Errorf("line %d: unclosed tag", lx.line)
		}
		if lx.src[lx.pos] == '-' && strings.HasPrefix(lx.src[lx.pos+1:], closeStr) {
			lx.trimNext = true
			lx.pos += 1 + len(closeStr)
			lx.emit(closer, "")
			return nil
		}
		if strings.HasPrefix(lx.src[lx.pos:], closeStr) {
			lx.pos += len(closeStr)
			lx.emit(closer, "")
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
		lx.emit(kw, word)
		return nil
	}
	lx.emit(TIdent, word)
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
		lx.emit(TFloat, lx.src[start:lx.pos])
	} else {
		lx.emit(TInt, lx.src[start:lx.pos])
	}
	return nil
}

func (lx *lexer) lexString(quote byte) error {
	lx.pos++
	var sb strings.Builder
	for lx.pos < len(lx.src) {
		c := lx.src[lx.pos]
		if c == quote {
			lx.pos++
			lx.emit(TString, sb.String())
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
		}
		sb.WriteByte(c)
		lx.pos++
	}
	return fmt.Errorf("line %d: unterminated string", lx.line)
}

func (lx *lexer) lexOperator() error {
	rest := lx.src[lx.pos:]
	two := ""
	if len(rest) >= 2 {
		two = rest[:2]
	}
	switch two {
	case "==":
		lx.pos += 2
		lx.emit(TEq, two)
		return nil
	case "!=":
		lx.pos += 2
		lx.emit(TNeq, two)
		return nil
	case "<=":
		lx.pos += 2
		lx.emit(TLte, two)
		return nil
	case ">=":
		lx.pos += 2
		lx.emit(TGte, two)
		return nil
	case "**":
		lx.pos += 2
		lx.emit(TPow, two)
		return nil
	case "&&":
		lx.pos += 2
		lx.emit(TAnd, two)
		return nil
	case "||":
		lx.pos += 2
		lx.emit(TOr, two)
		return nil
	}

	c := lx.src[lx.pos]
	lx.pos++
	switch c {
	case '+':
		lx.emit(TPlus, "+")
	case '-':
		lx.emit(TMinus, "-")
	case '*':
		lx.emit(TStar, "*")
	case '/':
		lx.emit(TSlash, "/")
	case '%':
		lx.emit(TPercent, "%")
	case '~':
		lx.emit(TTilde, "~")
	case '|':
		lx.emit(TPipe, "|")
	case '.':
		lx.emit(TDot, ".")
	case ',':
		lx.emit(TComma, ",")
	case ':':
		lx.emit(TColon, ":")
	case '(':
		lx.emit(TLParen, "(")
	case ')':
		lx.emit(TRParen, ")")
	case '[':
		lx.emit(TLBracket, "[")
	case ']':
		lx.emit(TRBracket, "]")
	case '{':
		lx.emit(TLBrace, "{")
	case '}':
		lx.emit(TRBrace, "}")
	case '=':
		lx.emit(TAssign, "=")
	case '<':
		lx.emit(TLt, "<")
	case '>':
		lx.emit(TGt, ">")
	case '!':
		lx.emit(TNot, "!")
	case '?':
		lx.emit(TQuestion, "?")
	default:
		return fmt.Errorf("line %d: unexpected character %q", lx.line, string(c))
	}
	return nil
}
