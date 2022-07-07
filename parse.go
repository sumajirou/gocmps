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
	ND_IF_STMT                     // "if"
	ND_FOR_STMT                    // "for"
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
	cond   *Node    // Used if king == ND_IF_STMT or ND_FOR_STMT
	then   *Node    // Used if king == ND_IF_STMT or ND_FOR_STMT
	els    *Node    // Used if king == ND_IF_STMT
	init   *Node    // Used if king == ND_FOR_STMT
	inc    *Node    // Used if king == ND_FOR_STMT
	block  []*Node  // Used if king == ND_BLOCK
	val    string   // Used if king == ND_NUM or ND_VAR
	offset int      // Used if king == ND_VAR
}

type Parser struct {
	code   string
	tokens []*Token
	i      int
	scope  []lVar
	offset int
}
type lVar map[string]int

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
	p.offset = 0
	p.scope = []lVar{map[string]int{}} // ファイルスコープを追加
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
	node.offset = p.offset
	return node
}

// 以下構文規則
// block         = "{" statementList "}" .
// statementList = { statement ";" } .
func (p *Parser) block() *Node {
	p.scope = append([]lVar{map[string]int{}}, p.scope...) // ブロックスコープを追加
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
			p.read(1)             // "}"をスキップ
			p.scope = p.scope[1:] // ブロックスコープを削除
			return node
		}
		node.block = append(node.block, p.stmt())
		if !p.startsWithValue(";") && !p.startsWithValue("}") {
			error_tok(p.code, p.peek(1)[0], "セミコロンが見つかりません")
		}
	}
}

// statement     = "return" expr | VarDecl | IfStmt | ForStmt | block | assignStmt .
func (p *Parser) stmt() *Node {
	// return statement
	if p.startsWithValue("return") {
		p.read(1) // "returnをスキップ"
		return &Node{kind: ND_RETURN_STMT, lhs: p.expr()}
	}
	// VarDecl
	if p.startsWithValue("var") {
		return p.varDecl()
	}
	// IfStmt
	if p.startsWithValue("if") {
		return p.ifStmt()
	}
	// ForStmt
	if p.startsWithValue("for") {
		return p.forStmt()
	}
	// block
	if p.startsWithValue("{") {
		return p.block()
	}
	// assign statement
	return p.assignStmt()
}

// VarDecl       = "var" ident ( "int" [ "=" expr ] | "=" expr ) .
func (p *Parser) varDecl() *Node {
	if !p.startsWithValue("var") {
		error_tok(p.code, p.peek(1)[0], "varが見つかりません")
	}
	p.read(1) // "var"をスキップ

	token := p.read(1)[0]
	if _, ok := p.scope[0][token.val]; ok {
		// 変数が現在のスコープで宣言済みなのでエラー
		error_tok(p.code, token, "変数は宣言済みです。")
	}
	// 宣言されていないならスコープに加える。
	p.offset += 8
	p.scope[0][token.val] = p.offset
	lhs := &Node{kind: ND_VAR, token: token, val: token.val, offset: p.offset}

	if p.startsWithValue("=") {
		p.read(1) // "="をスキップ
		rhs := p.expr()
		return &Node{kind: ND_ASSIGN_STMT, lhs: lhs, rhs: rhs}
	}

	if p.startsWithValue("int") {
		p.read(1) // "int"をスキップ
		if p.startsWithValue("=") {
			p.read(1) // "="をスキップ
			rhs := p.expr()
			return &Node{kind: ND_ASSIGN_STMT, lhs: lhs, rhs: rhs}
		}
		// 宣言のみなのでゼロ値で初期化
		rhs := &Node{kind: ND_NUM, val: "0"}
		return &Node{kind: ND_ASSIGN_STMT, lhs: lhs, rhs: rhs}
	}
	error_tok(p.code, lhs.token, "型名か初期化子が必要です。")
	return nil
}

// IfStmt        = "if" Expression Block [ "else" ( IfStmt | Block ) ] .
func (p *Parser) ifStmt() *Node {
	if !p.startsWithValue("if") {
		error_tok(p.code, p.peek(1)[0], "ifが見つかりません")
	}
	p.read(1)                                              // "if"をスキップ
	p.scope = append([]lVar{map[string]int{}}, p.scope...) // ifスコープを追加
	node := &Node{kind: ND_IF_STMT, cond: p.expr(), then: p.block()}
	if p.startsWithValue("else") {
		p.read(1) // "else"をスキップ
		if p.startsWithValue("if") {
			node.els = p.ifStmt()
		} else {
			node.els = p.block()
		}
	}
	p.scope = p.scope[1:] // ifスコープを削除
	return node
}

// ForStmt       = "for" [ Condition | ForClause ] Block .
func (p *Parser) forStmt() *Node {
	if !p.startsWithValue("for") {
		error_tok(p.code, p.peek(1)[0], "forが見つかりません")
	}
	p.read(1)                                              // "for"をスキップ
	p.scope = append([]lVar{map[string]int{}}, p.scope...) // forスコープを追加

	node := &Node{kind: ND_FOR_STMT}
	if p.startsWithValue("{") {
		// 条件式を省略したパターン
		node.then = p.block()
		p.scope = p.scope[1:] // forスコープを削除
		return node
	}

	if !p.startsWithValue(";") {
		stmt := p.stmt()
		if p.startsWithValue("{") {
			// 条件式のみのパターン
			node.cond = stmt
			node.then = p.block()
			p.scope = p.scope[1:] // forスコープを削除
			return node
		}
		node.init = stmt
	}
	// for句を用いるパターン
	p.read(1) // 最初のセミコロンをスキップ

	if !p.startsWithValue(";") {
		node.cond = p.expr()
	}
	p.read(1) // 2つ目のセミコロンをスキップ

	if !p.startsWithValue("{") {
		node.inc = p.stmt()
	}
	node.then = p.block()
	p.scope = p.scope[1:] // forスコープを削除
	return node
}

// assignStmt = expr [ "=" expr ].
func (p *Parser) assignStmt() *Node {
	lhs := p.expr()
	if p.startsWithValue("=") {
		if lhs.kind != ND_VAR {
			error_tok(p.code, lhs.token, "左辺が変数ではありません")
		}
		p.read(1) // "="をスキップ
		rhs := p.expr()
		return &Node{kind: ND_ASSIGN_STMT, lhs: lhs, rhs: rhs}
	}
	return &Node{kind: ND_EXPR_STMT, lhs: lhs}
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
	for _, m := range p.scope {
		if offset, ok := m[token.val]; ok {
			return &Node{kind: ND_VAR, token: token, val: token.val, offset: offset}
		}
	}
	// 変数がいずれのスコープにも宣言されていないならエラー
	error_tok(p.code, token, "変数が宣言されていません。")
	return nil
}
