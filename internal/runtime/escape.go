package runtime

type SafeString string

func writeEscaped(dst []byte, s string) []byte {
	last := 0
	for i := 0; i < len(s); i++ {
		var repl string
		switch s[i] {
		case '&':
			repl = "&amp;"
		case '<':
			repl = "&lt;"
		case '>':
			repl = "&gt;"
		case '"':
			repl = "&#34;"
		case '\'':
			repl = "&#39;"
		default:
			continue
		}
		dst = append(dst, s[last:i]...)
		dst = append(dst, repl...)
		last = i + 1
	}
	return append(dst, s[last:]...)
}

func isSafe(v Value) (string, bool) {
	if v.kind == KindObject {
		if s, ok := v.obj.(SafeString); ok {
			return string(s), true
		}
	}
	return "", false
}
