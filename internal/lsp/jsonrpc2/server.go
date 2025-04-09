package jsonrpc2

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

type Handler func(context.Context, chan protocol.Response, protocol.Request)

type Server struct {
	conn              *Conn
	lifecycleHandlers map[string]Handler // lifecycle events
	handlers          map[string]Handler // notification and request events
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

func (s *Server) HandleRequest(method string, handler Handler) {
	s.handlers[method] = handler
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
	log.Println("server is listening")

	// responses from handlers are written asynchronously using this goroutine
	responseQueue := make(chan protocol.Response)
	go func() {
		for response := range responseQueue {
			if s.isShutdown {
				break
			}
			err := s.conn.Write(response)
			if err != nil {
				log.Println("error writing response: ", err)
			}
		}
	}()

	// main loop
	for {
		request, err := s.conn.Read()
		if err != nil {
			log.Println("error reading request: ", err)
			continue
		}

		if request.Method == "exit" {
			log.Println("server exit")
			// graceful exit
			return nil
		}

		// handle lifecycle events first
		if s.lifecycleHandlers[request.Method] != nil {
			s.lifecycleHandlers[request.Method](context.Background(), responseQueue, *request)
			continue
		}

		// do not process requests if lifecycle is before initialization or after shutdown
		if !s.isInitialized || s.isShutdown {
			// requests should error with InvalidRequest
			err := s.conn.Write(protocol.NewResponseError(request.Id, protocol.ResponseError{Code: protocol.ErrorCodeInvalidRequest, Message: "received request after shutdown"}))
			if err != nil {
				log.Println("error writing response: ", err)
			}
			continue
		}

		// handle cancel request
		if request.Method == "$/cancelRequest" {
			s.Cancel(*request)
			continue
		}

		if s.handlers[request.Method] != nil {
			ctx := s.startRequestWithContext(request.Id)
			c := make(chan bool)
			go func() {
				c <- true
				s.handlers[request.Method](ctx, responseQueue, *request)
				s.endRequest(request.Id)
			}()
			<-c
			continue
		}

		log.Println("unhandled method: ", request.Method)
	}
}
