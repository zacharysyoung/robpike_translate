package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

const usageText = `usage: translate [-to] [-from] TEXT [...TEXT]

-to string
  	target language (two-letter BCP-47 code) (default "en")
-from string
  	source language (two-letter BCP-47 code); auto-detected by default
`

var (
	target = flag.String("to", "en", "")
	source = flag.String("from", "", "")
)

func main() {
	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) < 1 {
		usage()
	}

	c, err := NewTranslateClient()
	if err != nil {
		fatalf("error: %v", err)
	}

	translations, err := c.Translate(
		*target,
		*source,
		flag.Args())
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
