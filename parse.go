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
	params   []*Var   // Used if king == ND_FUNCCALL
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

func (p *Parser) consume(s string) *Token {
	if !p.startsWithValue(s) {
		error_tok(p.code, p.peek(1)[0], "%sが見つかりません", s)
	}
	return p.read(1)[0]
}

func (p *Parser) consumeWithTokenKind(kind TokenKind) *Token {
	if !p.startsWithTokenKind(kind) {
		error_tok(p.code, p.peek(1)[0], "%sは不正なトークン種別です", kind)
	}
	return p.read(1)[0]
}

func (p *Parser) consumeIfPossible(s string) *Token {
	if p.startsWithValue(s) {
		return p.read(1)[0]
	}
	return nil
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
		p.consume(";")
	}
	p.leave_scope() // ファイルスコープを削除
	return functions
}

// FunctionDecl     = "func" ident Parameters [ "int" ] Block .
// Parameters       = "(" [ ident "int" { "," ident "int" } [ "," ] ] ")" .
func (p *Parser) funcDecl() *Node {
	p.enter_scope()    // スコープを追加
	p.lvar = []*Var{}  // 関数のローカル変数のリスト
	params := []*Var{} // 仮引数のリスト

	p.consume("func")
	funcname := p.consumeWithTokenKind(TK_IDENT) // 関数名
	p.consume("(")
	for !p.startsWithValue(")") {
		param := p.consumeWithTokenKind(TK_IDENT) // 仮引数名
		if _, ok := p.scope[0][param.val]; ok {
			error_tok(p.code, param, "仮引数名が重複しています")
		}
		variable := &Var{name: param.val}    // 仮引数
		params = append(params, variable)    // 変数リストに仮引数を追加
		p.scope[0][variable.name] = variable // 現在のスコープに仮引数を追加
		p.consume("int")                     // "int"をスキップ
		if !p.startsWithValue(",") && !p.startsWithValue(")") {
			error_tok(p.code, p.peek(1)[0], "不正なトークン")
		}
		p.consumeIfPossible(",") // ","があればスキップ
	}
	p.consume(")")             // ")"をスキップ
	p.consumeIfPossible("int") // "int"があればスキップ
	fn := &Node{kind: ND_FUNCDECL, token: funcname, val: funcname.val, body: p.block(), params: params, lvar: p.lvar}

	// 変数のオフセット計算
	offset := 0
	for _, variable := range fn.params {
		offset += 8
		variable.offset = offset
	}
	for _, variable := range fn.lvar {
		offset += 8
		variable.offset = offset
	}

	fn.offset = align_to(offset, 16) // 関数のオフセット計算

	p.leave_scope() // スコープを削除
	return fn
}

// 以下構文規則
// Block            = "{" statementList "}" .
// statementList = { statement ";" } .
func (p *Parser) block() *Node {
	p.enter_scope() // ブロックスコープを追加
	p.consume("{")

	node := &Node{kind: ND_BLOCK, block: []*Node{}}
	for {
		p.consumeIfPossible(";") // ";"があればスキップ
		if p.startsWithValue("}") {
			p.consume("}")  // "}"をスキップ
			p.leave_scope() // ブロックスコープを削除
			return node
		}
		node.block = append(node.block, p.stmt())
		if !p.startsWithValue(";") && !p.startsWithValue("}") {
			error_tok(p.code, p.peek(1)[0], "セミコロンが見つかりません")
		}
	}
}

// statement        = "return" expr | VarDecl | IfStmt | ForStmt | block | SimpleStmt .
func (p *Parser) stmt() *Node {
	switch {
	case p.startsWithValue("return"): // return statement
		p.consume("return")
		return &Node{kind: ND_RETURN_STMT, lhs: p.expr()}
	case p.startsWithValue("var"): // VarDecl
		return p.varDecl()
	case p.startsWithValue("if"): // IfStmt
		return p.ifStmt()
	case p.startsWithValue("for"): // ForStmt
		return p.forStmt()
	case p.startsWithValue("{"): // block
		return p.block()
	default: // simple statement
		return p.simpleStmt()
	}
}

// VarDecl       = "var" ident ( "int" [ "=" expr ] | "=" expr ) .
func (p *Parser) varDecl() *Node {
	p.consume("var")

	token := p.consumeWithTokenKind(TK_IDENT)
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

	p.consumeIfPossible("int")
	rhs := &Node{kind: ND_NUM, val: "0"} // 宣言のみの場合はゼロ値で初期化
	if p.startsWithValue("=") {
		p.consume("=") // "="をスキップ
		rhs = p.expr()
	}
	return &Node{kind: ND_ASSIGN_STMT, lhs: lhs, rhs: rhs}
}

// IfStmt           = "if" [ SimpleStmt ";" ] expr Block [ "else" ( IfStmt | Block ) ] .
func (p *Parser) ifStmt() *Node {
	p.consume("if")
	p.enter_scope() // ifスコープを追加
	node := &Node{kind: ND_IF_STMT}
	if condOrInit := p.simpleStmt(); p.startsWithValue(";") {
		// 初期化子あり
		p.consume(";")
		node.init = condOrInit
		node.cond = p.expr()
	} else {
		// 初期化子なし
		node.cond = condOrInit.lhs
	}
	node.then = p.block()
	if p.startsWithValue("else") {
		p.consume("else") // "else"をスキップ
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
	p.consume("for")
	p.enter_scope() // forスコープを追加
	node := &Node{kind: ND_FOR_STMT}

	// 条件式を省略したパターン
	if p.startsWithValue("{") {
		node.then = p.block()
		p.leave_scope() // forスコープを削除
		return node
	}

	// 条件式のみのパターン
	if !p.startsWithValue(";") {
		stmt := p.simpleStmt()
		if p.startsWithValue("{") {
			node.cond = stmt.lhs
			node.then = p.block()
			p.leave_scope() // forスコープを削除
			return node
		}
		node.init = stmt
	}

	// for句を用いるパターン
	p.consume(";") // 最初のセミコロンをスキップ
	if !p.startsWithValue(";") {
		node.cond = p.expr()
	}
	p.consume(";") // 2つ目のセミコロンをスキップ

	if !p.startsWithValue("{") {
		node.inc = p.simpleStmt()
	}
	node.then = p.block()
	p.leave_scope() // forスコープを削除
	return node
}

// SimpleStmt       = ExpressionStmt | Assignment .
// ExpressionStmt   = expr .
// Assignment       = expr "=" expr .
func (p *Parser) simpleStmt() *Node {
	lhs := p.expr()
	if !p.startsWithValue("=") {
		return &Node{kind: ND_EXPR_STMT, lhs: lhs}
	}

	if lhs.kind != ND_VAR {
		error_tok(p.code, lhs.token, "左辺が変数ではありません")
	}
	p.consume("=") // "="をスキップ
	return &Node{kind: ND_ASSIGN_STMT, lhs: lhs, rhs: p.expr()}
}

// expr = add { "==" add | "!=" add | "<" add | "<=" add | ">" add | ">=" add } .
func (p *Parser) expr() *Node {
	node := p.add()
	for {
		switch {
		case p.startsWithValue("=="):
			p.consume("==")
			node = &Node{kind: ND_EQ, lhs: node, rhs: p.add()}
			continue
		case p.startsWithValue("!="):
			p.consume("!=")
			node = &Node{kind: ND_NE, lhs: node, rhs: p.add()}
			continue
		case p.startsWithValue("<"):
			p.consume("<")
			node = &Node{kind: ND_LT, lhs: node, rhs: p.add()}
			continue
		case p.startsWithValue("<="):
			p.consume("<=")
			node = &Node{kind: ND_LE, lhs: node, rhs: p.add()}
			continue
		case p.startsWithValue(">"):
			p.consume(">")
			node = &Node{kind: ND_LT, lhs: p.add(), rhs: node}
			continue
		case p.startsWithValue(">="):
			p.consume(">=")
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
		switch {
		case p.startsWithValue("+"):
			p.consume("+")
			node = &Node{kind: ND_ADD, lhs: node, rhs: p.mul()}
			continue
		case p.startsWithValue("-"):
			p.consume("-")
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
		switch {
		case p.startsWithValue("*"):
			p.consume("*")
			node = &Node{kind: ND_MUL, lhs: node, rhs: p.unary()}
			continue
		case p.startsWithValue("/"):
			p.consume("/")
			node = &Node{kind: ND_DIV, lhs: node, rhs: p.unary()}
			continue
		}
		return node
	}
}

// unary = primary | [ "+" | "-" ] unary .
func (p *Parser) unary() *Node {
	switch {
	case p.startsWithValue("+"):
		p.consume("+")
		return p.unary()
	case p.startsWithValue("-"):
		p.consume("-")
		zero := &Node{kind: ND_NUM, val: "0"}
		return &Node{kind: ND_SUB, lhs: zero, rhs: p.unary()}
	}
	return p.primary()
}

// primary       = num | ident | funccall | "(" expr ")" .
func (p *Parser) primary() *Node {
	switch {
	case p.startsWithTokenKind(TK_NUM):
		return p.num()
	case p.startsWithTokenKind(TK_IDENT):
		if p.peek(2)[1].val == "(" {
			return p.funccall()
		}
		return p.ident()
	case p.startsWithValue("("):
		p.consume("(")
		node := p.expr()
		p.consume(")")
		return node
	}
	error_tok(p.code, p.peek(1)[0], "不正なトークンです")
	return nil
}

// num = digit { digit } .
func (p *Parser) num() *Node {
	token := p.consumeWithTokenKind(TK_NUM)
	return &Node{kind: ND_NUM, token: token, val: token.val}
}

// ident = letter { alnum } .
func (p *Parser) ident() *Node {
	token := p.consumeWithTokenKind(TK_IDENT)
	// スコープから変数を探す
	for _, m := range p.scope {
		if variable, ok := m[token.val]; ok {
			return &Node{kind: ND_VAR, token: token, val: token.val, variable: variable}
		}
	}
	// 変数がいずれのスコープにも宣言されていないならエラー
	error_tok(p.code, token, "変数が宣言されていません。")
	return nil
}

// funccall = ident "(" [ ExpressionList [ "," ] ] ")" .
// ExpressionList = Expression { "," Expression } .
func (p *Parser) funccall() *Node {
	funcname := p.consumeWithTokenKind(TK_IDENT)
	node := &Node{kind: ND_FUNCCALL, val: funcname.val, args: []*Node{}}
	p.consume("(")
	for !p.startsWithValue(")") {
		node.args = append(node.args, p.expr())
		if !p.startsWithValue(",") && !p.startsWithValue(")") {
			error_tok(p.code, p.peek(1)[0], "不正なトークン")
		}
		p.consumeIfPossible(",")
	}
	if len(node.args) > 6 {
		error_tok(p.code, funcname, "引数が多すぎます(6個以内)")
	}
	p.consume(")")
	return node
}
