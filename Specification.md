言語仕様

```ebnf
program          = block ";" .
block            = "{" statementList "}" .
statementList    = { statement ";" } .
statement        = "return" expr | VarDecl | IfStmt | forStmt | block | SimpleStmt .
SimpleStmt       = Assignment .
IfStmt           = "if" Expression Block [ "else" ( IfStmt | Block ) ] .

ForStmt          = "for" [ Condition | ForClause ] Block .
Condition        = Expression .
ForClause        = [ InitStmt ] ";" [ Condition ] ";" [ PostStmt ] .
InitStmt         = SimpleStmt .
PostStmt         = SimpleStmt .

VarDecl          = "var" ident ( "int" [ "=" expr ] | "=" expr ) .
assignStmt       = expr [ "=" expr ] .
expr             = add { "==" add | "!=" add | "<" add | "<=" add | ">" add | ">=" add } .
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
