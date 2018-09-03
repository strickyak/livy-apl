package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/chzyer/readline"

	_ "github.com/strickyak/livy-apl/extend"
	. "github.com/strickyak/livy-apl/livy"
)

var Prompt = flag.String("p", "      ", "APL interpreter prompt")
var Quiet = flag.Bool("q", false, "supress log messages")
var CrashOnError = flag.Bool("e", false, "crash dump on error")
var Raw = flag.Bool("r", false, "print raw results")

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
	expr, _ := ParseSeq(lex, 0)
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

	home := os.Getenv("HOME")
	if home == "" {
		home = "."
	}
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "      ",
		HistoryFile:     filepath.Join(home, ".livy.history"),
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

		if strings.HasPrefix(line, ")") {
			c.Command(line[1:])
			continue
		}

		result, complaint := EvalString(c, line)
		if complaint != nil {
			fmt.Fprintf(os.Stderr, "****** ERROR: %s\n", complaint)
			continue
		}

		name := fmt.Sprintf("_%d", i)
		c.Globals[name] = result
		c.Globals["_"] = result
		if *Raw {
			fmt.Fprintf(os.Stdout, "%s = %s\n", name, result)
		} else {
			bb := bytes.NewBuffer(nil)
			shape := result.Shape()
			if len(shape) > 0 {
				for _, x := range shape {
					fmt.Fprintf(bb, "%d ", x)
				}
				fmt.Fprintf(bb, "rho")
			}
			fmt.Fprintf(os.Stdout, "   %s = %s\n%s\n", name, bb.String(), result.Pretty())
		}
		i++
	}
}
