package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ListenForAuthCode starts an ad-hoc localhost server that
// listens for the redirect call in the auth flow process and
// returns the auth code.
//
// It blocks until the redirect handler receives the request,
// timout expires, or the user interrupts with Ctrl-C.
//
// The redirect handler will also compare state to the received
// state query param in the request.
func ListenForAuthCode(
	port string,
	redirectPath string,
	state string,
	timeout time.Duration,
) (
	code string,
	err error,
) {
	var (
		handledRedirect = make(chan bool)
		timedOut        = time.Tick(timeout)
		userInterrupted = make(chan os.Signal, 1)
	)

	http.HandleFunc(redirectPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		code = q.Get("code")
		if q.Get("state") != state {
			err = errors.New("handling redirect, got invalid state")
		}
		handledRedirect <- true
	})

	srv := http.Server{Addr: port}

	go waitForShutdown(&srv, handledRedirect, timedOut, userInterrupted)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		err = fmt.Errorf("problem with server: %v", err)
	}

	return code, err
}

func waitForShutdown(
	srv *http.Server,
	handledRedirect <-chan bool,
	timedOut <-chan time.Time,
	userInterrupted chan os.Signal,
) error {
	var err error

	signal.Notify(userInterrupted, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-handledRedirect:
		break
	case <-timedOut:
		err = errors.New("timed out")
	case <-userInterrupted:
		fmt.Print("\r") // clear terminal line of "^C"
		err = errors.New("interrupted by user")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		err = fmt.Errorf("trouble shutting down listeners: %v", err)
	}
	cancel()

	return err
}
