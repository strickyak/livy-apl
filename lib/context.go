package livy

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
)

var Log *log.Logger = log.New(os.Stderr, "livy: ", 0)

type StringExtensionFunc func(s string) Expression

type Context struct {
	Globals    map[string]Val
	Monadics   map[string]MonadicFunc
	Dyadics    map[string]DyadicFunc
	LocalStack []map[string]Val

	FormatReal string
	FormatImagPlus string
	FormatImagMinus string
	FormatComplexPlus string
	FormatComplexMinus string

	StringExtension StringExtensionFunc
	Extra           map[string]interface{}
}

func NewContext() *Context {
	c := &Context{
		Globals:  make(map[string]Val),
		Monadics: StandardMonadics,
		Dyadics:  StandardDyadics,
		Extra:    make(map[string]interface{}),
		FormatReal: "%g",
		FormatImagPlus: "+j%g",
		FormatImagMinus: "-j%g",
		FormatComplexPlus: "%g-j%g",
		FormatComplexMinus: "%g+j%g",
	}
	c.Globals["Pi"] = &Num{math.Pi}
	c.Globals["Tau"] = &Num{2.0 * math.Pi}
	c.Globals["E"] = &Num{math.E}
	c.Globals["Phi"] = &Num{math.Phi}
	c.Globals["J"] = &Num{complex(0, 1)}
	return c
}

func (c *Context) Command(s string) {
	if s == "" {
		s = "?"
	}

	switch s[0] {
	case 'v':
		var names []string
		maxLen := 0
		for k, _ := range c.Globals {
			names = append(names, k)
			if len(k) > maxLen {
				maxLen = len(k)
			}
		}
		sort.Strings(names)
		for _, k := range names {
			if k[0] == '_' {
				continue // Skip _ variables
			}
			format := fmt.Sprintf("%%%ds : %%s\n", maxLen)
			fmt.Fprintf(os.Stderr, format, k, c.Globals[k])
		}
	case 'm':
		var names []string
		for k, _ := range c.Monadics {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			fmt.Fprintf(os.Stderr, "%s ", k)
		}
		fmt.Fprintf(os.Stderr, "\n")
	case 'd':
		var names []string
		for k, _ := range c.Dyadics {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			fmt.Fprintf(os.Stderr, "%s ", k)
		}
		fmt.Fprintf(os.Stderr, "\n")
	default:
		fmt.Fprintf(os.Stderr, `Unknown command.

Commands:  )v[ars]  )m[onadics]  )d[yadics]

`)
		return
	}
}
