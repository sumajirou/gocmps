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
type Parser struct {
	code   string
	tokens []Token
	i      int
}

func (p *Parser) peek(n int) []Token {
	if p.i+n > len(p.tokens) {
		return p.tokens[p.i:len(p.tokens)]
	}
	return p.tokens[p.i : p.i+n]
}

func (p *Parser) read(n int) []Token {
	if p.i+n > len(p.tokens) {
		result := p.tokens[p.i:len(p.tokens)]
		p.i = len(p.tokens)
		return result
	}
	result := p.tokens[p.i : p.i+n]
	p.i += n
	return result
}

func (p *Parser) startsWithTokenKind(kind TokenKind) bool {
	return p.tokens[p.i].kind == kind
}

func (p *Parser) startsWithValue(s string) bool {
	return p.tokens[p.i].val == s
}

func (p *Parser) parse() *Node {
	return p.expr()
}

// 以下構文規則

func (p *Parser) num() *Node {
	return &Node{kind: ND_NUM, val: p.read(1)[0].val}
}

// expr = mul | expr "+" mul | expr "-" mul
func (p *Parser) expr() *Node {
	node := p.mul()
	for {
		if p.startsWithValue("+") {
			p.read(1)
			node = &Node{kind: ND_ADD, lhs: node, rhs: p.mul()}
			continue
		}
		if p.startsWithValue("-") {
			p.read(1)
			node = &Node{kind: ND_SUB, lhs: node, rhs: p.mul()}
			continue
		}
		return node
	}
}

// mul  = primary | mul "*" primary | mul "/" primary
func (p *Parser) mul() *Node {
	node := p.primary()
	for {
		if p.startsWithValue("*") {
			p.read(1)
			node = &Node{kind: ND_MUL, lhs: node, rhs: p.primary()}
			continue
		}
		if p.startsWithValue("/") {
			p.read(1)
			node = &Node{kind: ND_DIV, lhs: node, rhs: p.primary()}
			continue
		}
		return node
	}
}

// primary = num | "(" expr ")"
func (p *Parser) primary() *Node {
	if p.startsWithTokenKind(TK_NUM) {
		return p.num()
	}

	if !p.startsWithValue("(") {
		error_tok(p.code, p.peek(1)[0], "不正なトークンです")
	}
	p.read(1)

	node := p.expr()

	if !p.startsWithValue(")") {
		error_tok(p.code, p.peek(1)[0], "括弧が閉じられていません")
	}
	p.read(1)

	return node
}
