言語仕様

```ebnf
expr    = mul { "+" mul | "-" mul } .
mul     = primary { "*" primary | "/" primary } .
primary = num | "(" expr ")" .
num     = digit { digit } .
```

```ebnf
digit    = "0" … "9" .
```
