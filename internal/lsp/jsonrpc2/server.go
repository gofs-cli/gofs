package jsonrpc2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

type (
	Handler func(context.Context, chan protocol.Response, protocol.Request, any)
	Decoder func(protocol.Request) (any, error)
)

func DecodeParams[T any]() Decoder {
	return func(req protocol.Request) (any, error) {
		var params T
		err := json.NewDecoder(bytes.NewReader(*req.Params)).Decode(&params)
		if err != nil {
			return nil, fmt.Errorf("text document request decode error: %s", err)
		}

		return params, nil
	}
}

type Server struct {
	conn              *Conn
	lifecycleHandlers map[string]Handler // lifecycle events
	handlers          map[string]Handler // notification and request events
	decoders          map[string]Decoder // decode request params
	isInitialized     bool
	isShutdown        bool
	initializer       func(string) error // function to call when rootPath is known
	capabilities      protocol.ServerCapabilities
	active            sync.Map
}

// return a server with the connection set and a function that takes the rootPath tp initialize the lsp
func NewServer(conn *Conn, initializer func(string) error, capabilities protocol.ServerCapabilities) (*Server, error) {
	if conn == nil {
		return nil, fmt.Errorf("no connection")
	}

	return &Server{
		conn:              conn,
		lifecycleHandlers: make(map[string]Handler),
		handlers:          make(map[string]Handler),
		decoders:          make(map[string]Decoder),
		isInitialized:     false,
		isShutdown:        false,
		initializer:       initializer,
		capabilities:      capabilities,
		active:            sync.Map{},
	}, nil
}

func (s *Server) HandleLifecycle(method string, handler Handler) {
	s.lifecycleHandlers[method] = handler
}

func (s *Server) HandleRequest(method string, handler Handler, decoder Decoder) {
	s.handlers[method] = handler
	s.decoders[method] = decoder
}

func (s *Server) startRequestWithContext(id int) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	s.active.Store(id, cancel)
	return ctx
}

func (s *Server) endRequest(id int) {
	s.active.Delete(id)
}

func (s *Server) isRunning(id int) (bool, context.CancelFunc) {
	if cancelFunc, ok := s.active.Load(id); ok {
		return true, cancelFunc.(context.CancelFunc)
	}
	return false, nil
}

func (s *Server) ListenAndServe() error {
	slog.Info("server is listening")

	// responses from handlers are written asynchronously using this goroutine
	responseQueue := make(chan protocol.Response)
	go func() {
		for response := range responseQueue {
			if s.isShutdown {
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
		request, err := s.conn.Read()
		if err != nil {
			slog.Error("error reading request", "err", err)
			continue
		}

		if request.Method == "exit" {
			slog.Info("server exit")
			// graceful exit
			return nil
		}

		// handle lifecycle events first
		if s.lifecycleHandlers[request.Method] != nil {
			s.lifecycleHandlers[request.Method](context.Background(), responseQueue, *request, nil)
			continue
		}

		// do not process requests if lifecycle is before initialization or after shutdown
		if !s.isInitialized || s.isShutdown {
			// requests should error with InvalidRequest
			err := s.conn.Write(protocol.NewResponseError(request.Id, protocol.ResponseError{Code: protocol.ErrorCodeInvalidRequest, Message: "received request after shutdown"}))
			if err != nil {
				slog.Error("error writing response", "err", err)
			}
			continue
		}

		// handle cancel request
		if request.Method == "$/cancelRequest" {
			s.Cancel(*request)
			continue
		}

		if s.handlers[request.Method] != nil {
			params, err := s.decoders[request.Method](*request)
			if err != nil {
				slog.Error("error decoding request", "err", err)
				_ = s.conn.Write(protocol.NewResponseError(request.Id, protocol.ResponseError{Code: protocol.ErrorCodeInvalidParams, Message: "error decoding request"}))
				continue
			}
			ctx := s.startRequestWithContext(request.Id)
			c := make(chan bool)
			go func() {
				c <- true
				s.handlers[request.Method](ctx, responseQueue, *request, params)
				s.endRequest(request.Id)
			}()
			<-c
			continue
		}

		slog.Warn("unhandled method", "method", request.Method)
	}
}
