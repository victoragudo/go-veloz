package compile

type TokenType int

const (
	TEOF TokenType = iota
	TText
	TVarOpen
	TVarClose
	TBlockOpen
	TBlockClose

	TIdent
	TInt
	TFloat
	TString
	TTrue
	TFalse
	TNull

	TPlus
	TMinus
	TStar
	TSlash
	TPercent
	TPow
	TTilde
	TPipe
	TDot
	TComma
	TColon
	TLParen
	TRParen
	TLBracket
	TRBracket
	TLBrace
	TRBrace
	TAssign
	TEq
	TNeq
	TLt
	TGt
	TLte
	TGte
	TAnd
	TOr
	TNot
	TIn
	TQuestion
)

type Token struct {
	Type TokenType
	Val  string
	Line int
}

var keywords = map[string]TokenType{
	"true":  TTrue,
	"false": TFalse,
	"null":  TNull,
	"none":  TNull,
	"and":   TAnd,
	"or":    TOr,
	"not":   TNot,
	"in":    TIn,
}

func (t TokenType) String() string {
	switch t {
	case TEOF:
		return "end of input"
	case TText:
		return "text"
	case TVarOpen:
		return "{{"
	case TVarClose:
		return "}}"
	case TBlockOpen:
		return "{%"
	case TBlockClose:
		return "%}"
	case TIdent:
		return "identifier"
	case TInt:
		return "integer"
	case TFloat:
		return "float"
	case TString:
		return "string"
	case TTrue:
		return "true"
	case TFalse:
		return "false"
	case TNull:
		return "null"
	case TPlus:
		return "+"
	case TMinus:
		return "-"
	case TStar:
		return "*"
	case TSlash:
		return "/"
	case TPercent:
		return "%"
	case TPow:
		return "**"
	case TTilde:
		return "~"
	case TPipe:
		return "|"
	case TDot:
		return "."
	case TComma:
		return ","
	case TColon:
		return ":"
	case TLParen:
		return "("
	case TRParen:
		return ")"
	case TLBracket:
		return "["
	case TRBracket:
		return "]"
	case TLBrace:
		return "{"
	case TRBrace:
		return "}"
	case TAssign:
		return "="
	case TEq:
		return "=="
	case TNeq:
		return "!="
	case TLt:
		return "<"
	case TGt:
		return ">"
	case TLte:
		return "<="
	case TGte:
		return ">="
	case TAnd:
		return "and"
	case TOr:
		return "or"
	case TNot:
		return "not"
	case TIn:
		return "in"
	case TQuestion:
		return "?"
	}
	return "unknown"
}
