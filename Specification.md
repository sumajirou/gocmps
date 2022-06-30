言語仕様

```ebnf
expr    = mul { "+" mul | "-" mul } .
mul     = unary { "*" unary | "/" unary } .
unary   = primary | [ "+" | "-" ] unary .
primary = num | "(" expr ")" .
num     = digit { digit } .
```

```ebnf
digit    = "0" … "9" .
```
