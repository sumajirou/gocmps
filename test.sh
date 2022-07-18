#! /bin/bash
cat <<EOF | gcc -xc -c -o tmp2.o -
int ret3() { return 3; }
int ret5() { return 5; }
int add(int x, int y) { return x+y; }
int sub(int x, int y) { return x-y; }

int add6(int a, int b, int c, int d, int e, int f) {
  return a+b+c+d+e+f;
}
EOF

assert() {
  expected="$1"
  input="$2"

  ./gocmps "$input" > tmp.s || exit
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

assert 0  'func main() { return 0 }'
assert 42 'func main() { return 42 }'
assert 21 'func main() { return 5+20-4 }'
assert 41 'func main() { return 12 + 34 - 5 }'
assert 47 'func main() { return 5+6*7 }'
assert 15 'func main() { return 5*(9-6) }'
assert 4  'func main() { return (3+5)/2 }'
assert 10 'func main() { return -10+20 }'
assert 10 'func main() { return - -10 }'
assert 10 'func main() { return - - +10 }'
assert 25 'func main() { return - 5 * - 5 }'

assert 0  'func main() { return 0==1 }'
assert 1  'func main() { return 42==42 }'
assert 1  'func main() { return 0!=1 }'
assert 0  'func main() { return 42!=42 }'
assert 1  'func main() { return 0<1 }'
assert 0  'func main() { return 1<1 }'
assert 0  'func main() { return 2<1 }'
assert 1  'func main() { return 0<=1 }'
assert 1  'func main() { return 1<=1 }'
assert 0  'func main() { return 2<=1 }'
assert 1  'func main() { return 1>0 }'
assert 0  'func main() { return 1>1 }'
assert 0  'func main() { return 1>2 }'
assert 1  'func main() { return 1>=0 }'
assert 1  'func main() { return 1>=1 }'
assert 0  'func main() { return 1>=2 }'

assert 1  'func main() { return 1; 2; 3 }'
assert 2  'func main() { 1; return 2; 3 }'
assert 3  'func main() { 1; 2; return 3 }'

assert 3  'func main() { var a=3; return a }'
assert 8  'func main() { var a=3; var z=5; return a+z }'
assert 3  'func main() { var foo=3; return foo }'
assert 8  'func main() { var foo123=3; var bar=5; return foo123+bar }'

assert 3  'func main() { { 1; { 2; }; return 3; }; }'
assert 4  'func main() { {}; {;}; {1;}; {2;3}; return 4}'

assert 6  'func main() { var a int = 1; var b int; b=2; var c=3; return a+b+c}'
assert 4  'func main() { var a int; {a=4}; return a}'
assert 0  'func main() { var a int; {var a int = 4}; return a}'

assert 3  'func main() { if 0 { return 2 }; return 3 }'
assert 3  'func main() { if 1-1 { return 2 }; return 3 }'
assert 2  'func main() { if 1 { return 2 }; return 3 }'
assert 2  'func main() { if 2-1 { return 2 }; return 3 }'
assert 4  'func main() { if 0 { 1; 2; return 3 } else { return 4 } }'
assert 3  'func main() { if 1 { 1; 2; return 3 } else { return 4 } }'
assert 5  'func main() { if 0 { return 3 } else if 0 { return 4 } else { return 5 } }'
assert 2  'func main() { if ;1 { return 2 }; return 3 }'
assert 3  'func main() { if ;0 { return 2 }; return 3 }'
assert 2  'func main() { var i int; if i=1;i { return 2 }; return 3 }'
assert 3  'func main() { var i int; if i=0;i { return 2 }; return 3 }'

assert 2  'func main() { var i int; if i=1;i { return 2 }; return 3 }'
assert 3  'func main() { var i int; if i=0;i { return 2 }; return 3 }'

assert 55 'func main() { var i=0; var j=0; for i=0; i<=10; i=i+1 { j=i+j }; return j; }'
assert 3  'func main() { for { return 3 }; return 5 }'
assert 3  'func main() { for 1 { return 3 }; return 5 }'
assert 5  'func main() { for 0 { return 3 }; return 5 }'
assert 3  'func main() { for ;; { return 3 }; return 5 }'
assert 5  'func main() { for ;0; { return 3 }; return 5 }'
assert 3  'func main() { var i int; for ;;i=i+1 { return 3 }; return 5 }'

assert 3  'func main() { return ret3() }'
assert 1  'func main() { if ret5() == 5 {return 1}; return 0 }'
assert 8  'func main() { return add(3, 5) }'
assert 2  'func main() { return sub(5, 3) }'
assert 21 'func main() { return add6(1,2,3,4,5,6) }'

assert 32 'func main() { return ret32() }; func ret32() int { return 32 }'
assert 5  'func main() { return myadd(2,3) }; func myadd(a int, b int) int { return a+b }'
# fibonacci = [0,1,1,2,3,5,8,13,21,34,55]
assert 55 'func fib_for(n int) int {
  var a int = 0
  var b int = 1
  var i int
  for i = 0; i<n; i = i + 1{
    b = a+b
    a = b-a
  }
  return a
}

func main() {
  return fib_for(10)
}
'
assert 55 'func fib_rec(n int) int {
  if n == 0 {
    return 0
  }
  if n == 1 {
    return 1
  }
  return fib_rec(n-1) + fib_rec(n-2)
}

func main() {
  return fib_rec(10)
}
'

assert 3 'func main() { var x int = 3; return *&x; }'
assert 3 'func main() { var x int = 3; return *&*&x; }'
assert 3 'func main() { var x int = 3; var y = &x; var z = &y; return **z; }'
assert 5 'func main() { var x int = 3; var y = &x; *y = 5; return x; }'

printf '\033[32m%s\033[m\n' 'OK'
