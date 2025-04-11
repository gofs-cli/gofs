package jsonrpc2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

type Handler func(context.Context, chan protocol.Response, any, int)

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
			var params any
			switch request.Method {
			case "initialize":
				if request.Params == nil {
					log.Println("initialize request missing params")
					err := s.conn.Write(protocol.NewResponseError(request.Id, protocol.ResponseError{
						Code:    protocol.ErrorCodeInvalidParams,
						Message: "initialize request missing params",
					}))
					if err != nil {
						log.Printf("error sending initialize error: %v", err)
					}
					continue
				}

				var p protocol.InitializeRequest
				err := json.NewDecoder(bytes.NewReader(*request.Params)).Decode(&p)
				if err != nil {
					log.Printf("initialize request decode error: %s", err)
					err := s.conn.Write(protocol.NewResponseError(request.Id, protocol.ResponseError{
						Code:    protocol.ErrorCodeInvalidParams,
						Message: "initialize request decode error",
					}))
					if err != nil {
						log.Printf("error sending initialize error: %v", err)
					}
					continue
				}
				params = &p

			case "initialized":

			case "shutdown":
			}

			s.lifecycleHandlers[request.Method](context.Background(), responseQueue, params, request.Id)
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

			var params any

			switch request.Method {
			case "textDocument/didOpen":
				p, err := protocol.DecodeParams[protocol.DidOpenRequest](*request)
				if err != nil {
					responseQueue <- protocol.NewResponseError(request.Id, protocol.ResponseError{
						Code:    protocol.ErrorCodeInvalidParams,
						Message: "error converting request to DidOpenRequest",
					})
					continue
				}
				params = p

			case "textDocument/didChange":
				p, err := protocol.DecodeParams[protocol.DidChangeRequest](*request)
				if err != nil {
					responseQueue <- protocol.NewResponseError(request.Id, protocol.ResponseError{
						Code:    protocol.ErrorCodeInvalidParams,
						Message: "error converting request to DidChangeRequest",
					})
					continue
				}
				params = p

			case "textDocument/didClose":
				p, err := protocol.DecodeParams[protocol.DidCloseRequest](*request)
				if err != nil {
					responseQueue <- protocol.NewResponseError(request.Id, protocol.ResponseError{
						Code:    protocol.ErrorCodeInvalidParams,
						Message: "error converting request to DidCloseRequest",
					})
					continue
				}
				params = p

			case "textDocument/didSave":

			case "textDocument/hover":
				p, err := protocol.DecodeParams[protocol.HoverRequest](*request)
				if err != nil {
					responseQueue <- protocol.NewResponseError(request.Id, protocol.ResponseError{
						Code:    protocol.ErrorCodeInvalidParams,
						Message: "error converting request to HoverRequest",
					})
					continue
				}
				params = p

			case "textDocument/diagnostic":
				p, err := protocol.DecodeParams[protocol.DiagnosticRequest](*request)
				if err != nil {
					responseQueue <- protocol.NewResponseError(request.Id, protocol.ResponseError{
						Code:    protocol.ErrorCodeInvalidParams,
						Message: "error converting request to DiagnosticRequest",
					})
					continue
				}
				params = p
			}

			go func() {
				c <- true
				s.handlers[request.Method](ctx, responseQueue, params, request.Id)
				s.endRequest(request.Id)
			}()
			<-c
			continue
		}

		log.Println("unhandled method: ", request.Method)
	}
}
