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
	if c == ' ' || c == '\n' || c == '\r' {
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
				error_at(code, i, "引数が数値ではありません(unreachable)")
			}
			token := Token{kind: TK_NUM, val: value, loc: i, len: j}
			tokens = append(tokens, token)
			i += j
			continue
		}

		// Single-letter punctuators
		if isPunct(code[i]) {
			token := Token{kind: TK_RESERVED, label: string(code[i]), loc: i, len: 1}
			tokens = append(tokens, token)
			i++
			continue
		}
		error_at(code, i, "%vは認識できません", string(code[i]))
	}
	eof_token := Token{kind: TK_EOF, loc: i}
	tokens = append(tokens, eof_token)

	return tokens
}

// Parser
type NodeKind int

const (
	ND_ADD NodeKind = iota // +
	ND_SUB                 // -
	ND_MUL                 // *
	ND_DIV                 // /
	ND_NUM                 // Integer
)

type Node struct {
	kind NodeKind // Node kind
	lhs  *Node    // Left-hand side
	rhs  *Node    // Right-hand side
	val  int      // Used if king == ND_NUM
}

// expr = num ("+" num | "-" num)*
// expr = num | expr "+" num | expr "-" num
// num = ( "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ) ( "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" )*
func num(code string, tokens []Token) (*Node, []Token) {
	node := &Node{kind: ND_NUM, val: tokens[0].val}
	return node, tokens[1:]
}

func expr(code string, tokens []Token) (*Node, []Token) {
	var node *Node
	if tokens[0].kind != TK_NUM {
		error_tok(code, tokens[0], "exprが数字で始まっていません")
	}
	node, tokens = num(code, tokens)
	for tokens[0].kind != TK_EOF {
		if tokens[0].label == "+" {
			tokens = tokens[1:]
			if tokens[0].kind != TK_NUM {
				error_tok(code, tokens[0], "+の後が数字ではありません")
			}
			var lhs, rhs *Node
			lhs = node
			rhs, tokens = num(code, tokens)
			node = &Node{kind: ND_ADD, lhs: lhs, rhs: rhs}
			continue
		}
		if tokens[0].label == "-" {
			tokens = tokens[1:]
			if tokens[0].kind != TK_NUM {
				error_tok(code, tokens[0], "-の後が数字ではありません")
			}
			var lhs, rhs *Node
			lhs = node
			rhs, tokens = num(code, tokens)
			node = &Node{kind: ND_SUB, lhs: lhs, rhs: rhs}
			continue
		}
		error_tok(code, tokens[0], "余計なトークンがあります")
	}
	return node, tokens
}

func gen_expr(node *Node) {
	if node.kind == ND_NUM {
		fmt.Printf("  push %d\n", node.val)
		return
	}

	gen_expr(node.lhs)
	gen_expr(node.rhs)
	fmt.Printf("  pop rdi\n")
	fmt.Printf("  pop rax\n")
	switch node.kind {
	case ND_ADD:
		fmt.Printf("  add rax, rdi\n")
	case ND_SUB:
		fmt.Printf("  sub rax, rdi\n")
	default:
		panic("コード生成できません")
	}
	fmt.Printf("  push rax\n")
}

func codegen(code string, node *Node) {
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	gen_expr(node)
	fmt.Printf("  pop rax\n")

	fmt.Printf("  ret\n")
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "引数の個数が正しくありません")
		os.Exit(1)
	}

	code := os.Args[1]
	tokens := tokenize(code)
	node, tokens := expr(code, tokens)
	if len(tokens) == 0 {
		fmt.Fprintln(os.Stderr, "EOFが見つかりません")
	}
	if tokens[0].kind != TK_EOF {
		error_tok(code, tokens[0], "構文解析に失敗しています")
	}

	codegen(code, node)
}
