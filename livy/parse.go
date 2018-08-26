package livy

// “Should array indices start at 0 or 1?
// My compromise of 0.5 was rejected without, I thought, proper consideration.”
//      — Stan Kelly-Bootle
//   — http://exple.tive.org/blarg/2013/10/22/citation-needed/

// "So let us let our ordinals start at zero: ..."
//   — Edsger W. Dijkstra

// I used to describe [Pascal] as a ‘fascist programming language’,
// because it is dictatorially rigid. …
// If Pascal is fascist, APL is anarchist.
//   — Brad McCormick
// http://www.computerhistory.org/atchm/the-apl-programming-language-source-code/#footnote-13
// http://www.users.cloud9.net/~bradmcc/APL.html

// APL is a mistake, carried through to perfection. It is the language
// of the future for the programming techniques of the past: it creates a
// new generation of coding bums.
//   — Edsger W. Dijkstra
// http://www.computerhistory.org/atchm/the-apl-programming-language-source-code/#footnote-14
// https://www.cs.virginia.edu/~evans/cs655/readings/ewd498.html

import (
	"log"
	"strconv"
)

func ParseBracket(lex *Lex, i int) ([]Expression, int) {
	i++
	var vec []Expression
	var tmp Expression
	for {
		switch lex.Tokens[i].Type {
		case KetToken:
			vec = append(vec, tmp)
			i++
			return vec, i
		case SemiToken:
			i++
			vec = append(vec, tmp)
			tmp = nil
		default:
			// This is a bit weak.
			tmp, i = ParseExpr(lex, i)
		}
	}
}

func ParseSeq(lex *Lex, i int) (*Seq, int) {
	tt := lex.Tokens
	var vec []Expression
LOOP:
	for i < len(tt) && tt[i].Type != EndToken {
		log.Printf("ParseSeq: i=%d max=%d token=%s", i, len(tt), tt[i])
		b, j := ParseExpr(lex, i)
		log.Printf("ParseSeq: i=%d b=%s", i, b)
		vec = append(vec, b)
		i = j

		switch tt[i].Type {
		case KeywordToken:
			switch tt[i].Str {
			case "then", "else", "fi":
				break LOOP
			default:
				log.Fatalf("unexpected keyword: %q", tt[i].Str)
			}
		case EndToken, CloseCurlyToken:
			break LOOP
		case SemiToken:
			i++
			continue LOOP
		default:
			log.Fatalf("default: %d %s", i, tt[i])
		}
	}

	return &Seq{vec}, i
}

func ParseIf(lex *Lex, i int) (*Cond, int) {
	tt := lex.Tokens
	t := tt[i]

	if_seq, j := ParseSeq(lex, i)
	i = j

	t = tt[i]
	if t.Str != "then" {
		log.Fatalf("expected `then` but got %q", t.Str)
	}

	i++
	t = tt[i]
	then_seq, j := ParseSeq(lex, i)
	i = j
	t = tt[i]
	if t.Str != "else" {
		log.Fatalf("expected `else` but got %q", t.Str)
	}
	i++
	t = tt[i]
	else_seq, j := ParseSeq(lex, i)
	i = j
	t = tt[i]
	if t.Str != "fi" {
		log.Fatalf("expected `else` but got %q", t.Str)
	}
	return &Cond{if_seq, then_seq, else_seq}, i + 1
}

func ParseDef(lex *Lex, i int) (*Def, int) {
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
		log.Fatalf("expected operator after def, but got %v", t)
	}
	name := t.Str
	i++
	t = tt[i]
	if t.Type == BraToken {
		i++
		t = tt[i]
		if t.Type != VariableToken {
			log.Fatalf("expected AXIS variable after def operator open-bracket, but got %v", t)
		}
		axis = t.Str
		locals = append(locals, axis)
		i++
		t = tt[i]
		if t.Type != KetToken {
			log.Fatalf("expected close-bracket def operator open-bracket axis, but got %v", t)
		}
		i++
		t = tt[i]
	}
	if t.Type != VariableToken {
		log.Fatalf("expected RHS variable after def, but got %v", t)
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
			log.Fatalf("expected local variable name after def semicolon, but got %v", t)
		}
		locals = append(locals, t.Str)
		i++
		t = tt[i]
	}

	if t.Type != OpenCurlyToken {
		log.Fatalf("expected open-curly-brace after operator after def, but got %v", t)
	}
	i++

	seq, j := ParseSeq(lex, i)
	i = j
	t = tt[i]

	if t.Type != CloseCurlyToken {
		log.Fatalf("expected close-curly-brace after operator after def, but got %v", t)
	}
	i++
	return &Def{name, seq, lhs, axis, rhs, locals}, i
}

func ParseExpr(lex *Lex, i int) (z Expression, zi int) {
	tt := lex.Tokens
	var vec []Expression
LOOP:
	for {
		t := tt[i]
		switch t.Type {
		case KeywordToken:
			switch t.Str {
			case "def":
				def, j := ParseDef(lex, i+1)
				vec = append(vec, def)
				i = j
			case "if":
				def, j := ParseIf(lex, i+1)
				vec = append(vec, def)
				i = j
			case "then", "else", "fi":
				break LOOP
			default:
				log.Panicf("initial keyword not implemented: %q", t.Str)
			}
		case EndToken, CloseToken, KetToken, SemiToken, CloseCurlyToken:
			break LOOP
		case BraToken:
			log.Panicf("Unexpected `[` at position %d: %s", t.Pos, lex.Source)
		case OpenCurlyToken:
			log.Panicf("Unexpected `{` at position %d: %s", t.Pos, lex.Source)
		case InnerProductToken, OuterProductToken, OperatorToken:
			axis := Expression(nil)
			var j int
			if tt[i+1].Type == BraToken {
				log.Printf("Axis1")
				axis, j = ParseExpr(lex, i+2)
				log.Printf("Axis2 %d %s", j, axis)
				if tt[j].Type != KetToken {
					log.Panicf("Expected ']' but got %q after subscript", tt[i].Str)
				}
				i = j // Don't add 1 here; ParseExpr just below gets i+1.
			}

			b, j := ParseExpr(lex, i+1)
			switch len(vec) {
			case 0:
				return &Monad{t, t.Str, b, axis}, j
			case 1:
				return &Dyad{t, vec[0], t.Str, b, axis}, j
			default:
				return &Dyad{t, &List{vec}, t.Str, b, axis}, j
			}
		case NumberToken:
			num, err := strconv.ParseFloat(t.Str, 64)
			if err != nil {
				log.Panicf("Error parsing number %q at position %d: %s", t.Str, t.Pos, lex.Source)
			}
			vec = append(vec, &Number{num})
			i++
		case VariableToken:
			variable := &Variable{t.Str}
			i++
			if tt[i].Type == BraToken {
				log.Printf("B1")
				v, j := ParseBracket(lex, i)
				log.Printf("B2 %d %s", j, v)
				vec = append(vec, &Subscript{variable, v})
				log.Printf("B3 %s", vec)
				i = j
			} else {
				vec = append(vec, variable)
			}
		case OpenToken:
			b, j := ParseExpr(lex, i+1)
			vec = append(vec, b)
			i = j + 1
		default:
			log.Fatalf("bad default: %d", t.Type)
		}
	}

	if len(vec) > 1 {
		return &List{vec}, i
	}
	log.Printf("VEC=%v", vec)
	return vec[0], i
}
