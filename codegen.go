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
	case ND_EQ:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  sete al\n")
		fmt.Printf("  movzb rax, al\n")
	case ND_NE:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setne al\n")
		fmt.Printf("  movzb rax, al\n")
	case ND_LT:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setl al\n")
		fmt.Printf("  movzb rax, al\n")
	case ND_LE:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setle al\n")
		fmt.Printf("  movzb rax, al\n")
	default:
		panic("コード生成できません")
	}
	fmt.Printf("  push rax\n")
}
func gen_stmt(node *Node) {
	switch node.kind {
	case ND_RETURN:
		gen_expr(node.lhs)
		fmt.Printf("  pop rax\n")
		fmt.Printf("  jmp .L.return\n")
		return
	case ND_EXPR_STMT:
		gen_expr(node.lhs)
		fmt.Printf("  pop rax\n") // スタックの値を捨てる
	default:
		panic("コード生成できません")
	}
}

func codegen(nodes []*Node) {
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	for _, node := range nodes {
		gen_stmt(node)
	}

	fmt.Printf(".L.return:\n")
	fmt.Printf("  ret\n")
}
