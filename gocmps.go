package main

import (
	"fmt"
	"os"
	"strconv"
)

type TokenKind int

const (
	TK_RESERVED TokenKind = iota // Keywords or punctuators
	TK_NUM                       // Numeric literals
	TK_EOF                       // End-of-file markers
)

type Token struct {
	kind TokenKind // Token kind
	val  int       // If kind is TK_NUM, its value
	loc  int       // Token location
	len  int       // Token length
}

func isDigit(c byte) bool {
	if '0' <= c && c <= '9' {
		return true
	}
	return false
}

func tokenize(code string) []Token {
	if len(code) == 0 {
		panic("コードが空文字列")
	}

	tokens := []Token{}
	i := 0 // codeのインデックス
	for i < len(code) {
		// Numeric literal
		if isDigit(code[i]) {
			token := Token{kind: TK_NUM, loc: i}
			j := 0
			for ; i+j < len(code) && isDigit(code[i+j]); j++ {
			}
			value, err := strconv.Atoi(code[i : i+j])
			if err != nil {
				panic("引数が数値ではありません。")
			}
			token.val = value
			token.len = j
			tokens = append(tokens, token)
			i += j
			continue
		}
		panic("不正な文字")
	}
	eof_token := Token{kind: TK_EOF, loc: i}
	tokens = append(tokens, eof_token)

	return tokens
}

func codegen(tokens []Token) {
	if tokens[0].kind != TK_NUM {
		panic("数字で始まっていない")
	}

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	fmt.Printf("  mov rax, %d\n", tokens[0].val)
	fmt.Printf("  ret\n")
}

func main() {
	if len(os.Args) != 2 {
		panic("引数の個数が正しくありません")
	}

	code := os.Args[1]
	tokens := tokenize(code)
	codegen(tokens)
}
