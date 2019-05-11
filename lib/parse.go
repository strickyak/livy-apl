package livy

/*
	“Should array indices start at 0 or 1?
	 My compromise of 0.5 was rejected without, I thought, proper consideration.”
	  — Stan Kelly-Bootle
	http://exple.tive.org/blarg/2013/10/22/citation-needed/

	"So let us let our ordinals start at zero: an element's ordinal (subscript)
	 equals the number of elements preceding it in the sequence."
	  — Edsger W. Dijkstra
	https://www.cs.utexas.edu/users/EWD/transcriptions/EWD08xx/EWD831.html

	"I used to describe [Pascal] as a ‘fascist programming language’,
	 because it is dictatorially rigid. …
	 If Pascal is fascist, APL is anarchist."
	  — Brad McCormick
	http://www.computerhistory.org/atchm/the-apl-programming-language-source-code/#footnote-13
	http://www.users.cloud9.net/~bradmcc/APL.html

	"APL is a mistake, carried through to perfection. It is the language
	 of the future for the programming techniques of the past: it creates a
	 new generation of coding bums."
	  — Edsger W. Dijkstra
	http://www.computerhistory.org/atchm/the-apl-programming-language-source-code/#footnote-14
	https://www.cs.virginia.edu/~evans/cs655/readings/ewd498.html
*/

import (
	"strconv"
)

type Parser struct {
	Context *Context
}

func (p *Parser) ParseBracket(lex *Lex, i int) ([]Expression, int) {
	i++
	var vec []Expression
	var tmp Expression
	for {
		switch lex.Tokens[i].Type {
		case CloseSquareToken:
			vec = append(vec, tmp)
			i++
			return vec, i
		case SemiToken:
			i++
			vec = append(vec, tmp)
			tmp = nil
		default:
			// This is a bit weak.
			tmp, i = p.ParseExpr(lex, i)
		}
	}
}

func (p *Parser) ParseSeq(lex *Lex, i int) (*Seq, int) {
	tt := lex.Tokens
	var vec []Expression
LOOP:
	for i < len(tt) && tt[i].Type != EndToken {
		Log.Printf("ParseSeq: i=%d max=%d token=%s", i, len(tt), tt[i])
		b, j := p.ParseExpr(lex, i)
		Log.Printf("ParseSeq: i=%d b=%s", i, b)
		vec = append(vec, b)
		i = j

		switch tt[i].Type {
		case KeywordToken:
			switch tt[i].Str {
			case "then", "else", "fi", "do", "done":
				break LOOP
			default:
				Log.Panicf("unexpected keyword: %q", tt[i].Str)
			}
		case EndToken, CloseCurlyToken:
			break LOOP
		case SemiToken:
			i++
			continue LOOP
		default:
			Log.Panicf("default: %d %s", i, tt[i])
		}
	}

	return &Seq{vec}, i
}

func (p *Parser) ParseWhile(lex *Lex, i int) (*While, int) {
	tt := lex.Tokens
	t := tt[i]

	whileSeq, j := p.ParseSeq(lex, i)
	i = j

	t = tt[i]
	if t.Str != "do" {
		Log.Panicf("expected `do` but got %q", t.Str)
	}

	i++
	t = tt[i]
	doSeq, j := p.ParseSeq(lex, i)
	i = j
	t = tt[i]
	if t.Str != "done" {
		Log.Panicf("expected `done` but got %q", t.Str)
	}
	z := &While{whileSeq, doSeq}
	Log.Printf("ParseWhile returns %v", z)
	return z, i + 1
}

func (p *Parser) ParseIf(lex *Lex, i int) (*Cond, int) {
	tt := lex.Tokens
	t := tt[i]

	ifSeq, j := p.ParseSeq(lex, i)
	i = j

	t = tt[i]
	if t.Str != "then" {
		Log.Panicf("expected `then` but got %q", t.Str)
	}

	i++
	t = tt[i]
	thenSeq, j := p.ParseSeq(lex, i)
	i = j
	t = tt[i]
	if t.Str != "else" {
		Log.Panicf("expected `else` but got %q", t.Str)
	}
	i++
	t = tt[i]
	elseSeq, j := p.ParseSeq(lex, i)
	i = j
	t = tt[i]
	if t.Str != "FI" {
		Log.Panicf("expected `FI` but got %q", t.Str)
	}
	z := &Cond{ifSeq, thenSeq, elseSeq}
	Log.Printf("ParseIf returns %v", z)
	return z, i + 1
}

func (p *Parser) ParseDef(lex *Lex, i int) (*Def, int) {
	var lhs, axis, rhs string
	var locals []string

	tt := lex.Tokens
	t := tt[i]
	// Expect operator.
	if t.Type == VariableToken {
		lhs = t.Str
		locals = append(locals, lhs)
		i++
		t = tt[i]
	}

	if t.Type != OperatorToken {
		Log.Panicf("expected operator after def, but got %v", t)
	}
	name := t.Str
	i++
	t = tt[i]
	if t.Type == OpenSquareToken {
		i++
		t = tt[i]
		if t.Type != VariableToken {
			Log.Panicf("expected AXIS variable after def operator open-bracket, but got %v", t)
		}
		axis = t.Str
		locals = append(locals, axis)
		i++
		t = tt[i]
		if t.Type != CloseSquareToken {
			Log.Panicf("expected close-bracket def operator open-bracket axis, but got %v", t)
		}
		i++
		t = tt[i]
	}
	if t.Type != VariableToken {
		Log.Panicf("expected RHS variable after def, but got %v", t)
	}
	rhs = t.Str
	locals = append(locals, rhs)
	i++
	t = tt[i]

	for t.Type == SemiToken {
		i++
		t = tt[i]
		// Allow extra semicolon before open curly.
		if t.Type == OpenCurlyToken {
			break
		}
		if t.Type != VariableToken {
			Log.Panicf("expected local variable name after def semicolon, but got %v", t)
		}
		locals = append(locals, t.Str)
		i++
		t = tt[i]
	}

	if t.Type != OpenCurlyToken {
		Log.Panicf("expected open-curly-brace after operator after def, but got %v", t)
	}
	i++

	seq, j := p.ParseSeq(lex, i)
	i = j
	t = tt[i]

	if t.Type != CloseCurlyToken {
		Log.Panicf("expected close-curly-brace after operator after def, but got %v", t)
	}
	i++
	return &Def{name, seq, lhs, axis, rhs, locals}, i
}

func (p *Parser) ParseExpr(lex *Lex, i int) (z Expression, zi int) {
	tt := lex.Tokens
	var vec []Expression
LOOP:
	for {
		t := tt[i]
		Log.Printf("........ [%d] %q %v", i, t.Str, t)
		switch t.Type {
		case KeywordToken:
			switch t.Str {
			case "break":
				vec = append(vec, BREAK)
				i++
			case "continue":
				vec = append(vec, CONTINUE)
				i++
			case "def":
				def, j := p.ParseDef(lex, i+1)
				vec = append(vec, def)
				i = j
			case "if":
				cond, j := p.ParseIf(lex, i+1)
				vec = append(vec, cond)
				i = j
			case "while":
				while, j := p.ParseWhile(lex, i+1)
				vec = append(vec, while)
				i = j
			case "then", "else", "fi", "do", "done":
				break LOOP
			default:
				Log.Panicf("initial keyword not implemented: %q", t.Str)
			}
		case EndToken, CloseToken, CloseSquareToken, SemiToken, CloseCurlyToken:
			break LOOP
		case OpenSquareToken:
			Log.Panicf("Unexpected `[` at position %d: %s", t.Pos, lex.Source)
		case OpenCurlyToken:
			Log.Panicf("Unexpected `{` at position %d: %s", t.Pos, lex.Source)
		case EachToken, ScanToken, ReduceToken, InnerProductToken, OuterProductToken, OperatorToken:
			axis := Expression(nil)
			var j int
			if tt[i+1].Type == OpenSquareToken {
				Log.Printf("Axis1")
				axis, j = p.ParseExpr(lex, i+2)
				Log.Printf("Axis2 %d %s", j, axis)
				if tt[j].Type != CloseSquareToken {
					Log.Panicf("Expected ']' but got %q after subscript", tt[i].Str)
				}
				i = j // Don't add 1 here; ParseExpr just below gets i+1.
			}

			Log.Printf("===== PE [%d]", i+1)
			b, j := p.ParseExpr(lex, i+1)
			Log.Printf("===== PE [%d] --> %v %d", i+1, b, j)
			switch len(vec) {
			case 0:
				return &Monad{t, t.Str, b, axis}, j
			case 1:
				return &Dyad{t, vec[0], t.Str, b, axis}, j
			default:
				return &Dyad{t, &List{vec}, t.Str, b, axis}, j
			}
		case ComplexToken:
			{
				_m := MatchComplexSplit(t.Str)
				if _m == nil {
					Log.Panicf("Error parsing ComplexToken %q at position %d: %s", t.Str, t.Pos, lex.Source)
				}
				_r, _j, _i := _m[1], _m[2], _m[3]
				var rl float64
				if _r != "" {
					_rl, err := strconv.ParseFloat(_r, 64)
					if err != nil {
						Log.Panicf("Error parsing number %q at position %d: %s", t.Str, t.Pos, lex.Source)
					}
					rl = _rl
				}
				cx, err := strconv.ParseFloat(_i, 64)
				if err != nil {
					Log.Panicf("Error parsing number %q at position %d: %s", t.Str, t.Pos, lex.Source)
				}
				if _j[0] == '-' {
					cx = -cx
				}
				vec = append(vec, &Number{complex(rl, cx)})
				i++
			}
		case NumberToken:
			num, err := strconv.ParseFloat(t.Str, 64)
			if err != nil {
				Log.Panicf("Error parsing number %q at position %d: %s", t.Str, t.Pos, lex.Source)
			}
			vec = append(vec, &Number{complex(num, 0)})
			i++
		case StringToken:
			s, err := strconv.Unquote(t.Str)
			if err != nil {
				Log.Panicf("Error parsing string %s at position %d: %s", t.Str, t.Pos, lex.Source)
			}
			if p.Context.StringExtension == nil {
				Log.Panicf("No StringExtension in this interpreter")
			}
			vec = append(vec, p.Context.StringExtension(s))
			i++
		case VariableToken:
			variable := &Variable{t.Str}
			i++
			if tt[i].Type == OpenSquareToken {
				Log.Printf("B1")
				subs, j := p.ParseBracket(lex, i)
				Log.Printf("B2 %d %s", j, subs)
				vec = append(vec, &Subscript{variable, subs})
				Log.Printf("B3 %s", vec)
				i = j
			} else {
				vec = append(vec, variable)
			}
		case OpenToken:
			expr, j := p.ParseExpr(lex, i+1)
			i = j + 1
			// Allow brackets after parens e.g. (iota1 10)[2 4 6]
			if tt[i].Type == OpenSquareToken {
				Log.Printf("B1")
				subs, j := p.ParseBracket(lex, i)
				Log.Printf("B2 %d %s", j, subs)
				vec = append(vec, &Subscript{expr, subs})
				Log.Printf("B3 %s", vec)
				i = j
			} else {
				vec = append(vec, expr)
			}
		default:
			Log.Panicf("bad default: %d", t.Type)
		}
	}

	if len(vec) == 0 {
		Log.Panicf("Error parsing expression; perhaps an operator followed by no expression: %q %q", tt[i-1].Str, tt[i].Str)
	}
	if len(vec) > 1 {
		return &List{vec}, i
	}
	Log.Printf("ParseExpr returns VEC=%v; i=%d", vec, i)
	return vec[0], i
}
