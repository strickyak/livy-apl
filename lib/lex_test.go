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
			{EndToken, "", 0, nil},
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
			{SemiToken, "\n", 2, []string{"\n"}},
			{EndToken, "", len(s), nil},
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
			{VariableToken, "Abc", 1, []string{"Abc", "Abc"}},
			{OperatorToken, "==", 4, []string{"==", "=="}},
			{NumberToken, "666", 6, []string{"666", "666", "", ""}},
			{EndToken, "", len(s), nil},
		},
		Source: s,
		p:      len(s),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %s wanted %s", got, want)
	}
	println("250 OK")
}
