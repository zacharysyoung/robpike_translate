package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/robpike/translate/auth"
)

const usageText = `usage: translate [-to] [-from] TEXT [...TEXT]

-to string
  	target language (two-letter code) (default "en")
-from string
  	source language (two-letter code); auto-detected by default
`

var (
	target = flag.String("to", "en", "target language (two-letter code)")
	source = flag.String("from", "", "source language (two-letter code); auto-detected by default")
)

func main() {
	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) < 1 {
		usage()
	}

	c, err := auth.NewClient()
	if err != nil {
		fatalf("error: %v", err)
	}

	translations, err := c.Translate(*target, *source, flag.Args())
	if err != nil {
		fatalf("error: %v", errors.Unwrap(err)) // experimentation showed that unwrapping provided a good one-line error message
	}

	for _, x := range translations {
		fmt.Println(x.Text)
	}
}

func usage() {
	fatalf("%s", usageText)
}

func fatalf(format string, args ...any) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(2)
}
