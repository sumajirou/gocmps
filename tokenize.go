package main

type TokenKind int

const (
	TK_RESERVED TokenKind = iota // Keywords or punctuators
	TK_NUM                       // Numeric literals
	TK_EOF                       // End-of-file markers
)

type Token struct {
	kind TokenKind // Token kind
	val  string    // token value
	loc  int       // Token location
}
type Tokenizer struct {
	code   string
	tokens []Token
	i      int
}

func isDigit(c byte) bool {
	if '0' <= c && c <= '9' {
		return true
	}
	return false
}

func isLetter(c byte) bool {
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' {
		return true
	}
	return false
}

func isSpace(c byte) bool {
	if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
		return true
	}
	return false
}

func isPunct(c byte) bool {
	if isDigit(c) || isLetter(c) {
		return false
	}
	if '!' <= c && c <= '~' {
		return true
	}
	return false
}

func (tn *Tokenizer) peek(n int) string {
	if tn.i+n > len(tn.code) {
		return tn.code[tn.i:len(tn.code)]
	}
	return tn.code[tn.i : tn.i+n]
}

func (tn *Tokenizer) read(n int) string {
	if tn.i+n > len(tn.code) {
		result := tn.code[tn.i:len(tn.code)]
		tn.i = len(tn.code)
		return result
	}
	result := tn.code[tn.i : tn.i+n]
	tn.i += n
	return result
}

func (tn *Tokenizer) startswith(s string) bool {
	n := len(s)
	if tn.i+n > len(tn.code) {
		return false
	}
	return tn.code[tn.i:tn.i+n] == s
}

func (tn Tokenizer) tokenize() []Token {
	for tn.i < len(tn.code) {
		c := tn.peek(1)[0]

		// Skip whitespace characters.
		if isSpace(c) {
			tn.read(1)
			continue
		}

		// Numeric literal
		if isDigit(c) {
			token := Token{kind: TK_NUM, loc: tn.i, val: tn.read(1)}
			for isDigit(tn.peek(1)[0]) {
				token.val += tn.read(1)
			}
			tn.tokens = append(tn.tokens, token)
			continue
		}

		// Single-letter punctuators
		if isPunct(c) {
			token := Token{kind: TK_RESERVED, loc: tn.i, val: tn.read(1)}
			tn.tokens = append(tn.tokens, token)
			continue
		}

		error_at(tn.code, tn.i, "%sは認識できません", string(c))
	}
	eof_token := Token{kind: TK_EOF, loc: tn.i}
	tn.tokens = append(tn.tokens, eof_token)

	return tn.tokens
}
