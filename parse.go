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
	ND_FUNCCALL                    // Function call
	ND_FUNCDECL                    // Function declaration
	ND_EXPR_STMT                   // Expression statement
	ND_VAR                         // Variable
	ND_NUM                         // Integer
)

type Node struct {
	kind     NodeKind // Node kind
	token    *Token   // Token
	lhs      *Node    // Left-hand side
	rhs      *Node    // Right-hand side
	cond     *Node    // Used if king == ND_IF_STMT or ND_FOR_STMT
	then     *Node    // Used if king == ND_IF_STMT or ND_FOR_STMT
	els      *Node    // Used if king == ND_IF_STMT
	init     *Node    // Used if king == ND_FOR_STMT
	inc      *Node    // Used if king == ND_FOR_STMT
	block    []*Node  // Used if king == ND_BLOCK
	val      string   // Used if king == ND_NUM or ND_VAR or ND_FUNCCALL or ND_FUNCDECL
	args     []*Node  // Used if king == ND_FUNCCALL
	offset   int      // Used if king == ND_VAR or ND_FUNCDECL
	body     *Node    // Used if king == ND_FUNCDECL
	lvar     []*Var   // Used if king == ND_FUNCDECL
	variable *Var     // Used if king == ND_VAR
}

type Var struct {
	name   string
	offset int
}

type Parser struct {
	code   string
	tokens []*Token
	i      int
	scope  []map[string]*Var
	lvar   []*Var
	offset int
}

// Round up `n` to the nearest multiple of `align`. For instance,
// align_to(5, 8) returns 8 and align_to(11, 8) returns 16.
func align_to(n int, align int) int {
	return (n + align - 1) / align * align
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

func (p *Parser) enter_scope() {
	scope := map[string]*Var{}
	p.scope = append([]map[string]*Var{scope}, p.scope...) // スコープを追加
}

func (p *Parser) leave_scope() {
	p.scope = p.scope[1:] // スコープを抜ける
}

// program          = { FunctionDecl ";" } .
func (p *Parser) parse() []*Node {
	var functions []*Node
	p.enter_scope() // ファイルスコープを追加
	for !p.startsWithTokenKind(TK_EOF) {
		fn := p.funcDecl()
		functions = append(functions, fn)
		if !p.startsWithValue(";") {
			error_tok(p.code, p.peek(1)[0], "セミコロンが見つかりません")
		}
		p.read(1) // ";"をスキップ
	}
	p.leave_scope() // ファイルスコープを削除
	return functions
}

// FunctionDecl     = "func" ident Parameters [ "int" ] Block .
// Parameters       = "(" ")" .
func (p *Parser) funcDecl() *Node {
	p.lvar = []*Var{}
	if !p.startsWithValue("func") {
		error_tok(p.code, p.peek(1)[0], "funcが見つかりません")
	}
	p.read(1) // "func"をスキップ
	token := p.read(1)[0]
	p.read(1) // "("をスキップ
	p.read(1) // ")"をスキップ
	if p.startsWithValue("int") {
		p.read(1) // "int"をスキップ
	}
	fn := &Node{kind: ND_FUNCDECL, token: token, val: token.val, body: p.block(), lvar: p.lvar}
	// 変数のオフセット計算
	offset := 0
	for _, variable := range fn.lvar {
		offset += 8
		variable.offset = offset
	}
	// 関数のオフセット計算
	fn.offset = align_to(offset, 16)
	return fn
}

// 以下構文規則
// block         = "{" statementList "}" .
// statementList = { statement ";" } .
func (p *Parser) block() *Node {
	p.enter_scope() // ブロックスコープを追加
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
			p.read(1)       // "}"をスキップ
			p.leave_scope() // ブロックスコープを削除
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
	// 宣言されていないならスコープとローカル変数リストに加える。
	variable := &Var{name: token.val}
	p.scope[0][variable.name] = variable
	p.lvar = append(p.lvar, variable)

	lhs := &Node{kind: ND_VAR, token: token, val: token.val, variable: variable}

	if p.startsWithValue(";") {
		error_tok(p.code, lhs.token, "型名か初期化子が必要です。")
	}

	if p.startsWithValue("int") {
		p.read(1) // "int"をスキップ
	}
	rhs := &Node{kind: ND_NUM, val: "0"} // 宣言のみの場合はゼロ値で初期化
	if p.startsWithValue("=") {
		p.read(1) // "="をスキップ
		rhs = p.expr()
	}
	return &Node{kind: ND_ASSIGN_STMT, lhs: lhs, rhs: rhs}
}

// IfStmt        = "if" Expression Block [ "else" ( IfStmt | Block ) ] .
func (p *Parser) ifStmt() *Node {
	if !p.startsWithValue("if") {
		error_tok(p.code, p.peek(1)[0], "ifが見つかりません")
	}
	p.read(1)       // "if"をスキップ
	p.enter_scope() // ifスコープを追加
	node := &Node{kind: ND_IF_STMT, cond: p.expr(), then: p.block()}
	if p.startsWithValue("else") {
		p.read(1) // "else"をスキップ
		if p.startsWithValue("if") {
			node.els = p.ifStmt()
		} else {
			node.els = p.block()
		}
	}
	p.leave_scope() // ifスコープを削除
	return node
}

// ForStmt       = "for" [ Condition | ForClause ] Block .
func (p *Parser) forStmt() *Node {
	if !p.startsWithValue("for") {
		error_tok(p.code, p.peek(1)[0], "forが見つかりません")
	}
	p.read(1)       // "for"をスキップ
	p.enter_scope() // forスコープを追加

	node := &Node{kind: ND_FOR_STMT}
	if p.startsWithValue("{") {
		// 条件式を省略したパターン
		node.then = p.block()
		p.leave_scope() // forスコープを削除
		return node
	}

	if !p.startsWithValue(";") {
		stmt := p.stmt()
		if p.startsWithValue("{") {
			// 条件式のみのパターン
			node.cond = stmt
			node.then = p.block()
			p.leave_scope() // forスコープを削除
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
	p.leave_scope() // forスコープを削除
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

// primary       = num | ident | funcall | "(" expr ")" .
func (p *Parser) primary() *Node {
	if p.startsWithTokenKind(TK_NUM) {
		return p.num()
	}

	if p.startsWithTokenKind(TK_IDENT) {
		if p.peek(2)[1].val == "(" {
			return p.funcall()
		}
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
		if variable, ok := m[token.val]; ok {
			return &Node{kind: ND_VAR, token: token, val: token.val, variable: variable}
		}
	}
	// 変数がいずれのスコープにも宣言されていないならエラー
	error_tok(p.code, token, "変数が宣言されていません。")
	return nil
}

// funcall = ident "(" [ ExpressionList [ "," ] ] ")" .
// ExpressionList = Expression { "," Expression } .
func (p *Parser) funcall() *Node {
	funcname := p.read(1)[0]
	node := &Node{kind: ND_FUNCCALL, val: funcname.val}
	p.read(1)             // "("をスキップ
	node.args = []*Node{} //ExpressionList
	for !p.startsWithValue(")") {
		node.args = append(node.args, p.expr())
		if p.startsWithValue(",") {
			p.read(1) // ","をスキップ
		}
	}
	if len(node.args) > 6 {
		error_tok(p.code, funcname, "引数が多すぎます(6個以内)")
	}
	p.read(1) // ")"をスキップ
	return node
}
