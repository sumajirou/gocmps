言語仕様

```ebnf
program          = { FunctionDecl ";" } .
FunctionDecl     = "func" ident Parameters [ "int" ] Block .
Parameters       = "(" [ ident "int" { "," ident "int" } [ "," ] ] ")" .
Block            = "{" statementList "}" .
statementList    = { statement ";" } .
statement        = "return" expr | VarDecl | IfStmt | ForStmt | block | SimpleStmt .
SimpleStmt       = ExpressionStmt | Assignment .
IfStmt           = "if" [ SimpleStmt ";" ] expr Block [ "else" ( IfStmt | Block ) ] .
ForStmt          = "for" [ Condition | ForClause ] Block .
Condition        = expr .
ForClause        = [ InitStmt ] ";" [ Condition ] ";" [ PostStmt ] .
InitStmt         = SimpleStmt .
PostStmt         = SimpleStmt .
VarDecl          = "var" ident ( "int" [ "=" expr ] | "=" expr ) .
ExpressionStmt   = expr .
Assignment       = expr "=" expr .
expr       = add { "==" add | "!=" add | "<" add | "<=" add | ">" add | ">=" add } .
add              = mul { "+" mul | "-" mul } .
mul              = unary { "*" unary | "/" unary } .
unary            = primary | [ "+" | "-" ] unary .
primary          = num | ident | funcall | "(" expr ")" .
funcall          = ident "(" [ ExpressionList [ "," ] ] ")" .
ExpressionList   = Expression { "," Expression } .
num              = digit { digit } .
ident            = letter { alnum } .
```

```ebnf
digit    = "0" … "9" .
letter   = "A" … "Z" | "a" … "z" | "_" .
alnum    = digit | letter
```
