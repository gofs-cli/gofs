package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gofs-cli/gofs/template/internal/config"
	"github.com/gofs-cli/gofs/template/internal/server"
)

func main() {
	log.Println("Go version:", runtime.Version())
	log.Println("Go OS/Arch:", runtime.GOOS, runtime.GOARCH)

	conf := config.New()
	srv, err := server.New(conf)
	if err != nil {
		log.Fatalf("server: %v\n", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v\n", err)
		}
	}()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here, databases etc
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
}
