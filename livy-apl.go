package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime/debug"
	"strings"

	"github.com/chzyer/readline"

	. "github.com/strickyak/livy-apl/livy"
)

var Prompt = flag.String("p", "      ", "APL interpreter prompt")
var Quiet = flag.Bool("q", false, "supress log messages")
var CrashOnError = flag.Bool("e", false, "crash dump on error")

func EvalString(c *Context, line string) (val Val, err error) {
	if !*CrashOnError {
		defer func() {
			r := recover()
			if r != nil {
				if !*Quiet {
					debug.PrintStack()
				}
				err = errors.New(fmt.Sprintf("%v", r))
			}
		}()
	}
	lex := Tokenize(line)
	expr, j := ParseExpr(lex, 0)
	if j != len(lex.Tokens)-1 {
		log.Fatalf("FATAL: Parse unfinished: Got %d expected %d", j, len(lex.Tokens)-1)
	}
	val = expr.Eval(c)
	return
}

type SinkToNowhere struct{}

func (SinkToNowhere) Write(bb []byte) (int, error) {
	return len(bb), nil
}

func main() {
	flag.Parse()
	if *Quiet {
		log.SetOutput(SinkToNowhere{})
	} else {
		log.SetFlags(log.Ltime + log.Lshortfile)
		log.SetPrefix("##")
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "      ",
		HistoryFile:     "/tmp/livy-apl.tmp",
		InterruptPrompt: "*SIGINT*",
		EOFPrompt:       "*EOF*",
		// AutoComplete:    completer,
		// HistorySearchFold:   true,
		// FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	c := &Context{
		Globals:  make(map[string]Val),
		Monadics: StandardMonadics,
		Dyadics:  StandardDyadics,
	}
	c.Globals["Pi"] = &Num{math.Pi}
	c.Globals["Tau"] = &Num{2.0 * math.Pi}
	c.Globals["E"] = &Num{math.E}
	c.Globals["Phi"] = &Num{math.Phi}

	i := 0
	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		result, complaint := EvalString(c, line)
		if complaint != nil {
			fmt.Fprintf(os.Stderr, "*** ERROR *** %s\n", complaint)
		} else {
			name := fmt.Sprintf("_%d", i)
			c.Globals[name] = result
			c.Globals["_"] = result
			fmt.Fprintf(os.Stdout, "%s = %s\n", name, result)
			i++
		}
	}
}
