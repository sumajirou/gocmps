package main

type NodeKind int

const (
	ND_ADD         NodeKind = iota // +
	ND_SUB                         // -
	ND_MUL                         // *
	ND_DIV                         // /
	ND_EQ                          // ==
	ND_NE                          // !=
	ND_LT                          // <
	ND_LE                          // <=
	ND_ASSIGN_STMT                 // =
	ND_RETURN_STMT                 // "return"
	ND_BLOCK                       // "{ ... }"
	ND_EXPR_STMT                   // Expression statement
	ND_VAR                         // Variable
	ND_NUM                         // Integer
)

type Node struct {
	kind   NodeKind // Node kind
	token  *Token   // Token
	lhs    *Node    // Left-hand side
	rhs    *Node    // Right-hand side
	block  []*Node  // Used if king == ND_BLOCK
	val    string   // Used if king == ND_NUM
	offset int      // Used if king == ND_VAR
}

type Parser struct {
	code   string
	tokens []*Token
	i      int
	lVar   map[string]int
}

func (p *Parser) peek(n int) []*Token {
	if p.i+n > len(p.tokens) {
		return p.tokens[p.i:len(p.tokens)]
	}
	return p.tokens[p.i : p.i+n]
}

func (p *Parser) read(n int) []*Token {
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

// program       = block ";" .
func (p *Parser) parse() *Node {
	node := p.block()
	if !p.startsWithValue(";") {
		error_tok(p.code, p.peek(1)[0], "セミコロンが見つかりません")
	}
	p.read(1) // ";"をスキップ

	if p.i == len(p.tokens) {
		error_at(p.code, p.i, "EOFが見つかりません")
	}
	if !p.startsWithTokenKind(TK_EOF) {
		error_tok(p.code, p.peek(1)[0], "EOFの前にトークンが残っています")
	}
	return node
}

// 以下構文規則
// block         = "{" statementList "}" .
// statementList = { statement ";" } .
func (p *Parser) block() *Node {
	if !p.startsWithValue("{") {
		error_tok(p.code, p.peek(1)[0], "{が見つかりません")
	}
	p.read(1) // "{"をスキップ

	node := &Node{kind: ND_BLOCK, block: []*Node{}}
	for {
		if p.startsWithValue(";") {
			p.read(1) // ";"をスキップ
		}
		if p.startsWithValue("}") {
			p.read(1) // "}"をスキップ
			return node
		}
		node.block = append(node.block, p.stmt())
		if !p.startsWithValue(";") && !p.startsWithValue("}") {
			error_tok(p.code, p.peek(1)[0], "セミコロンが見つかりません")
		}
	}
}

// stmt = "return" expr | block | assignStmt .
func (p *Parser) stmt() *Node {
	// return statement
	if p.startsWithValue("return") {
		p.read(1) // "returnをスキップ"
		return &Node{kind: ND_RETURN_STMT, lhs: p.expr()}
	}
	// block
	if p.startsWithValue("{") {
		return p.block()
	}
	// assign statement
	return p.assignStmt()
}

// assignStmt = expr [ "=" expr ].
func (p *Parser) assignStmt() *Node {
	lhs := p.expr()
	if p.startsWithValue("=") {
		p.read(1) // "="をスキップ
		rhs := p.expr()
		return &Node{kind: ND_ASSIGN_STMT, lhs: lhs, rhs: rhs}
	} else {
		return &Node{kind: ND_EXPR_STMT, lhs: lhs}
	}
}

// expr = add { "==" add | "!=" add | "<" add | "<=" add | ">" add | ">=" add } .
func (p *Parser) expr() *Node {
	node := p.add()
	for {
		if p.startsWithValue("==") {
			p.read(1)
			node = &Node{kind: ND_EQ, lhs: node, rhs: p.add()}
			continue
		}
		if p.startsWithValue("!=") {
			p.read(1)
			node = &Node{kind: ND_NE, lhs: node, rhs: p.add()}
			continue
		}
		if p.startsWithValue("<") {
			p.read(1)
			node = &Node{kind: ND_LT, lhs: node, rhs: p.add()}
			continue
		}
		if p.startsWithValue("<=") {
			p.read(1)
			node = &Node{kind: ND_LE, lhs: node, rhs: p.add()}
			continue
		}
		if p.startsWithValue(">") {
			p.read(1)
			node = &Node{kind: ND_LT, lhs: p.add(), rhs: node}
			continue
		}
		if p.startsWithValue(">=") {
			p.read(1)
			node = &Node{kind: ND_LE, lhs: p.add(), rhs: node}
			continue
		}
		return node
	}
}

// add = mul { "+" mul | "-" mul } .
func (p *Parser) add() *Node {
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

// mul = unary { "*" unary | "/" unary } .
func (p *Parser) mul() *Node {
	node := p.unary()
	for {
		if p.startsWithValue("*") {
			p.read(1)
			node = &Node{kind: ND_MUL, lhs: node, rhs: p.unary()}
			continue
		}
		if p.startsWithValue("/") {
			p.read(1)
			node = &Node{kind: ND_DIV, lhs: node, rhs: p.unary()}
			continue
		}
		return node
	}
}

// unary = primary | [ "+" | "-" ] unary .
func (p *Parser) unary() *Node {
	if p.startsWithValue("+") {
		p.read(1)
		return p.unary()
	}
	if p.startsWithValue("-") {
		p.read(1)
		zero := &Node{kind: ND_NUM, val: "0"}
		return &Node{kind: ND_SUB, lhs: zero, rhs: p.unary()}
	}
	return p.primary()
}

// primary = num | ident | "(" expr ")" .
func (p *Parser) primary() *Node {
	if p.startsWithTokenKind(TK_NUM) {
		return p.num()
	}

	if p.startsWithTokenKind(TK_IDENT) {
		return p.ident()
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

// num = digit { digit } .
func (p *Parser) num() *Node {
	token := p.read(1)[0]
	return &Node{kind: ND_NUM, token: token, val: token.val}
}

// ident = letter { alnum } .
func (p *Parser) ident() *Node {
	token := p.read(1)[0]
	offset := p.lVar[token.val]
	return &Node{kind: ND_VAR, token: token, val: token.val, offset: offset}
}
