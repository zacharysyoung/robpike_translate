package main

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

// listenForAuthCode starts an ad-hoc localhost server that
// listens for the redirect call in the auth flow process and
// returns the auth code.
//
// It blocks until the redirect handler receives the request,
// timout expires, or the user interrupts with Ctrl-C.
//
// The redirect handler will also compare state to the received
// state query param in the request.
func listenForAuthCode(port, redirectPath, state string, timeout time.Duration) (code string, err error) {
	var (
		handledRedirect = make(chan bool)
		timedOut        = time.Tick(timeout)
		cancelledByUser = make(chan os.Signal, 1)

		errWaiting = make(chan error, 1)
	)

	signal.Notify(cancelledByUser, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	http.HandleFunc(redirectPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		code = q.Get("code")
		if q.Get("state") != state {
			err = errors.New("handling redirect, got invalid state")
		}
		handledRedirect <- true
	})

	srv := http.Server{Addr: port}

	go waitToShutdown(&srv, handledRedirect, timedOut, cancelledByUser, errWaiting)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		err = fmt.Errorf("problem with server: %v", err)
	}

	if err := <-errWaiting; err != nil {
		return "", err
	}

	return code, err
}

// waitToShutdown blocks until one of the three channels
// receives a value, then gracefully shuts down srv, and
// reports any errors on errWaiting.
func waitToShutdown(
	srv *http.Server,

	handledRedirect <-chan bool,
	timedOut <-chan time.Time,
	cancelledByUser <-chan os.Signal,

	errWaiting chan<- error,
) {
	var earlyErr error
	select {
	case <-handledRedirect:
		break
	case <-timedOut:
		earlyErr = errors.New("timed out waiting for redirect")
	case <-cancelledByUser:
		fmt.Print("\r") // clear terminal line of "^C"
		earlyErr = errors.New("user cancelled auth flow")
	}

	var srvErr error
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	if srvErr = srv.Shutdown(ctx); srvErr != nil {
		srvErr = fmt.Errorf("could not cleanly shut down: %v", srvErr)
	}
	cancel()

	var err error
	switch {
	case earlyErr != nil && srvErr != nil:
		err = fmt.Errorf("after %v, %v", earlyErr, srvErr)
	case earlyErr != nil:
		err = earlyErr
	case srvErr != nil:
		err = srvErr
	}
	errWaiting <- err
}
