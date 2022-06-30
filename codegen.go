package main

import "fmt"

func gen_expr(node *Node) {
	if node.kind == ND_NUM {
		fmt.Printf("  push %s\n", node.val)
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
	case ND_MUL:
		fmt.Printf("  imul rax, rdi\n")
	case ND_DIV:
		fmt.Printf("  cqo\n")
		fmt.Printf("  idiv rdi\n")
	default:
		panic("コード生成できません")
	}
	fmt.Printf("  push rax\n")
}

func codegen(node *Node) {
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	gen_expr(node)
	fmt.Printf("  pop rax\n")

	fmt.Printf("  ret\n")
}
