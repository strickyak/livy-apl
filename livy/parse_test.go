package livy

import (
	"log"
	"math"
	"reflect"
	"testing"
)

type srcWantPair struct {
	src  string
	want string
}

var parserTests = []srcWantPair{
	{"+/ rho iota 8", "Monad(+/ Monad(rho Monad(iota (#8))))"},
	{"+/ rho iota Pi", "Monad(+/ Monad(rho Monad(iota (Pi))))"},
}

func Standard() *Context {
	c := &Context{
		Globals:  make(map[string]Val),
		Monadics: StandardMonadics,
		Dyadics:  StandardDyadics,
	}
	c.Globals["Pi"] = &Num{math.Pi}
	c.Globals["Tau"] = &Num{2.0 * math.Pi}
	c.Globals["Zero"] = &Num{0.0}
	c.Globals["One"] = &Num{1.0}
	c.Globals["Two"] = &Num{2.0}
	return c
}

func TestMonadic(t *testing.T) {
	for _, test := range parserTests {
		lex := Tokenize(test.src)
		expr, _ := ParseExpr(lex, 0)
		log.Printf("EXPR: %s", expr)
		got := expr.String()

		if got != test.want {
			t.Errorf("Got %q wanted %q, for %q", got, test.want, test.src)
		}
	}
	println("250 OK")
}

func TestEvalMonadic(t *testing.T) {
	c := Standard()
	lex := Tokenize("double Pi")
	expr, _ := ParseExpr(lex, 0)
	got := expr.Eval(c)

	want := &Num{2.0 * math.Pi}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %s wanted %s", got, want)
	}
	println("250 OK")
}

func TestEvalDyadic(t *testing.T) {
	c := Standard()
	lex := Tokenize("One + Two")
	expr, _ := ParseExpr(lex, 0)
	got := expr.Eval(c)

	want := &Num{3.0}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %s wanted %s", got, want)
	}
	println("250 OK")
}

func TestEvalParens(t *testing.T) {
	c := Standard()
	lex := Tokenize("( 1 + iota 4 ) rho iota 5")
	expr, _ := ParseExpr(lex, 0)
	got := expr.Eval(c)

	const want = "[1 2 3 4 ]{0 1 2 3 4 0 1 2 3 4 0 1 2 3 4 0 1 2 3 4 0 1 2 3 } "
	if got.String() != want {
		t.Errorf("Got %q wanted %q", got, want)
	}
	println("250 OK")
}

func TestEvalLiteralList(t *testing.T) {
	c := Standard()
	lex := Tokenize("2 3  5 rho 4 6 8")
	expr, _ := ParseExpr(lex, 0)
	got := expr.Eval(c)

	const want = "[2 3 5 ]{4 6 8 4 6 8 4 6 8 4 6 8 4 6 8 4 6 8 4 6 8 4 6 8 4 6 8 4 6 8 } "
	if got.String() != want {
		t.Errorf("Got %q wanted %q", got, want)
	}
	println("250 OK")
}

var evalTests = []srcWantPair{
	{"( 4 5 6 )",
		"[3 ]{4 5 6 } "},

	{"( 4 + 5 + 6 )",
		"15 "},

	{"rho ( 4 + 5 + 6 )",
		"[]{} "},

	{"( 1 + iota 4 ) rho 8",
		"[1 2 3 4 ]{8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 } "},

	{"(3) 4 (5)",
		"[3 ]{3 4 5 } "},

	{"9 rho 9",
		"[9 ]{9 9 9 9 9 9 9 9 9 } "},

	{"3 3 rho rho 3 8 rho iota 9",
		"[3 3 ]{3 8 3 8 3 8 3 8 3 } "},

	{"A=3 3 rho square iota 10; A[ 1 2 ; 1 2 ]",
		"[2 2 ]{16 25 49 64 } "},

	{"A=3 3 rho square iota 10; A[ 1 2 ; ]",
		"[2 3 ]{9 16 25 36 49 64 } "},

	{"A=3 3 rho square iota 10; A[ ; 1 2 ]",
		"[3 2 ]{1 4 16 25 49 64 } "},

	{"A=3 3 rho square iota 10; A[;]",
		"[3 3 ]{0 1 4 9 16 25 36 49 64 } "},

	{"A=3 3 rho square iota 10; A[2;2]",
		"[1 1 ]{64 } "},

	{"+/ 888 + iota 1",
		"888 "},

	{"*/ 1+ iota 6",
		"720 "},

	{"+/ 3 3 3 rho 1 + iota 100",
		"[3 3 ]{6 15 24 33 42 51 60 69 78 } "},

	{`+\ iota 10`,
		"[10 ]{0 1 3 6 10 15 21 28 36 45 } "},

	{`*\ 3 rho 8`,
		"[3 ]{8 64 512 } "},

	{`+/ 8 + iota 1`,
		"8 "},

	{`+/ 8 + iota 0`,
		"0 "},

	{`+\ 8 + iota 1`,
		"[1 ]{8 } "},

	{`+\ 8 + iota 0`,
		"[0 ]{} "},

	{`+/[0]  3 3 3 rho  iota 100`,
		"[3 3 ]{27 30 33 36 39 42 45 48 51 } "},

	{`+/[1]  3 3 3 rho  iota 100`,
		"[3 3 ]{9 12 15 36 39 42 63 66 69 } "},

	{`(iota 9) ..== iota 9`,
		`[9 9 ]{1 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 1 } `},

	{`(3 9 rho iota 99)  +.* (iota 9) ..== iota 9`,
		`[3 9 ]{0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 } `},

	{`(3 3 rho 1 0 0  0 1 0  0 0 1)  +.*  (3 9 rho iota 99)  +.* (iota 9) ..== iota 9`,
		`[3 9 ]{0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 } `},

	{`2 -2 take 3 9 rho 1 + iota 27`,
		`[2 2 ]{8 9 17 18 } `},

	{`2 -2 drop 3 9 rho 1 + iota 27`,
		`[1 7 ]{19 20 21 22 23 24 25 } `},

	{`rot 2 -2 drop 3 9 rho 1 + iota 27`,
		`[1 7 ]{25 24 23 22 21 20 19 } `},

	{`2 3 2 3 2 3 rot 2 -2 drop 3 9 rho 1 + iota 27`,
		`[1 6 ]{21 22 21 22 21 22 } `},

	{`1 0 2 rot[0] 3 9 rho 1 + iota 27`,
		`[3 9 ]{10 11 12 13 14 15 16 17 18 1 2 3 4 5 6 7 8 9 19 20 21 22 23 24 25 26 27 } `},

	{`1 0 1 0 1 \[0] 10 + 10 10 rho iota 100`,
		`[5 10 ]{10 11 12 13 14 15 16 17 18 19 0 0 0 0 0 0 0 0 0 0 20 21 22 23 24 25 26 27 28 29 0 0 0 0 0 0 0 0 0 0 30 31 32 33 34 35 36 37 38 39 } `},
	{`1 0 1 0 1 \[1] 10 + 10 10 rho iota 100`,
		`[10 5 ]{10 0 11 0 12 20 0 21 0 22 30 0 31 0 32 40 0 41 0 42 50 0 51 0 52 60 0 61 0 62 70 0 71 0 72 80 0 81 0 82 90 0 91 0 92 100 0 101 0 102 } `},
	{`1 0 1 0 1 /[0] 10 + 10 10 rho iota 100`,
		`[3 10 ]{10 11 12 13 14 15 16 17 18 19 30 31 32 33 34 35 36 37 38 39 50 51 52 53 54 55 56 57 58 59 } `},
	{`1 0 1 0 1 /[1] 10 + 10 10 rho iota 100`,
		`[10 3 ]{10 12 14 20 22 24 30 32 34 40 42 44 50 52 54 60 62 64 70 72 74 80 82 84 90 92 94 100 102 104 } `},

	// primes
	{`N=100; ( 2 == +/ 0 == (iota1 N) ..mod iota1 N ) / iota1 N`,
		`[25 ]{2 3 5 7 11 13 17 19 23 29 31 37 41 43 47 53 59 61 67 71 73 79 83 89 97 } `},

	// resize to scalar
	{`(iota 0) rho 111 + iota 10`, `111 `},

	// def
	{`def twice _x { _x * 2 } ; twice 4 5 6`, `[3 ]{8 10 12 } `},

	{`def primes N { ( 2 == +/ 0 == (iota1 N) ..mod iota1 N ) / iota1 N } ; primes 20`, `[8 ]{2 3 5 7 11 13 17 19 } `},

	{` def A dot B; C; D; E { C = D = E = F = 999 ;  A +.* B } ; 10 20 30 dot 1 2 3 `, `140 `},
	{` def A dot B; C; D; E; { C = D = E = F = 999 ;  A +.* B } ; 10 20 30 dot 1 2 3 `, `140 `},

	{`def *** A { (iota1 A) ..* iota1 A } ; *** 4`, `[4 4 ]{1 2 3 4 2 4 6 8 3 6 9 12 4 8 12 16 } `},
	{`def sum [Axis] B { +/[Axis] B } ; sum[0] 3 5 rho iota1 20`, `[5 ]{18 21 24 27 30 } `},
	{`def sum [Axis] B { +/[Axis] B } ; sum[1] 3 5 rho iota1 20`, `[3 ]{15 40 65 } `},

	// cond
	{`Pi=3.14 ; if Pi < 3 then -10 else +10 fi + iota 3`, `[3 ]{10 11 12 } `},
	{`Pi=3.14 ; if Pi < 4 then -10 else +10 fi + iota 3`, `[3 ]{-10 -9 -8 } `},
}

func TestEval(t *testing.T) {
	for _, p := range evalTests {
		log.Printf("TestEval <<< %q", p.src)
		c := Standard()
		lex := Tokenize(p.src)
		expr, _ := ParseSeq(lex, 0)
		got := expr.Eval(c)
		log.Printf("TestEval === %q", p.src)
		log.Printf("TestEval >>> %v", got)

		if got.String() != p.want {
			t.Errorf("Got %q, wanted %q, for src %q", got, p.want, p.src)
		}
	}
	println("250 OK")
}
