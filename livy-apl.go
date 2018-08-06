package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"

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

	c := &Context{
		Globals:  make(map[string]Val),
		Monadics: StandardMonadics,
		Dyadics:  StandardDyadics,
	}
	c.Globals["Pi"] = &Num{math.Pi}
	c.Globals["Tau"] = &Num{2.0 * math.Pi}
	c.Globals["E"] = &Num{math.E}
	c.Globals["Phi"] = &Num{math.Phi}

	r := bufio.NewReader(os.Stdin)
	i := 0
	for {
		fmt.Fprintf(os.Stderr, "%s", *Prompt)

		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Cannot read stdin: %s", err)
		}

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
