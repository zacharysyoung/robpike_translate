package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

const usageText = `usage: translate [-to] [-from] [-credsPath] TEXT [...TEXT]

The text arguments will be joined with a space, translated
with Google's Translate/v3 API, and printed to stdout.  It
uses the API's default translation models.

This program requires a Service Account for a Google Cloud
Project that has enabled the Translate API, and the down-
loaded credentials file for that account.

-to string
  	target language (two-letter BCP-47 code) (default "en")
-from string
  	source language (two-letter BCP-47 code); auto-detected by default

-credsPath string
  	path to service account credentials file; (defaults to
  	$TRANSLATE_SVCACCTCREDSFILE); MUST be set in env, or by this flag
`

var (
	targetFlag = flag.String("to", "en", "")
	sourceFlag = flag.String("from", "", "")
	credsFlag  = flag.String("credsPath", "", "")
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) == 0 {
		fatalf("error: no text arguments; run translate -h")
	}

	if *credsFlag == "" {
		*credsFlag = os.Getenv("TRANSLATE_SVCACCTCREDSFILE")
	}
	if *credsFlag == "" {
		fatalf("error: empty path to service account credentials file; run translate -h")
	}

	tgtLang := language.Make(*targetFlag)
	srcLang := language.Make(*sourceFlag)
	if tgtLang == language.Und {
		fatalf("error: could not parse target language tag %s; double-check the IETF BCP 47 language tag specificication", *targetFlag)
	}

	ctx := context.Background()

	client, err := translate.NewClient(
		ctx, option.WithAuthCredentialsFile(
			option.ServiceAccount, *credsFlag))
	if err != nil {
		fatalf("error: could not create translate client: %v", err)
	}

	translations, err := client.Translate(
		ctx,
		flag.Args(),
		tgtLang,
		&translate.Options{
			Source: srcLang,
			Format: translate.Text,
		},
	)
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
