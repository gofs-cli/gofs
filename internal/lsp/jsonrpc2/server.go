package jsonrpc2

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

const (
	DefaultTimeout = 2 * time.Second
)

type Handler func(context.Context, chan protocol.Response, protocol.Request) error

type Server struct {
	conn          *Conn
	handlers      map[string]Handler // notification and request events
	isInitialized bool
	isShutdown    bool
	initializer   func(string) error // function to call when rootPath is known
	capabilities  protocol.ServerCapabilities
	active        sync.Map // active requests
}

// return a server with the connection set and a function that takes the rootPath to initialize the lsp
func NewServer(conn *Conn, initializer func(string) error, capabilities protocol.ServerCapabilities) (*Server, error) {
	if conn == nil {
		return nil, fmt.Errorf("no connection")
	}

	return &Server{
		conn:          conn,
		handlers:      make(map[string]Handler),
		isInitialized: false,
		isShutdown:    false,
		initializer:   initializer,
		capabilities:  capabilities,
		active:        sync.Map{},
	}, nil
}

func (s *Server) HandleRequest(method string, handler Handler) {
	s.handlers[method] = handler
}

func (s *Server) ListenAndServe() error {
	slog.Info("server is listening")

	// responses from handlers are written asynchronously using this goroutine
	responseQueue := make(chan protocol.Response)
	defer close(responseQueue)
	go func() {
		for response := range responseQueue {
			if s.isShutdown {
				// the spec says that once shutdown is initiated, the server
				// should not send any more responses
				break
			}
			err := s.conn.Write(response)
			if err != nil {
				slog.Error("error writing response", "err", err)
			}
		}
	}()

	// main loop
	for {
		// read the next request
		request, err := s.conn.Read()
		if err != nil {
			slog.Error("error reading request", "err", err)
			continue
		}

		// handle lifecycle events first synchronously
		switch request.Method {
		case "initialize":
			slog.Info("server is initializing")
			s.Initialize(*request)
			continue
		case "initialized":
			slog.Info("server is initialized")
			s.isInitialized = true // ready to process requests
			continue
		case "shutdown":
			slog.Info("server is shutting down")
			s.Shutdown(*request)
			continue
		case "exit":
			slog.Info("server exit")
			// graceful exit
			return nil
		}

		// do not process requests if lifecycle is before initialization or after shutdown
		if !s.isInitialized || s.isShutdown {
			// requests should error with InvalidRequest according to the spec
			responseQueue <- protocol.NewResponseError(request.Id, protocol.ResponseError{
				Code:    protocol.ErrorCodeInvalidRequest,
				Message: "received request after shutdown",
			})
			continue
		}

		// handle cancel request
		if request.Method == "$/cancelRequest" {
			s.Cancel(*request)
			continue
		}

		// handle unhandled methods
		if s.handlers[request.Method] == nil {
			slog.Warn("unhandled method", "method", request.Method)
			continue
		}

		// start request with timeout
		ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
		s.active.Store(request.Id, cancel)

		// handle request asynchronously
		go func() {
			defer s.active.Delete(request.Id)

			err := s.handlers[request.Method](ctx, responseQueue, *request)
			if err != nil {
				switch err {
				case ErrInvalidParams:
					slog.Error("invalid params", "err", err)
					responseQueue <- protocol.NewResponseError(request.Id, protocol.ResponseError{
						Code:    protocol.ErrorCodeInvalidParams,
						Message: fmt.Sprintf("error decoding params for request %d", request.Id),
					})
				case ErrInternalError:
					slog.Error("handler error", "err", err)
					responseQueue <- protocol.NewResponseError(request.Id, protocol.ResponseError{
						Code:    protocol.ErrorCodeInternalError,
						Message: fmt.Sprintf("internal error processing request %d", request.Id),
					})
				default:
					slog.Error("unhandled error", "err", err)
					responseQueue <- protocol.NewResponseError(request.Id, protocol.ResponseError{
						Code:    protocol.ErrorCodeInternalError,
						Message: fmt.Sprintf("internal error processing request %d", request.Id),
					})
				}
				return
			}

			if ctx.Err() != nil {
				// Handler timed out or was cancelled
				slog.Error("handler timed out or was cancelled", "method", request.Method, "id", request.Id)
				responseQueue <- protocol.NewResponseError(request.Id, protocol.ResponseError{
					Code:    protocol.ErrorCodeInternalError,
					Message: "handler timed out or was cancelled",
				})
			}
		}()

	}
}
