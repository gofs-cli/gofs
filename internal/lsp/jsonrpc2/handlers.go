package jsonrpc2

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

func Initialize(s *Server) Handler {
	return func(ctx context.Context, _ chan protocol.Response, params any, id int) {
		p := params.(*protocol.InitializeRequest)

		// call the initializer function with the rootPath
		err := s.initializer(p.RootPath)
		if err != nil {
			log.Printf("fatal: initialize error: %s", err)
			err := s.conn.Write(protocol.NewResponseError(id, protocol.ResponseError{
				Code:    protocol.ErrorCodeInternalError,
				Message: "error calling initializer",
			}))
			if err != nil {
				log.Printf("error sending initialize error: %v", err)
			}
			return
		}

		// respond with the server's capabilities
		b, err := json.Marshal(protocol.InitializeResponse{
			Capabilities: s.capabilities,
		})
		if err != nil {
			log.Printf("json marshal error: %s", err)
			return
		}
		err = s.conn.Write(protocol.NewResponse(id, json.RawMessage(b)))
		if err != nil {
			log.Printf("error acknowledging shutdown: %v", err)
		}

		log.Println("completed the initialization")
	}
}

func Initialized(s *Server) Handler {
	return func(ctx context.Context, _ chan protocol.Response, params any, id int) {
		log.Println("server is initialized")
		s.isInitialized = true
	}
}

func Shutdown(s *Server) Handler {
	return func(ctx context.Context, _ chan protocol.Response, params any, id int) {
		log.Println("server is shutting down")
		s.isShutdown = true
		err := s.conn.Write(protocol.NewResponse(id, nil)) // acknowledge shutdown
		if err != nil {
			log.Printf("error acknowledging shutdown: %v", err)
		}
	}
}

func (s *Server) Cancel(request protocol.Request) {
	if request.Params == nil {
		return
	}

	// get the cancel message params
	var params protocol.CancelRequest
	err := json.NewDecoder(bytes.NewReader(*request.Params)).Decode(&params)
	if err != nil {
		log.Printf("cancel request decode error: %s", err)
		return
	}

	isRunning, cancelFunc := s.isRunning(params.Id)
	if isRunning {
		cancelFunc() // call the context cancel func
		s.endRequest(params.Id)
	}
}
