package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/gofs-cli/gofs/internal/lsp/jsonrpc2"
	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

type testReader struct {
	reqs  []protocol.Request
	cur   int
	open  bool
	mutex sync.Mutex
}

func newTestReader(reqs []protocol.Request) *testReader {
	return &testReader{
		reqs: reqs,
		cur:  0,
		open: true,
	}
}

func (r *testReader) addReq(req protocol.Request) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.reqs = append(r.reqs, req)
}

func (r *testReader) Close() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.open = false
}

func (r *testReader) Read(p []byte) (int, error) {
	r.mutex.Lock()
	open := r.open
	r.mutex.Unlock()
	if !open {
		return 0, io.EOF
	}

	for {
		r.mutex.Lock()
		if r.cur < len(r.reqs) {
			break
		}
		r.mutex.Unlock()
		time.Sleep(10 * time.Millisecond) // reduce wait time
	}

	msg, err := protocol.BaseMessage(r.reqs[r.cur])
	if err != nil {
		r.mutex.Unlock()
		return 0, err
	}
	r.cur++
	r.mutex.Unlock()

	// copy outside of lock
	n := copy(p, msg)
	return n, nil
}

func testServer(c *jsonrpc2.Conn) *jsonrpc2.Server {
	s, _ := jsonrpc2.NewServer(c, func(string) error { return nil }, protocol.ServerCapabilities{})
	s.HandleLifecycle("initialize", jsonrpc2.Initialize(s))
	s.HandleLifecycle("initialized", jsonrpc2.Initialized(s))
	s.HandleLifecycle("shutdown", jsonrpc2.Shutdown(s))
	return s
}

func TestListenAndServe(t *testing.T) {
	t.Parallel()

	t.Run("handler blocks when it requires a file and didOpen is running", func(t *testing.T) {
		initParams := json.RawMessage(`{"rootPath": "/foo/bar"}`)
		reader := newTestReader([]protocol.Request{
			{
				Version: "2.0",
				Id:      1,
				Method:  protocol.RequestInitialize,
				Params:  &initParams,
			},
			{
				Version: "2.0",
				Id:      2,
				Method:  protocol.NotificationInitialized,
				Params:  nil,
			},
			{
				Version: "2.0",
				Id:      3,
				Method:  "didOpen",
				Params:  nil,
			},
			{
				Version: "2.0",
				Id:      4,
				Method:  "foo", // foo should wait for didOpen to finish
				Params:  nil,
			},
		})
		writer := new(bytes.Buffer)
		conn := jsonrpc2.NewConn(reader, writer)
		s := testServer(conn)
		r := NewRepo()
		s.HandleRequest("didOpen", func(ctx context.Context, rq chan protocol.Response, rs protocol.Request) {
			r.OpenTemplFile(DidOpenRequest{
				TextDocument: protocol.TextDocument{
					Path: "/foo/bar/templ.templ",
					Text: `package test

templ Test() {}
`,
				},
			})
			rq <- protocol.NewResponse(rs.Id, json.RawMessage(`{"didOpen": "response"}`))
		})
		s.HandleRequest("foo", func(ctx context.Context, rq chan protocol.Response, rs protocol.Request) {
			t := r.GetTemplFile("/foo/bar/templ.templ")
			if t == nil {
				rq <- protocol.NewResponse(rs.Id, json.RawMessage(`{"foo": "fail"}`))
			} else {
				rq <- protocol.NewResponse(rs.Id, json.RawMessage(`{"foo": "response"}`))
			}
		})
		go func() {
			// give the handlers time to respond
			time.Sleep(200 * time.Millisecond)
			reader.addReq(protocol.Request{
				Version: "2.0",
				Id:      5,
				Method:  protocol.RequestShutdown,
				Params:  nil,
			})
			reader.addReq(protocol.Request{
				Version: "2.0",
				Id:      6,
				Method:  protocol.NotificationExit,
				Params:  nil,
			})
		}()
		err := s.ListenAndServe()
		if err != nil {
			t.Error(err)
		}

		// init response
		expected := "Content-Length: 206\r\n\r\n" + `{"jsonrpc":"2.0","id":1,"result":{"capabilities":{"textDocumentSync":0,"hoverProvider":false,"diagnosticProvider":{"identifier":"","interFileDependencies":false,"workspaceDiagnostics":false}}},"error":null}`
		// didOpen first
		expected += "Content-Length: 69\r\n\r\n" + `{"jsonrpc":"2.0","id":3,"result":{"didOpen":"response"},"error":null}`
		// foo second, should wait for didOpen to finish
		expected += "Content-Length: 65\r\n\r\n" + `{"jsonrpc":"2.0","id":4,"result":{"foo":"response"},"error":null}`
		// shutdown response
		expected += "Content-Length: 51\r\n\r\n" + `{"jsonrpc":"2.0","id":5,"result":null,"error":null}`

		if writer.String() != expected {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, writer.String())
		}
	})

	t.Run("handler does not blocks when it does not requires a file", func(t *testing.T) {
		initParams := json.RawMessage(`{"rootPath": "/foo/bar"}`)
		reader := newTestReader([]protocol.Request{
			{
				Version: "2.0",
				Id:      1,
				Method:  protocol.RequestInitialize,
				Params:  &initParams,
			},
			{
				Version: "2.0",
				Id:      2,
				Method:  protocol.NotificationInitialized,
				Params:  nil,
			},
			{
				Version: "2.0",
				Id:      3,
				Method:  "didOpen",
				Params:  nil,
			},
			{
				Version: "2.0",
				Id:      4,
				Method:  "foo", // foo should wait for didOpen to finish
				Params:  nil,
			},
		})
		writer := new(bytes.Buffer)
		conn := jsonrpc2.NewConn(reader, writer)
		s := testServer(conn)
		r := NewRepo()
		s.HandleRequest("didOpen", func(ctx context.Context, rq chan protocol.Response, rs protocol.Request) {
			r.OpenTemplFile(DidOpenRequest{
				TextDocument: protocol.TextDocument{
					Path: "/foo/bar/templ.templ",
					Text: `package test

templ Test() {}
`,
				},
			})
			rq <- protocol.NewResponse(rs.Id, json.RawMessage(`{"didOpen": "response"}`))
		})
		s.HandleRequest("foo", func(ctx context.Context, rq chan protocol.Response, rs protocol.Request) {
			rq <- protocol.NewResponse(rs.Id, json.RawMessage(`{"foo": "response"}`))
		})
		go func() {
			// give the handlers time to respond
			time.Sleep(200 * time.Millisecond)
			reader.addReq(protocol.Request{
				Version: "2.0",
				Id:      5,
				Method:  protocol.RequestShutdown,
				Params:  nil,
			})
			reader.addReq(protocol.Request{
				Version: "2.0",
				Id:      6,
				Method:  protocol.NotificationExit,
				Params:  nil,
			})
		}()
		err := s.ListenAndServe()
		if err != nil {
			t.Error(err)
		}

		// init response
		expected := "Content-Length: 206\r\n\r\n" + `{"jsonrpc":"2.0","id":1,"result":{"capabilities":{"textDocumentSync":0,"hoverProvider":false,"diagnosticProvider":{"identifier":"","interFileDependencies":false,"workspaceDiagnostics":false}}},"error":null}`
		// foo should return first
		expected += "Content-Length: 65\r\n\r\n" + `{"jsonrpc":"2.0","id":4,"result":{"foo":"response"},"error":null}`
		// didOpen second
		expected += "Content-Length: 69\r\n\r\n" + `{"jsonrpc":"2.0","id":3,"result":{"didOpen":"response"},"error":null}`
		// shutdown response
		expected += "Content-Length: 51\r\n\r\n" + `{"jsonrpc":"2.0","id":5,"result":null,"error":null}`

		if writer.String() != expected {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, writer.String())
		}
	})
}
