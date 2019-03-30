package livy

import (
	"fmt"
	"regexp"
)

type TokenType int

const (
	EndToken TokenType = iota
	ComplexToken
	NumberToken
	VariableToken
	OperatorToken
	OpenToken
	CloseToken
	BraToken
	KetToken
	SemiToken
	InnerProductToken
	OuterProductToken
	OpenCurlyToken
	CloseCurlyToken
	KeywordToken
	ReduceToken
	ScanToken
	EachToken
	StringToken
)

const RE_JUST_OPERATOR = `([-+*/\\,&|!=<>]+|[a-z][A-Za-z0-9_]*)`
const RE_OPERATOR = `([-+*/\\,&|!=<>]+|[a-z][A-Za-z0-9_]*[/\\]?)`
const RE_KEYWORD = `(DEF|IF|THEN|ELIF|ELSE|FI|WHILE|DO|DONE|BREAK|CONTINUE)\b`
const RE_REAL = `([-+]?[0-9]+([.][0-9]+)?([eE][-+]?[0-9]+)?)`
const RE_COMPLEX = RE_REAL + `?([+-][jJ])` + RE_REAL
const RE_COMPLEX_SPLIT = `(.*)([+-][jJ])(.*)`

var MatchWhite = regexp.MustCompile(`^([ \t\r]*)`).FindStringSubmatch
var MatchNumber = regexp.MustCompile(`^` + RE_REAL).FindStringSubmatch
var MatchComplex = regexp.MustCompile(`^` + RE_COMPLEX).FindStringSubmatch
var MatchComplexSplit = regexp.MustCompile(RE_COMPLEX_SPLIT).FindStringSubmatch
var MatchVariable = regexp.MustCompile(`^([A-Z_][A-Za-z0-9_]*)`).FindStringSubmatch
var MatchOperator = regexp.MustCompile("^" + RE_OPERATOR).FindStringSubmatch
var MatchOpen = regexp.MustCompile(`^[(]`).FindStringSubmatch
var MatchClose = regexp.MustCompile(`^[)]`).FindStringSubmatch
var MatchOpenCurly = regexp.MustCompile(`^[{]`).FindStringSubmatch
var MatchCloseCurly = regexp.MustCompile(`^[}]`).FindStringSubmatch
var MatchBra = regexp.MustCompile(`^[[]`).FindStringSubmatch
var MatchKet = regexp.MustCompile(`^[]]`).FindStringSubmatch
var MatchSemi = regexp.MustCompile(`^[;\n]`).FindStringSubmatch
var MatchString = regexp.MustCompile(`^(["]([^"\\]|[\\].)*["])`).FindStringSubmatch

var MatchReduce = regexp.MustCompile("^" + RE_JUST_OPERATOR + `[/]`).FindStringSubmatch
var MatchScan = regexp.MustCompile("^" + RE_JUST_OPERATOR + `[\\]`).FindStringSubmatch
var MatchEach = regexp.MustCompile("^" + RE_JUST_OPERATOR + `[~]`).FindStringSubmatch

var MatchInnerProduct = regexp.MustCompile("^" + RE_JUST_OPERATOR + "[.]" + RE_JUST_OPERATOR).FindStringSubmatch
var MatchOuterProduct = regexp.MustCompile("^[.][.]" + RE_JUST_OPERATOR).FindStringSubmatch
var MatchKeyword = regexp.MustCompile("^" + RE_KEYWORD).FindStringSubmatch

type Matcher struct {
	Type    TokenType
	MatchFn func(string) []string
}

var matchers = []Matcher{
	{KeywordToken, MatchKeyword},
	{ComplexToken, MatchComplex},
	{NumberToken, MatchNumber},
	{VariableToken, MatchVariable},
	{ReduceToken, MatchReduce},
	{ScanToken, MatchScan},
	{EachToken, MatchEach},
	{InnerProductToken, MatchInnerProduct},
	{OuterProductToken, MatchOuterProduct},
	{OperatorToken, MatchOperator},
	{OpenToken, MatchOpen},
	{CloseToken, MatchClose},
	{OpenCurlyToken, MatchOpenCurly},
	{CloseCurlyToken, MatchCloseCurly},
	{BraToken, MatchBra},
	{KetToken, MatchKet},
	{SemiToken, MatchSemi},
	{StringToken, MatchString},
}

type Token struct {
	Type  TokenType
	Str   string
	Pos   int
	Match []string
}

func (t Token) String() string {
	return fmt.Sprintf("T(%d,%q,%d,%#v)", t.Type, t.Str, t.Pos, t.Match)
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
	// Log.Printf("TOKENIZE %q", s)
	lex := &Lex{
		Source: s,
	}
	for lex.DoNextToken() {
		// Log.Printf("LEX... %s", *lex)
		continue
	}
	// Log.Printf("LEX... %s", *lex)

	llt := len(lex.Tokens)
	if llt == 0 || lex.Tokens[llt-1].Type != EndToken {
		Log.Panicf("Syntax error after %q before %q", s[:lex.p], s[lex.p:])
	}
	if lex.p != len(s) {
		Log.Panicf("OHNO did not parse all of %q: %d", s, lex.p)
	}
	for i, t := range lex.Tokens {
		Log.Printf("Token [%d]: %s", i, t)
	}
	return lex
}

func (lex *Lex) DoNextToken() bool {
	// Skip white space.
	mw := MatchWhite(lex.Source[lex.p:])
	if mw != nil {
		lex.p += len(mw[0])
	}

	// Check for end of string.
	if lex.p == len(lex.Source) {
		t := &Token{
			Type: EndToken,
			Str:  "",
			Pos:  lex.p,
		}
		lex.Tokens = append(lex.Tokens, t)
		return false
	}

	// Try each matcher until one works.
	for _, matcher := range matchers {
		m := matcher.MatchFn(lex.Source[lex.p:])
		if m != nil {
			lex.Tokens = append(lex.Tokens, &Token{
				Type:  matcher.Type,
				Str:   m[0],
				Pos:   lex.p,
				Match: m,
			})
			lex.p += len(m[0])
			return true
		}
	}

	// Or else we have a parse error.
	return false
}
