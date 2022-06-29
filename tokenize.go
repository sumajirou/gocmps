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
	len  int       // Token length
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

func tokenize(code string) []Token {
	if len(code) == 0 {
		error_at(code, 0, "コードが空文字列です")
	}

	tokens := []Token{}
	i := 0 // codeのインデックス
	for i < len(code) {
		// Skip whitespace characters.
		if isSpace(code[i]) {
			i++
			continue
		}

		// Numeric literal
		if isDigit(code[i]) {
			j := 0
			for ; i+j < len(code) && isDigit(code[i+j]); j++ {
			}
			token := Token{kind: TK_NUM, val: code[i : i+j], loc: i, len: j}
			tokens = append(tokens, token)
			i += j
			continue
		}

		// Single-letter punctuators
		if isPunct(code[i]) {
			token := Token{kind: TK_RESERVED, val: string(code[i]), loc: i, len: 1}
			tokens = append(tokens, token)
			i++
			continue
		}
		error_at(code, i, "%vは認識できません", string(code[i]))
	}
	eof_token := Token{kind: TK_EOF, loc: i}
	tokens = append(tokens, eof_token)

	return tokens
}
