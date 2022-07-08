#! /bin/bash
cat <<EOF | gcc -xc -c -o tmp2.o -
int ret3() { return 3; }
int ret5() { return 5; }
EOF

assert() {
  expected="$1"
  input="$2"

  ./gocmps "$input" > tmp.s
  cc -o tmp tmp.s tmp2.o
  ./tmp
  actual="$?"

  if [ "$actual" = "$expected" ]; then
    echo "$input => $actual"
  else
    echo "$input => $expected expected, but got $actual"
    printf '\033[31m%s\033[m\n' 'NG'
    exit 1
  fi
}

assert 0 '{ return 0 }'
assert 42 '{ return 42 }'
assert 21 '{ return 5+20-4 }'
assert 41 '{ return  12 + 34 - 5 }'
assert 47 '{ return 5+6*7 }'
assert 15 '{ return 5*(9-6) }'
assert 4 '{ return (3+5)/2 }'
assert 10 '{ return -10+20 }'
assert 10 '{ return - -10 }'
assert 10 '{ return - - +10 }'
assert 25 '{ return - 5 * - 5 }'

assert 0 '{ return 0==1 }'
assert 1 '{ return 42==42 }'
assert 1 '{ return 0!=1 }'
assert 0 '{ return 42!=42 }'
assert 1 '{ return 0<1 }'
assert 0 '{ return 1<1 }'
assert 0 '{ return 2<1 }'
assert 1 '{ return 0<=1 }'
assert 1 '{ return 1<=1 }'
assert 0 '{ return 2<=1 }'
assert 1 '{ return 1>0 }'
assert 0 '{ return 1>1 }'
assert 0 '{ return 1>2 }'
assert 1 '{ return 1>=0 }'
assert 1 '{ return 1>=1 }'
assert 0 '{ return 1>=2 }'

assert 1 '{ return 1; 2; 3 }'
assert 2 '{ 1; return 2; 3 }'
assert 3 '{ 1; 2; return 3 }'

assert 3 '{ var a=3; return a }'
assert 8 '{ var a=3; var z=5; return a+z }'
assert 3 '{ var foo=3; return foo }'
assert 8 '{ var foo123=3; var bar=5; return foo123+bar }'

assert 3 '{ { 1; { 2; }; return 3; }; }'
assert 4 '{{}; {;}; {1;}; {2;3}; return 4}'

assert 6 '{var a int = 1; var b int; b=2; var c=3; return a+b+c}'
assert 4 '{var a int; {a=4}; return a}'
assert 0 '{var a int; {var a int = 4}; return a}'

assert 3 '{ if 0 { return 2 }; return 3 }'
assert 3 '{ if 1-1 { return 2 }; return 3 }'
assert 2 '{ if 1 { return 2 }; return 3 }'
assert 2 '{ if 2-1 { return 2 }; return 3 }'
assert 4 '{ if 0 { 1; 2; return 3 } else { return 4 } }'
assert 3 '{ if 1 { 1; 2; return 3 } else { return 4 } }'
assert 5 '{ if 0 { return 3 } else if 0 { return 4 } else { return 5 } }'

assert 55 '{ var i=0; var j=0; for i=0; i<=10; i=i+1 { j=i+j }; return j; }'
assert 3 '{ for { return 3 }; return 5 }'
assert 3 '{ for ;; { return 3 }; return 5 }'
assert 3 '{ for var i int;; { return 3 }; return 5 }'
assert 5 '{ for ;0; { return 3 }; return 5 }'
assert 3 '{ var i int; for ;;i=i+1 { return 3 }; return 5 }'

assert 3 '{ return ret3(); }'
assert 1 '{ if ret5() == 5 {return 1}; return 0; }'


printf '\033[32m%s\033[m\n' 'OK'

