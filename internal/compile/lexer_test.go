package compile

import "testing"

func tokenTypes(toks []Token) []TokenType {
	out := make([]TokenType, 0, len(toks))
	for _, t := range toks {
		out = append(out, t.Type)
	}
	return out
}

func TestLexBasic(t *testing.T) {
	toks, err := Lex(`Hola {{ name }}!`)
	if err != nil {
		t.Fatal(err)
	}
	want := []TokenType{TText, TVarOpen, TIdent, TVarClose, TText, TEOF}
	got := tokenTypes(toks)
	if len(got) != len(want) {
		t.Fatalf("token count: got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token %d: got %s want %s", i, got[i], want[i])
		}
	}
}

func TestLexExpression(t *testing.T) {
	toks, err := Lex(`{{ (a + 3) * 2 | round }}`)
	if err != nil {
		t.Fatal(err)
	}
	want := []TokenType{
		TVarOpen, TLParen, TIdent, TPlus, TInt, TRParen,
		TStar, TInt, TPipe, TIdent, TVarClose, TEOF,
	}
	got := tokenTypes(toks)
	if len(got) != len(want) {
		t.Fatalf("token count mismatch: got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token %d: got %s want %s", i, got[i], want[i])
		}
	}
}

func TestLexStatement(t *testing.T) {
	toks, err := Lex(`{% for item in items %}x{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}
	want := []TokenType{
		TBlockOpen, TIdent, TIdent, TIn, TIdent, TBlockClose,
		TText, TBlockOpen, TIdent, TBlockClose, TEOF,
	}
	got := tokenTypes(toks)
	for i := range want {
		if i >= len(got) || got[i] != want[i] {
			t.Fatalf("mismatch at %d: got %v", i, got)
		}
	}
}

func TestLexString(t *testing.T) {
	toks, err := Lex(`{{ "a\"b\n" }}`)
	if err != nil {
		t.Fatal(err)
	}
	if toks[1].Type != TString || toks[1].Val != "a\"b\n" {
		t.Errorf("string lexing failed: %q", toks[1].Val)
	}
}

func TestLexComment(t *testing.T) {
	toks, err := Lex(`a{# ignore me #}b`)
	if err != nil {
		t.Fatal(err)
	}
	want := []TokenType{TText, TText, TEOF}
	got := tokenTypes(toks)
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
}
