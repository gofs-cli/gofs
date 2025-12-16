package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gofs-cli/gofs/templates/fs-app/internal/config"
	"github.com/gofs-cli/gofs/templates/fs-app/internal/server"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	log.Println("Go version:", runtime.Version())
	log.Println("Go OS/Arch:", runtime.GOOS, runtime.GOARCH)

	conf := config.New()
	srv, err := server.New(conf)
	if err != nil {
		return fmt.Errorf("server initialization error: %w", err)
	}

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server startup error: %v", err)
		}
		log.Println("stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}
	log.Println("server shutdown complete.")
	return nil
}
