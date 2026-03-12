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


-credsPath string
  	path to creds.json; (defaults to $TRANSLATE_CREDSJSON);
  	MUST be set in env or by flag
-tokenPath string
  	path to cached token.json; (defaults to $TRANSLATE_TOKENJSON)
`

var (
	targetFlag = flag.String("to", "en", "")
	sourceFlag = flag.String("from", "", "")

	credsFlag = flag.String("credsPath", "", "")
	tokenFlag = flag.String("tokenPath", "", "")
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) < 1 {
		usage()
	}

	if *credsFlag == "" {
		*credsFlag = os.Getenv("TRANSLATE_CREDSJSON")
	}
	if *credsFlag == "" {
		usage()
	}

	if *tokenFlag == "" {
		*tokenFlag = os.Getenv("TRANSLATE_TOKENJSON")
	}

	c, err := NewTranslateClient()
	if err != nil {
		fatalf("error: %v", err)
	}

	translations, err := c.Translate(
		*targetFlag,
		*sourceFlag,
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
