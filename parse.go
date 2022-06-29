package main

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
	val  string   // Used if king == ND_NUM
}

func num(code string, tokens []Token) (*Node, []Token) {
	node := &Node{kind: ND_NUM, val: tokens[0].val}
	return node, tokens[1:]
}

// expr = mul | expr "+" mul | expr "-" mul
func expr(code string, tokens []Token) (*Node, []Token) {
	var node *Node
	node, tokens = mul(code, tokens)
	for {
		if tokens[0].val == "+" {
			tokens = tokens[1:]
			var lhs, rhs *Node
			lhs = node
			rhs, tokens = mul(code, tokens)
			node = &Node{kind: ND_ADD, lhs: lhs, rhs: rhs}
			continue
		}
		if tokens[0].val == "-" {
			tokens = tokens[1:]
			var lhs, rhs *Node
			lhs = node
			rhs, tokens = mul(code, tokens)
			node = &Node{kind: ND_SUB, lhs: lhs, rhs: rhs}
			continue
		}
		return node, tokens
	}
}

// mul  = primary | mul "*" primary | mul "/" primary
func mul(code string, tokens []Token) (*Node, []Token) {
	var node *Node
	node, tokens = primary(code, tokens)
	for {
		if tokens[0].val == "*" {
			tokens = tokens[1:]
			var lhs, rhs *Node
			lhs = node
			rhs, tokens = primary(code, tokens)
			node = &Node{kind: ND_MUL, lhs: lhs, rhs: rhs}
			continue
		}
		if tokens[0].val == "/" {
			tokens = tokens[1:]
			var lhs, rhs *Node
			lhs = node
			rhs, tokens = primary(code, tokens)
			node = &Node{kind: ND_DIV, lhs: lhs, rhs: rhs}
			continue
		}
		return node, tokens
	}
}

// primary = num | "(" expr ")"
func primary(code string, tokens []Token) (*Node, []Token) {
	if tokens[0].kind == TK_NUM {
		return num(code, tokens)
	}
	if tokens[0].val != "(" {
		error_tok(code, tokens[0], "不正なトークンです")
	}
	tokens = tokens[1:]
	var node *Node
	node, tokens = expr(code, tokens)
	if tokens[0].val != ")" {
		error_tok(code, tokens[0], "括弧が閉じられていません")
	}
	tokens = tokens[1:]
	return node, tokens
}
