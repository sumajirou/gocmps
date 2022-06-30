package main

type TokenKind int

const (
	TK_RESERVED TokenKind = iota // Keywords or punctuators
	TK_IDENT                     // Identifier
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
	tokens []*Token
	i      int
}

func isDigit(c byte) bool {
	if '0' <= c && c <= '9' {
		return true
	}
	return false
}

func isLetter(c byte) bool {
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || c == '_' {
		return true
	}
	return false
}

func isAlnum(c byte) bool {
	return isLetter(c) || isDigit(c)
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

func (tn *Tokenizer) tokenize() []*Token {
	for tn.i < len(tn.code) {
		c := tn.peek(1)[0]

		// Add semicolon before newline
		if c == '\n' {
			//行の最後のトークンが以下のいずれかの場合，その後ろにセミコロンが自動的に挿入される．(識別子, 整数リテラル，浮動小数点リテラル， 虚数リテラル，ルーンリテラル，文字列リテラル,break, continue, fallthrough, return,++, --, ), ], } )
			// 改行の前にセミコロンを挿入する。浮動小数点リテラル， 虚数リテラル，ルーンリテラル，文字列リテラルが未実装(TODO)
			tk := tn.tokens[len(tn.tokens)-1] // last token in line
			if tk.kind == TK_NUM || tk.kind == TK_IDENT ||
				tk.val == "break" || tk.val == "continue" || tk.val == "fallthrough" || tk.val == "return" ||
				tk.val == "++" || tk.val == "--" || tk.val == ")" || tk.val == "]" || tk.val == "}" {
				semicolon := &Token{kind: TK_RESERVED, loc: tn.i, val: ";"}
				tn.tokens = append(tn.tokens, semicolon)
			}
			tn.read(1)
			continue
		}

		// Skip whitespace characters.
		if isSpace(c) {
			tn.read(1)
			continue
		}

		// Numeric literal
		if isDigit(c) {
			token := &Token{kind: TK_NUM, loc: tn.i, val: tn.read(1)}
			for isDigit(tn.peek(1)[0]) {
				token.val += tn.read(1)
			}
			tn.tokens = append(tn.tokens, token)
			continue
		}

		// Keywords
		if tn.startswith("return") && !isAlnum(tn.code[tn.i+6]) {
			token := &Token{kind: TK_RESERVED, loc: tn.i, val: tn.read(6)}
			tn.tokens = append(tn.tokens, token)
			continue
		}

		// Multi-letter punctuators
		if tn.startswith("==") || tn.startswith("!=") || tn.startswith("<=") || tn.startswith(">=") {
			token := &Token{kind: TK_RESERVED, loc: tn.i, val: tn.read(2)}
			tn.tokens = append(tn.tokens, token)
			continue
		}

		// Single-letter local variables
		if isLetter(c) {
			token := &Token{kind: TK_IDENT, loc: tn.i, val: tn.read(1)}
			tn.tokens = append(tn.tokens, token)
			continue
		}

		// Single-letter punctuators
		if isPunct(c) {
			token := &Token{kind: TK_RESERVED, loc: tn.i, val: tn.read(1)}
			tn.tokens = append(tn.tokens, token)
			continue
		}

		error_at(tn.code, tn.i, "%sは認識できません", string(c))
	}
	eof_token := &Token{kind: TK_EOF, loc: tn.i}
	tn.tokens = append(tn.tokens, eof_token)

	return tn.tokens
}
