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
}

func TestEval(t *testing.T) {
	for _, p := range evalTests {
		log.Printf("TestEval <<< %q", p.src)
		c := Standard()
		lex := Tokenize(p.src)
		expr := ParseSeq(lex)
		got := expr.Eval(c)
		log.Printf("TestEval === %q", p.src)
		log.Printf("TestEval >>> %q", got)

		if got.String() != p.want {
			t.Errorf("Got %q, wanted %q, for src %q", got, p.want, p.src)
		}
	}
	println("250 OK")
}
