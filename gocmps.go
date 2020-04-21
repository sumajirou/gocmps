package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		// panicの使い方これで正しいのかな……？
		panic("引数の個数が正しくありません")
	}

	value, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic("引数が数値ではありません。")
	}

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	fmt.Printf("  mov rax, %d\n", value)
	fmt.Printf("  ret\n")
}
