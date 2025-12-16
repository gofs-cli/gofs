package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gofs-cli/gofs/templates/fs-app/internal/config"
	"github.com/gofs-cli/gofs/templates/fs-app/internal/db"
	"github.com/gofs-cli/gofs/templates/fs-app/internal/repository"
)

type Server struct {
	conf config.Config
	r    *http.ServeMux
	srv  http.Server
	// repo is an sqlc generated repository for executing queries
	repo *repository.Queries
	// conn is a handle to the database connection used for the repository
	//
	// The handle is part of the server to allow closing the connection on shutdown
	conn *sql.DB
}

func New(conf config.Config) (*Server, error) {
	s := new(Server)
	s.conf = conf
	s.r = http.NewServeMux()
	s.srv = http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Handler:      s.r,
	}
	var err error
	s.conn, err = db.New()
	if err != nil {
		return nil, fmt.Errorf("db: opening connection: %w", err)
	}
	s.repo = repository.New(s.conn)

	err = db.MigrateTables(s.conn)
	if err != nil {
		return nil, fmt.Errorf("db: migrating tables: %w", err)
	}

	return s, nil
}

func (s *Server) ListenAndServe() error {
	s.Routes()
	// address for use when testing cookies locally
	if s.conf.Host == "0.0.0.0" {
		log.Printf("Server: listening on http://localhost:%s", s.conf.Port)
	} else {
		log.Printf("Server: listening on http://%s", s.srv.Addr)
	}
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	var err error

	// Shutdown the HTTP server gracefully
	if shutdownErr := s.srv.Shutdown(ctx); shutdownErr != nil {
		err = fmt.Errorf("server shutdown: %w", shutdownErr)
	}

	// Close database connection
	if closeErr := s.conn.Close(); closeErr != nil {
		if err != nil {
			err = fmt.Errorf("%w; db close: %v", err, closeErr)
		} else {
			err = fmt.Errorf("db close: %w", closeErr)
		}
	}

	return err
}
