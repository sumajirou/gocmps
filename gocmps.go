package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// TokenKind is kind of token
type TokenKind int

const (
	// Reserved is Keyword or punctuators
	Reserved TokenKind = iota
	// Num is Integer literals
	Num
	// EOF is End-of-file markers
	EOF
)

func (tk TokenKind) String() string {
	switch tk {
	case Reserved:
		return "Reserved"
	case Num:
		return "Num"
	case EOF:
		return "EOF"
	default:
		return "Unknown"
	}
}

// Token is struct
type Token struct {
	Kind TokenKind // Token kind
	Val  int       // Int Value
	Name string    // Token String
}

func tokenize(s string) []Token {
	var result []Token

	// 数値 | + | - | それ以外
	re, err := regexp.Compile("\\d+|[+-]|[^ \\d+-]+")
	if err != nil {
		panic("正規表現が不正です。")
	}
	for _, v := range re.FindAllStringSubmatch(s, -1) {
		var token Token
		if v[0] == "+" || v[0] == "-" {
			token.Kind = Reserved
			token.Name = v[0]
		} else if val, err := strconv.Atoi(v[0]); err == nil {
			token.Kind = Reserved
			token.Name = v[0]
			token.Val = val
		} else {
			panic(fmt.Sprintf("認識できないトークンです。: %s", v[0]))
		}
		result = append(result, token)
	}

	eof := Token{EOF, 0, ""}
	result = append(result, eof)
	return result
}

func main() {
	if len(os.Args) != 2 {
		panic("引数の個数が正しくありません")
	}

	tok := tokenize(os.Args[1])

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	// The first token must be a number
	fmt.Printf("  mov rax, %d\n", tok[0].Val)
	tok = tok[1:]
	for tok[0].Kind != EOF {
		if tok[0].Name == "+" {
			fmt.Printf("  add rax, %d\n", tok[1].Val)
			tok = tok[2:]
			continue
		} else if tok[0].Name == "-" {
			fmt.Printf("  sub rax, %d\n", tok[1].Val)
			tok = tok[2:]
			continue
		} else {
			panic(fmt.Sprintf("不正なトークンです。: %s", tok[1].Name))
		}
	}
	fmt.Printf("  ret\n")
}
