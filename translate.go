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

const usageText = `usage: translate [-to] [-from] [-key] TEXT [...TEXT]

The program translates the text arguments with Google's
Translate.V2 API and prints to stdout.

The program requires an API key for a Google Cloud Project (GCP)
that has enabled the Cloud Translation API.

-to string
  	target language (two-letter BCP-47 code) (default "en")
-from string
  	source language (two-letter BCP-47 code); auto-detected by default
-key string
  	API key for translation-enabled GCP (default $GOOGLEAPIKEY)

`

var (
	targetFlag = flag.String("to", "en", "")
	sourceFlag = flag.String("from", "", "")
	keyFlag    = flag.String("key", "", "")
)

func main() {
	flag.Usage = func() { fatalf("%s", usageText) }
	flag.Parse()

	if len(flag.Args()) == 0 {
		fatalf("no text arguments; run translate -h")
	}

	if *keyFlag == "" {
		*keyFlag = os.Getenv("GOOGLEAPIKEY")
	}
	if *keyFlag == "" {
		fatalf("empty API key; run translate -h")
	}

	tgtLang := language.Make(*targetFlag)
	srcLang := language.Make(*sourceFlag)
	if tgtLang == language.Und {
		fatalf("could not parse target language tag %s; double-check the IETF BCP 47 language tag specificication", *targetFlag)
	}

	ctx := context.Background()

	client, err := translate.NewClient(ctx, option.WithAPIKey(*keyFlag))
	if err != nil {
		fatalf("could not create translate client: %v", err)
	}

	translations, err := client.Translate(
		ctx,
		flag.Args(),
		tgtLang,
		&translate.Options{
			Source: srcLang, Format: translate.Text})
	if err != nil {
		fatalf("%v", errors.Unwrap(err)) // experimentation showed that unwrapping provided a good one-line error message
	}

	for _, x := range translations {
		fmt.Println(x.Text)
	}
}

func fatalf(format string, args ...any) {
	if !strings.HasPrefix(format, "error: ") {
		format = "error: " + format
	}
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(2)
}
