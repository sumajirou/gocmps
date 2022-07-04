package main

import (
	"fmt"
	"os"
	"strings"
)

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

func error_tok(code string, token *Token, fmtstr string, ap ...any) {
	verror_at(code, token.loc, fmtstr, ap...)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "引数の個数が正しくありません")
		os.Exit(1)
	}

	code := os.Args[1] + "\n" // ソースコードの終端が改行文字であることを保証
	tokenizer := Tokenizer{code: code}
	tokens := tokenizer.tokenize()
	parser := Parser{code: code, tokens: tokens, lVar: tokenizer.lVar}
	nodes := parser.parse()
	codegen := Codegen{code: code, nodes: nodes, lVar: parser.lVar}
	codegen.codegen()
}
