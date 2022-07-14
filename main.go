package main

import (
	"fmt"
	"os"
	"strings"
)

func verror_at(code string, line int, col int, fmtstr string, ap ...any) {
	// TODO: ファイル名とスコープ ファイル名と行数と文字数エラー内容 エラー位置表示
	lines := strings.Split(code, "\n")
	fmt.Fprintln(os.Stderr, lines[line-1])
	fmt.Fprint(os.Stderr, strings.Repeat(" ", col-1)+"^ ")
	fmt.Fprintf(os.Stderr, "[%d:%d] ", line, col)
	fmt.Fprintf(os.Stderr, fmtstr+"\n", ap...)
	os.Exit(1)
}

func error_at(code string, line int, col int, fmtstr string, ap ...any) {
	verror_at(code, line, col, fmtstr, ap...)
}

func error_tok(code string, token *Token, fmtstr string, ap ...any) {
	verror_at(code, token.line, token.col, fmtstr, ap...)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "引数の個数が正しくありません")
		os.Exit(1)
	}

	code := os.Args[1]
	if code[len(code)-1] != '\n' {
		code += "\n" // ソースコードの終端が改行文字であることを保証
	}
	tokenizer := Tokenizer{code: code}
	tokens := tokenizer.tokenize()
	parser := Parser{code: code, tokens: tokens}
	program := parser.parse()
	codegen := Codegen{code: code, program: program}
	codegen.codegen()
}
