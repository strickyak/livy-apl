package livy

import (
	"fmt"
	"log"
	"regexp"
)

type TokenType int

const (
	EndToken TokenType = iota
	NumberToken
	VariableToken
	OperatorToken
	OpenToken
	CloseToken
)

var MatchWhite = regexp.MustCompile(`^(\s*)`).FindStringSubmatch
var MatchNumber = regexp.MustCompile(`^([-]?[0-9]+([.][0-9]+)?([eE][-]?[0-9]+)?)`).FindStringSubmatch
var MatchVariable = regexp.MustCompile(`^([A-Z_][A-Za-z0-9_]*)`).FindStringSubmatch
var MatchOperator = regexp.MustCompile(`^(([-+*/,&|!=<>]([/=]?))|([a-z][A-Za-z0-9_]*))`).FindStringSubmatch
var MatchOpen = regexp.MustCompile(`^[(]`).FindStringSubmatch
var MatchClose = regexp.MustCompile(`^[)]`).FindStringSubmatch

type Token struct {
	Type TokenType
	Str  string
	Pos  int
}

func (t Token) String() string {
	return fmt.Sprintf("T(%d,%q,%d)", t.Type, t.Str, t.Pos)
}

type Lex struct {
	Tokens []*Token
	Source string
	p      int
}

func (x Lex) String() string {
	z := fmt.Sprintf("Lex{%d,%q,\n", x.p, x.Source)
	for i, t := range x.Tokens {
		z += fmt.Sprintf("  [%d]  %s\n", i, *t)
	}
	return z + "\n}"
}

func Tokenize(s string) *Lex {
	// log.Printf("TOKENIZE %q", s)
	lex := &Lex{
		Source: s,
	}
	for lex.Next() {
		// log.Printf("LEX... %s", *lex)
		continue
	}
	// log.Printf("LEX... %s", *lex)

	llt := len(lex.Tokens)
	if llt == 0 || lex.Tokens[llt-1].Type != EndToken {
		log.Panicf("Syntax error after %q before %q", s[:lex.p], s[lex.p:])
	}
	if lex.p != len(s) {
		log.Panicf("OHNO did not parse all of %q: %d", s, lex.p)
	}
	for i, t := range lex.Tokens {
		log.Printf("Token [%d]: %s", i, t)
	}
	return lex
}

func (lex *Lex) Next() bool {
	mw := MatchWhite(lex.Source[lex.p:])
	if mw != nil {
		lex.p += len(mw[0])
	}
	if lex.p == len(lex.Source) {
		t := &Token{
			Type: EndToken,
			Str:  "",
			Pos:  lex.p,
		}
		lex.Tokens = append(lex.Tokens, t)
		return false
	}
	mn := MatchNumber(lex.Source[lex.p:])
	if mn != nil {
		t := &Token{
			Type: NumberToken,
			Str:  mn[0],
			Pos:  lex.p,
		}
		lex.p += len(mn[0])
		lex.Tokens = append(lex.Tokens, t)
		return true
	}
	mv := MatchVariable(lex.Source[lex.p:])
	if mv != nil {
		t := &Token{
			Type: VariableToken,
			Str:  mv[0],
			Pos:  lex.p,
		}
		lex.p += len(mv[0])
		lex.Tokens = append(lex.Tokens, t)
		return true
	}
	mo := MatchOperator(lex.Source[lex.p:])
	if mo != nil {
		t := &Token{
			Type: OperatorToken,
			Str:  mo[0],
			Pos:  lex.p,
		}
		lex.p += len(mo[0])
		lex.Tokens = append(lex.Tokens, t)
		return true
	}
	mopen := MatchOpen(lex.Source[lex.p:])
	if mopen != nil {
		t := &Token{
			Type: OpenToken,
			Str:  mopen[0],
			Pos:  lex.p,
		}
		lex.p += len(mopen[0])
		lex.Tokens = append(lex.Tokens, t)
		return true
	}
	mclose := MatchClose(lex.Source[lex.p:])
	if mclose != nil {
		t := &Token{
			Type: CloseToken,
			Str:  mclose[0],
			Pos:  lex.p,
		}
		lex.p += len(mclose[0])
		lex.Tokens = append(lex.Tokens, t)
		return true
	}
	return false
}
