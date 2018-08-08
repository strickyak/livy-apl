package livy

import (
	"log"
	"math"
	"reflect"
	"testing"
)

type test struct {
	src  string
	want string
}

var tests = []test{
	{"+/ rho iota 8", "M(+/ M(rho M(iota [#8])))"},
	{"+/ rho iota Pi", "M(+/ M(rho M(iota [Pi])))"},
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
	for _, test := range tests {
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

var evalTests = map[string]string{
	"( 4 5 6 )":                  "[3 ]{4 5 6 } ",
	"( 4 + 5 + 6 )":              "15 ",
	"rho ( 4 + 5 + 6 )":          "[]{} ",
	"( 1 + iota 4 ) rho 8":       "[1 2 3 4 ]{8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 } ",
	"(3) 4 (5)":                  "[3 ]{3 4 5 } ",
	"9 rho 9":                    "[9 ]{9 9 9 9 9 9 9 9 9 } ",
	"3 3 rho rho 3 8 rho iota 9": "[3 3 ]{3 8 3 8 3 8 3 8 3 } ",
}

func TestEval(t *testing.T) {
	for src, want := range evalTests {
		log.Printf("TestEval <<< %q", src)
		c := Standard()
		lex := Tokenize(src)
		expr, _ := ParseExpr(lex, 0)
		got := expr.Eval(c)
		log.Printf("TestEval === %q", src)
		log.Printf("TestEval >>> %q", got)

		if got.String() != want {
			t.Errorf("Got %q, wanted %q, for src %q", got, want, src)
		}
	}
	println("250 OK")
}
