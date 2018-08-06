package livy

import (
	"log"
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

func TestMonadic(t *testing.T) {
	for _, test := range tests {
		lex := Tokenize(test.src)
		expr := Parse(lex, 0)
		log.Printf("EXPR: %s", expr)
		got := expr.String()

		if got != test.want {
			t.Errorf("Got %q wanted %q, for %q", got, test.want, test.src)
		}
	}
	println("250 OK")
}

func TestEval(t *testing.T) {
	c := &Context{
		Globals:  make(map[string]Val),
		Monadics: Monadics,
	}
	c.Globals["Pi"] = &Num{3.14}
	lex := Tokenize("double Pi")
	expr := Parse(lex, 0)
	got := expr.Eval(c)

	want := &Num{6.28}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %s wanted %s", got, want)
	}
	println("250 OK")
}
