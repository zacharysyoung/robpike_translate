package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/translate"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

type TranslateClient struct{ c *translate.Client }

// NewTranslateClient creates an authenticated *[translate.Client].
//
// It reads your authentication info from the environment,
// presently at least a credentials.json file, and optionally
// a cached token.json file.
func NewTranslateClient() (*TranslateClient, error) {
	httpClient, err := newHTTPClient()
	if err != nil {
		return nil, err
	}

	c, err := translate.NewClient(
		context.Background(),
		option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	return &TranslateClient{c}, nil
}

// doTranslations translates lines of text to target language.
//
// Use the two-letter BCP 47 language codes for source
// and target; source can be left empty to force the
// translator to auto-detect the language; target cannot be empty.
func (tc *TranslateClient) Translate(target, source string, lines []string) ([]translate.Translation, error) {
	tgt := language.Make(target)
	src := language.Make(source)

	if tgt == language.Und {
		return nil, fmt.Errorf("could not parse target language tag %s; double-check the IETF BCP 47 language tag specificication", target)
	}

	return tc.c.Translate(
		context.Background(),
		lines,
		tgt,
		&translate.Options{
			Source: src,
			Format: translate.Text,
		},
	)
}

// newHTTPClient handles all aspects of reading secrets and
// token JSON, and initiating a web-based OAuth2 auth flow.
func newHTTPClient() (*http.Client, error) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("could not read secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, translate.Scope)
	if err != nil {
		return nil, fmt.Errorf("could not create config: %v", err)
	}

	const tokenJSON = "token.json"
	var token *oauth2.Token

	b, err = os.ReadFile(tokenJSON)
	switch {
	default:
		return nil, fmt.Errorf("unexpected error reading token JSON: %v", err)
	case err == nil:
		err = json.Unmarshal(b, &token)
	case errors.Is(err, os.ErrNotExist):
		token, err = userGetsTokenFromWeb(config)
	}
	if err != nil {
		return nil, err
	}

	b, err = json.Marshal(token)
	if err == nil {
		os.WriteFile(tokenJSON, b, 0600)
	}

	client := config.Client(context.Background(), token)

	return client, nil
}

// userGetsTokenFromWeb initiates the OAuth2 auth flow, prompting
// the user to complete the process in their browser; it listens
// for the auth-code redirect and exchanges that code for a token.
func userGetsTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	const (
		port         = ":8999"
		redirectPath = "/redirect"
	)

	state := getRandState()

	config.RedirectURL = "http://localhost" + port + redirectPath

	authURL := config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline)

	browser.OpenURL(authURL)

	code, err := listenForAuthCode(
		port,
		redirectPath,
		state,
		30*time.Second)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	token, err := config.Exchange(ctx, code)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, errors.New("code exchange timed out")
	}

	return token, err
}

func getRandState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
