package main

import "fmt"

type Codegen struct {
	code    string
	program *Node
}

var counter int = 0

func count() int {
	counter++
	return counter
}

func (cg *Codegen) gen_lval(node *Node) {
	if node.kind != ND_VAR {
		error_tok(cg.code, node.token, "左辺値が変数ではありません")
	}
	fmt.Printf("  mov rax, rbp\n")
	fmt.Printf("  sub rax, %d\n", node.offset)
	fmt.Printf("  push rax\n") // 変数のアドレスをスタックに積む
}

func (cg *Codegen) gen_expr(node *Node) {
	switch node.kind {
	case ND_NUM:
		fmt.Printf("  push %s\n", node.val) // 整数リテラルをスタックに積む
		return
	case ND_VAR:
		cg.gen_lval(node)            // 変数のアドレスをスタックに積む
		fmt.Printf("  pop rax\n")    // 変数のアドレスをポップ
		fmt.Printf("  push [rax]\n") // 変数の値をスタックに積む
		return
	}

	cg.gen_expr(node.lhs)
	cg.gen_expr(node.rhs)
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
	fmt.Printf("  push rax\n") // 計算した値をスタックに積む
}

func (cg *Codegen) gen_stmt(node *Node) {
	switch node.kind {
	case ND_RETURN_STMT:
		cg.gen_expr(node.lhs)           // 式の値を計算してスタックに積み
		fmt.Printf("  pop rax\n")       // スタックからraxにポップし
		fmt.Printf("  jmp .L.return\n") // リターンする
	case ND_IF_STMT:
		c := count()
		cg.gen_expr(node.cond)              // condを計算してスタックに積み
		fmt.Printf("  pop rax\n")           // スタックからraxにポップし
		fmt.Printf("  cmp rax, 0\n")        // 比較
		fmt.Printf("  je  .L.else.%d\n", c) // condがfalseなら対応する.L.elseにジャンプ
		cg.gen_stmt(node.then)              // then節を実行
		fmt.Printf("  jmp .L.end.%d\n", c)  // 対応する.L.endにジャンプ
		fmt.Printf(".L.else.%d:\n", c)
		if node.els != nil {
			cg.gen_stmt(node.els) // els節があれば実行
		}
		fmt.Printf(".L.end.%d:\n", c)
	case ND_FOR_STMT:
		c := count()
		if node.init != nil {
			cg.gen_stmt(node.init) // init節があれば実行
		}
		fmt.Printf(".L.begin.%d:\n", c)
		if node.cond != nil {
			cg.gen_expr(node.cond)             // cond節があれば実行
			fmt.Printf("  pop rax\n")          // スタックからraxにポップし
			fmt.Printf("  cmp rax, 0\n")       // 比較
			fmt.Printf("  je  .L.end.%d\n", c) // condがfalseなら対応する.L.endにジャンプ
		}
		cg.gen_stmt(node.then)
		if node.inc != nil {
			cg.gen_stmt(node.inc) // inc節があれば実行
		}
		fmt.Printf("  jmp .L.begin.%d\n", c)
		fmt.Printf(".L.end.%d:\n", c)
	case ND_BLOCK:
		for _, stmt := range node.block {
			cg.gen_stmt(stmt) // 文を逐次実行
		}
	case ND_EXPR_STMT:
		cg.gen_expr(node.lhs)     // 式の値を計算してスタックに積み
		fmt.Printf("  pop rax\n") // スタックの値を捨てる
	case ND_ASSIGN_STMT:
		cg.gen_lval(node.lhs)            // 左辺のアドレスを計算してスタックに積み
		cg.gen_expr(node.rhs)            // 右辺の式の値を計算してスタックに積み
		fmt.Printf("  pop rdi\n")        // 式の値をrdiにポップし
		fmt.Printf("  pop rax\n")        // 変数のアドレスをraxにポップし
		fmt.Printf("  mov [rax], rdi\n") // 変数に値を代入
	default:
		panic("コード生成できません")
	}
}

func (cg *Codegen) codegen() {
	fmt.Printf(".intel_syntax noprefix\n") //Intel記法
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	// プロローグ
	// 変数の領域を確保する
	fmt.Printf("  push rbp\n")
	fmt.Printf("  mov rbp, rsp\n")
	fmt.Printf("  sub rsp, %d\n", cg.program.offset)

	for _, node := range cg.program.block {
		cg.gen_stmt(node) // 文を逐次実行
	}

	fmt.Printf(".L.return:\n")

	// エピローグ
	// 最後の式の結果がRAXに残っているのでそれが返り値になる
	fmt.Printf("  mov rsp, rbp\n")
	fmt.Printf("  pop rbp\n")
	fmt.Printf("  ret\n")
}
