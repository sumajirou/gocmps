package main

import (
	"fmt"
	"os"
	"regexp"
)

func main() {
	if len(os.Args) != 2 {
		panic("引数の個数が正しくありません")
	}

	s := os.Args[1]
	// 数値 | + | - | それ以外
	re, err := regexp.Compile("\\d+|[+-]|[^\\d+-]+")
	if err != nil {
		panic("正規表現が不正です。")
	}

	token := []string{}
	for _, v := range re.FindAllStringSubmatch(s, -1) {
		token = append(token, v[0])
	}

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	fmt.Printf("  mov rax, %s\n", token[0])
	token = token[1:]
	for len(token) > 0 {
		if token[0] == "+" {
			fmt.Printf("  add rax, %s\n", token[1])
		} else if token[0] == "-" {
			fmt.Printf("  sub rax, %s\n", token[1])
		} else {
			panic(fmt.Sprintf("不正なトークンです。: %s", token[1]))
		}
		token = token[2:]
	}
	fmt.Printf("  ret\n")
}
