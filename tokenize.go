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
	line int       // line number
	col  int       // column number
}

type Tokenizer struct {
	code   string
	tokens []*Token
	i      int
	line   int // line number
	col    int // column number
}

func contains(list []string, word string) bool {
	for _, v := range list {
		if v == word {
			return true
		}
	}
	return false
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

func isKeywords(ident string) bool {
	keywords := []string{"return", "if", "else", "for", "func"}
	return contains(keywords, ident)
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
	if result == "\n" {
		tn.line += 1
		tn.col = 1
	} else {
		tn.col += n
	}
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
	tn.line = 1
	tn.col = 1
	for tn.i < len(tn.code) {
		c := tn.peek(1)[0]
		switch {
		case c == '\n' && len(tn.tokens) == 0: // ファイルの先頭が空行
			tn.read(1)
		case c == '\n':
			keywords := []string{"break", "continue", "fallthrough", "return", "++", "--", ")", "]", "}"}
			tk := tn.tokens[len(tn.tokens)-1] // 改行直前のトークン
			// 特定の条件でセミコロンを自動挿入する
			if tk.line == tn.line && (tk.kind == TK_IDENT || tk.kind == TK_NUM || contains(keywords, tk.val)) {
				semicolon := &Token{kind: TK_RESERVED, line: tn.line, col: tn.col + 1, val: ";"}
				tn.tokens = append(tn.tokens, semicolon)
			}
			tn.read(1)
		case isSpace(c): // Skip whitespace characters.
			tn.read(1)
		case isDigit(c): // Numeric literal
			token := &Token{kind: TK_NUM, line: tn.line, col: tn.col}
			token.val = tn.read(1)
			for isDigit(tn.peek(1)[0]) {
				token.val += tn.read(1)
			}
			tn.tokens = append(tn.tokens, token)
		case isLetter(c): // Keywords or local variables
			token := &Token{kind: TK_IDENT, line: tn.line, col: tn.col}
			token.val = tn.read(1)
			for isAlnum(tn.peek(1)[0]) {
				token.val += tn.read(1)
			}
			if isKeywords(token.val) {
				token.kind = TK_RESERVED
			}
			tn.tokens = append(tn.tokens, token)
		case contains([]string{"==", "!=", "<=", ">="}, tn.peek(2)): // Multi-letter punctuators
			token := &Token{kind: TK_RESERVED, line: tn.line, col: tn.col}
			token.val = tn.read(2)
			tn.tokens = append(tn.tokens, token)
		case isPunct(c): // Single-letter punctuators
			token := &Token{kind: TK_RESERVED, line: tn.line, col: tn.col}
			token.val = tn.read(1)
			tn.tokens = append(tn.tokens, token)
		default:
			error_at(tn.code, tn.line, tn.col, "%sは認識できません", string(c))
		}
	}
	eof_token := &Token{kind: TK_EOF, line: tn.line, col: tn.col}
	tn.tokens = append(tn.tokens, eof_token)
	return tn.tokens
}
