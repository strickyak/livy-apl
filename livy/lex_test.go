package livy

import (
	"reflect"
	"testing"
)

func TestEmpty(t *testing.T) {
	s := ""
	got := Tokenize(s)
	want := &Lex{
		Tokens: []*Token{
			{EndToken, "", 0},
		},
		Source: s,
		p:      len(s),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %s wanted %s", got, want)
	}
	println("250 OK")
}

func TestWhite(t *testing.T) {
	s := " \t\n\r "
	got := Tokenize(s)
	want := &Lex{
		Tokens: []*Token{
			{EndToken, "", len(s)},
		},
		Source: s,
		p:      len(s),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %s wanted %s", got, want)
	}
	println("250 OK")
}

func TestVarOpNum(t *testing.T) {
	s := " Abc==666 "
	got := Tokenize(s)
	want := &Lex{
		Tokens: []*Token{
			{VariableToken, "Abc", 1},
			{OperatorToken, "==", 4},
			{NumberToken, "666", 6},
			{EndToken, "", len(s)},
		},
		Source: s,
		p:      len(s),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %s wanted %s", got, want)
	}
	println("250 OK")
}
