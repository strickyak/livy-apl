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

	// while
	{`S=0 ; N=6; Z = (while N>0 do S = N + S ; N = N - 1; iota S  done ) ; S`, `21 `},
	{`S=0 ; N=8; Z = (while N>0 do S = N + S ; N = N - 1; if S>26 then break else  S fi  done ) ; S`, `30 `},

	// and or xor
	{`1 1 0 0 and 1 0 1 0 `, `[4 ]{1 0 0 0 } `},
	{`1 1 0 0 or 1 0 1 0 `, `[4 ]{1 1 1 0 } `},
	{`1 1 0 0 xor 1 0 1 0 `, `[4 ]{0 1 1 0 } `},

	// ravel
	{`, 4 6 rho iota1 100`, `[24 ]{1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 } `},
	{`, 3.14`, `[1 ]{3.14 } `},

	{`(2 3 4 rho iota1 100) ,[0] ( 2 3 4 rho neg iota1 100)`,
		`[4 3 4 ]{1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 -1 -2 -3 -4 -5 -6 -7 -8 -9 -10 -11 -12 -13 -14 -15 -16 -17 -18 -19 -20 -21 -22 -23 -24 } `},

	{`(2 3 4 rho iota1 100) , ( 2 3 1 rho neg iota1 100)`,
		`[2 3 5 ]{1 2 3 4 -1 5 6 7 8 -2 9 10 11 12 -3 13 14 15 16 -4 17 18 19 20 -5 21 22 23 24 -6 } `},

	{`transpose 3 5 rho iota1 99`,
		`[5 3 ]{1 6 11 2 7 12 3 8 13 4 9 14 5 10 15 } `},
	{`transpose 2 3 5 rho iota1 99`,
		`[2 5 3 ]{1 6 11 2 7 12 3 8 13 4 9 14 5 10 15 16 21 26 17 22 27 18 23 28 19 24 29 20 25 30 } `},
	{`transpose[1] 2 3 5 rho iota1 99`,
		`[3 2 5 ]{1 2 3 4 5 16 17 18 19 20 6 7 8 9 10 21 22 23 24 25 11 12 13 14 15 26 27 28 29 30 } `},

	{`0 1 2 transpose 3 3 3 rho iota1 27`,
		`[3 3 3 ]{1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 } `},
	{`1 0 2  transpose 3 3 3 rho iota1 27`,
		`[3 3 3 ]{1 2 3 10 11 12 19 20 21 4 5 6 13 14 15 22 23 24 7 8 9 16 17 18 25 26 27 } `},
	{`0 2 1 transpose 3 3 3 rho iota1 27`,
		`[3 3 3 ]{1 4 7 2 5 8 3 6 9 10 13 16 11 14 17 12 15 18 19 22 25 20 23 26 21 24 27 } `},
	{`2 1 0 transpose 3 3 3 rho iota1 27`,
		`[3 3 3 ]{1 10 19 4 13 22 7 16 25 2 11 20 5 14 23 8 17 26 3 12 21 6 15 24 9 18 27 } `},

	{` (6 7 rho iota 42) member (7 7 rho 3 * iota 10) `,
		`[6 7 ]{1 0 0 1 0 0 1 0 0 1 0 0 1 0 0 1 0 0 1 0 0 1 0 0 1 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 } `},

	{` A = , 7 7 rho (3 * iota 30) mod 19 ; up A `, `[49 ]{0 19 30 13 43 37 7 26 31 20 1 14 44 38 27 8 32 2 21 45 15 39 28 9 33 3 22 16 46 29 40 10 4 23 34 47 17 11 41 5 24 35 18 48 12 42 36 25 6 } `},
	{` A = , 7 7 rho (3 * iota 30) mod 19 ; A [ up A ]`,
		`[49 ]{0 0 0 1 1 2 2 2 3 3 3 4 4 5 5 5 6 6 6 7 7 8 8 8 9 9 9 10 10 11 11 11 12 12 12 13 13 14 14 15 15 15 16 16 17 17 18 18 18 } `},
	{` A = , 7 7 rho (3 * iota 30) mod 19 ; A [ down A ]`,
		`[49 ]{18 18 18 17 17 16 16 15 15 15 14 14 13 13 12 12 12 11 11 11 10 10 9 9 9 8 8 8 7 7 6 6 6 5 5 5 4 4 3 3 3 2 2 2 1 1 0 0 0 } `},
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
