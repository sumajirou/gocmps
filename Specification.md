言語仕様

```ebnf
stmt       = "return" expr ";" | assignStmt .
assignStmt = expr [ "=" expr ] ";" .
expr       = add { "==" add | "!=" add | "<" add | "<=" add | ">" add | ">=" add } .
add        = mul { "+" mul | "-" mul } .
mul        = unary { "*" unary | "/" unary } .
unary      = primary | [ "+" | "-" ] unary .
primary    = num | ident | "(" expr ")" .
num        = digit { digit } .
ident      = letter { alnum } .
```

```ebnf
digit    = "0" … "9" .
letter   = "A" … "Z" | "a" … "z" | "_" .
alnum    = digit | letter
```
