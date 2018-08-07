package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"

	"github.com/chzyer/readline"

	. "github.com/strickyak/livy-apl/livy"
)

var Prompt = flag.String("p", "      ", "APL interpreter prompt")

func Run(c *Context, line string) (val Val, complaint string) {
	defer func() {
		r := recover()
		if r != nil {
			complaint = fmt.Sprintf("%v", r)
		}
	}()
	lex := Tokenize(line)
	expr, _ := Parse(lex, 0)
	val = expr.Eval(c)
	return
}

func main() {
	flag.Parse()

	rl, err := readline.NewEx(&readline.Config{
		Prompt:      "(>>>)",
		HistoryFile: "/tmp/livy-apl.tmp",
		//AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

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

	// r := bufio.NewReader(os.Stdin)
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
		log.Printf("<<< %q >>>", line)

		/*
			fmt.Fprintf(os.Stderr, "%s", *Prompt)

			line, err := r.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Cannot read stdin: %s", err)
			}
		*/

		result, complaint := Run(c, line)
		if complaint != "" {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", complaint)
		} else {
			name := fmt.Sprintf("_%d", i)
			c.Globals[name] = result
			c.Globals["_"] = result
			fmt.Fprintf(os.Stdout, "%s = %s\n", name, result)
			i++
		}
	}
}
