package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type TokenKind int

const (
	TK_RESERVED TokenKind = iota // Keywords or punctuators
	TK_NUM                       // Numeric literals
	TK_EOF                       // End-of-file markers
)

type Token struct {
	kind  TokenKind // Token kind
	label string    // If kind is TK_RESERVED, its value
	val   int       // If kind is TK_NUM, its value
	loc   int       // Token location
	len   int       // Token length
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
	if c == ' ' || c == '	' {
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

func verror_at(code string, loc int, fmtstr string, ap ...any) {
	// TODO: ファイル名とスコープ ファイル名と行数と文字数エラー内容 エラー位置表示
	fmt.Fprintln(os.Stderr, code)
	fmt.Fprint(os.Stderr, strings.Repeat(" ", loc)+"^ ")
	fmt.Fprintf(os.Stderr, fmtstr+"\n", ap...)
	os.Exit(1)
}

func error_at(code string, loc int, fmtstr string, ap ...any) {
	verror_at(code, loc, fmtstr, ap...)
}

func error_tok(code string, token Token, fmtstr string, ap ...any) {
	verror_at(code, token.loc, fmtstr, ap...)
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
			value, err := strconv.Atoi(code[i : i+j])
			if err != nil {
				error_at(code, i, "引数が数値ではありません")
			}
			token := Token{kind: TK_NUM, val: value, loc: i, len: j}
			tokens = append(tokens, token)
			i += j
			continue
		}

		// Single-letter punctuators
		if isPunct(code[i]) {
			if code[i] == '+' || code[i] == '-' {
				token := Token{kind: TK_RESERVED, label: string(code[i]), loc: i, len: 1}
				tokens = append(tokens, token)
			} else {
				error_at(code, i, "%vは認識できません", string(code[i]))
			}
			i++
			continue
		}
		error_at(code, i, "%vは認識できません", string(code[i]))
	}
	eof_token := Token{kind: TK_EOF, loc: i}
	tokens = append(tokens, eof_token)

	return tokens
}

func codegen(code string, tokens []Token) {
	i := 0
	if tokens[i].kind != TK_NUM {
		error_tok(code, tokens[i], "数字で始まっていません")
	}

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	fmt.Printf("  mov rax, %d\n", tokens[i].val)
	i++
	for tokens[i].kind != TK_EOF {
		if tokens[i].kind == TK_RESERVED && tokens[i].label == "+" {
			i++
			if tokens[i].kind != TK_NUM {
				error_tok(code, tokens[i], "+の後が数字ではありません")
			}
			fmt.Printf("  add rax, %d\n", tokens[i].val)
			i++
			continue
		}

		if tokens[i].kind == TK_RESERVED && tokens[i].label == "-" {
			i++
			if tokens[i].kind != TK_NUM {
				error_tok(code, tokens[i], "-の後が数字ではありません")
			}
			fmt.Printf("  sub rax, %d\n", tokens[i].val)
			i++
			continue
		}
		error_tok(code, tokens[i], "不正なトークンです")
	}

	fmt.Printf("  ret\n")
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "引数の個数が正しくありません")
		os.Exit(1)
	}

	code := os.Args[1]
	tokens := tokenize(code)
	codegen(code, tokens)
}
