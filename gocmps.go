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

func tokenize(code string) []Token {
	if len(code) == 0 {
		panic("コードが空文字列")
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

		// Single-letter punctuators
		if isPunct(code[i]) {
			if code[i] == '+' || code[i] == '-' {
				token := Token{kind: TK_RESERVED, label: string(code[i]), loc: i, len: 1}
				tokens = append(tokens, token)
			} else {
				panic("認識できません")
			}
			i++
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

	i := 0
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	fmt.Printf("  mov rax, %d\n", tokens[i].val)
	i++
	for tokens[i].kind != TK_EOF {
		if tokens[i].kind == TK_RESERVED && tokens[i].label == "+" {
			i++
			if tokens[i].kind != TK_NUM {
				panic("+の後が数字じゃない")
			}
			fmt.Printf("  add rax, %d\n", tokens[i].val)
			i++
			continue
		}

		if tokens[i].kind == TK_RESERVED && tokens[i].label == "-" {
			i++
			if tokens[i].kind != TK_NUM {
				panic("-の後が数字じゃない")
			}
			fmt.Printf("  sub rax, %d\n", tokens[i].val)
			i++
			continue
		}
		panic("不正なトークン")
	}

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
