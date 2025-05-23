package jsonrpc2

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

const msgInitializeFailed = "error sending initialize error"

func (s *Server) Initialize(request protocol.Request) {
	if request.Params == nil {
		slog.Warn("initialize request missing params")
		err := s.conn.Write(protocol.NewResponseError(request.Id, protocol.ResponseError{
			Code:    protocol.ErrorCodeInvalidParams,
			Message: "initialize request missing params",
		}))
		if err != nil {
			slog.Error(msgInitializeFailed, "err", err)
		}
		return
	}

	// get the initialize message p, which include the rootPath
	var p protocol.InitializeRequest
	err := json.NewDecoder(bytes.NewReader(*request.Params)).Decode(&p)
	if err != nil {
		slog.Error("initialize request decode error", "err", err)
		err := s.conn.Write(protocol.NewResponseError(request.Id, protocol.ResponseError{
			Code:    protocol.ErrorCodeInvalidParams,
			Message: "initialize request decode error",
		}))
		if err != nil {
			slog.Error(msgInitializeFailed, "err", err)
		}
		return
	}

	// call the initializer function with the rootPath
	err = s.initializer(p.RootPath)
	if err != nil {
		slog.Error("fatal: initialize error", "err", err)
		err := s.conn.Write(protocol.NewResponseError(request.Id, protocol.ResponseError{
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
	err = s.conn.Write(protocol.NewResponse(request.Id, json.RawMessage(b)))
	if err != nil {
		slog.Error("error acknowledging shutdown", "err", err)
	}
}

func (s *Server) Shutdown(request protocol.Request) {
	// acknowledge shutdown message immediately
	err := s.conn.Write(protocol.NewResponse(request.Id, nil))
	if err != nil {
		slog.Error("error acknowledging shutdown", "err", err)
	}
	s.isShutdown = true // do not process requests or send responses after this
	s.active.Range(func(key, value interface{}) bool {
		cancelFunc := value.(context.CancelFunc)
		cancelFunc()
		s.active.Delete(key)
		return true
	})
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

	if cancelFunc, ok := s.active.Load(params.Id); ok {
		cancelFunc.(context.CancelFunc)() // call the context cancel func
		s.active.Delete(params.Id)
	}
}
