言語仕様

```ebnf
stmt     = exprStmt .
exprStmt = expr ";" .
expr     = add { "==" add | "!=" add | "<" add | "<=" add | ">" add | ">=" add } .
add      = mul { "+" mul | "-" mul } .
mul      = unary { "*" unary | "/" unary } .
unary    = primary | [ "+" | "-" ] unary .
primary  = num | "(" expr ")" .
num      = digit { digit } .
```

```ebnf
digit    = "0" … "9" .
```
