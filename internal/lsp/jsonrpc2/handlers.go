package jsonrpc2

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

const msgInitializeFailed = "error sending initialize error"

func Initialize(s *Server) Handler {
	return func(ctx context.Context, _ chan protocol.Response, params any, id int) {
		p := params.(*protocol.InitializeRequest)

		// call the initializer function with the rootPath
		err := s.initializer(p.RootPath)
		if err != nil {
			slog.Error("fatal: initialize error", "err", err)
			err := s.conn.Write(protocol.NewResponseError(id, protocol.ResponseError{
				Code:    protocol.ErrorCodeInternalError,
				Message: "error calling initializer",
			}))
			if err != nil {
				slog.Error(msgInitializeFailed, "err", err)
			}
			return
		}

		// respond with the server's capabilities
		b, err := json.Marshal(protocol.InitializeResponse{
			Capabilities: s.capabilities,
		})
		if err != nil {
			slog.Error("json marshal error", "err", err)
			return
		}
		err = s.conn.Write(protocol.NewResponse(id, json.RawMessage(b)))
		if err != nil {
			slog.Error("error acknowledging shutdown", "err", err)
		}
	}
}

func Initialized(s *Server) Handler {
	return func(ctx context.Context, _ chan protocol.Response, params any, id int) {
		slog.Info("server is initialized")
		s.isInitialized = true
	}
}

func Shutdown(s *Server) Handler {
	return func(ctx context.Context, _ chan protocol.Response, params any, id int) {
		slog.Info("server is shutting down")
		s.isShutdown = true
		err := s.conn.Write(protocol.NewResponse(id, nil)) // acknowledge shutdown
		if err != nil {
			slog.Error("error acknowledging shutdown", "err", err)
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
		slog.Error("cancel request decode error", "err", err)
		return
	}

	isRunning, cancelFunc := s.isRunning(params.Id)
	if isRunning {
		cancelFunc() // call the context cancel func
		s.endRequest(params.Id)
	}
}
