package jsonrpc2

import (
	"io"
	"sync"
	"time"

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
	r.open = false
}

func (r *testReader) Read(p []byte) (int, error) {
	if !r.open {
		return 0, io.EOF
	}
	for r.cur >= len(r.reqs) {
		time.Sleep(100 * time.Millisecond)
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	msg, err := protocol.BaseMessage(r.reqs[r.cur])
	if err != nil {
		return 0, err
	}
	p = p[:len(msg)]
	copy(p, msg)
	r.cur++
	return len(p), nil
}

func testServer(c *Conn) *Server {
	s, _ := NewServer(c, func(string) error { return nil }, protocol.ServerCapabilities{})
	s.HandleLifecycle("initialize", Initialize(s))
	s.HandleLifecycle("initialized", Initialized(s))
	s.HandleLifecycle("shutdown", Shutdown(s))
	return s
}

// func TestListenAndServe(t *testing.T) {
// 	t.Parallel()

// 	t.Run("server starts, shuts downs, and calls initializer", func(t *testing.T) {
// 		initParams := json.RawMessage(`{"rootPath": "/foo/bar"}`)
// 		reader := newTestReader([]protocol.Request{
// 			{
// 				Version: "2.0",
// 				Id:      1,
// 				Method:  protocol.RequestInitialize,
// 				Params:  &initParams,
// 			},
// 			{
// 				Version: "2.0",
// 				Id:      2,
// 				Method:  protocol.NotificationInitialized,
// 				Params:  nil,
// 			},
// 			{
// 				Version: "2.0",
// 				Id:      3,
// 				Method:  protocol.RequestShutdown,
// 				Params:  nil,
// 			},
// 			{
// 				Version: "2.0",
// 				Id:      4,
// 				Method:  protocol.NotificationExit,
// 				Params:  nil,
// 			},
// 		})
// 		writer := new(bytes.Buffer)
// 		conn := NewConn(reader, writer)
// 		rp := ""
// 		s, err := NewServer(conn, func(rootPath string) error {
// 			rp = rootPath
// 			return nil
// 		}, protocol.ServerCapabilities{})
// 		s.HandleLifecycle("initialize", Initialize(s))
// 		s.HandleLifecycle("initialized", Initialized(s))
// 		s.HandleLifecycle("shutdown", Shutdown(s))
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		err = s.ListenAndServe()
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		if rp != "/foo/bar" {
// 			t.Errorf("expected root path /foo/bar, got %v", rp)
// 		}
// 	})

// 	t.Run("server calls handler", func(t *testing.T) {
// 		initParams := json.RawMessage(`{"rootPath": "/foo/bar"}`)
// 		reader := newTestReader([]protocol.Request{
// 			{
// 				Version: "2.0",
// 				Id:      1,
// 				Method:  protocol.RequestInitialize,
// 				Params:  &initParams,
// 			},
// 			{
// 				Version: "2.0",
// 				Id:      2,
// 				Method:  protocol.NotificationInitialized,
// 				Params:  nil,
// 			},
// 			{
// 				Version: "2.0",
// 				Id:      3,
// 				Method:  "foo",
// 				Params:  nil,
// 			},
// 		})
// 		writer := new(bytes.Buffer)
// 		conn := NewConn(reader, writer)
// 		s := testServer(conn)
// 		s.HandleRequest("foo", func(ctx context.Context, rq chan protocol.Response, r protocol.Request) {
// 			rq <- protocol.NewResponse(r.Id, json.RawMessage(`{"foo": "bar"}`))
// 		})
// 		go func() {
// 			// give the writer time to write the response
// 			time.Sleep(100 * time.Millisecond)
// 			reader.addReq(protocol.Request{
// 				Version: "2.0",
// 				Id:      4,
// 				Method:  protocol.RequestShutdown,
// 				Params:  nil,
// 			})
// 			reader.addReq(protocol.Request{
// 				Version: "2.0",
// 				Id:      5,
// 				Method:  protocol.NotificationExit,
// 				Params:  nil,
// 			})
// 		}()
// 		err := s.ListenAndServe()
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		// init response
// 		expected := "Content-Length: 206\r\n\r\n" + `{"jsonrpc":"2.0","id":1,"result":{"capabilities":{"textDocumentSync":0,"hoverProvider":false,"diagnosticProvider":{"identifier":"","interFileDependencies":false,"workspaceDiagnostics":false}}},"error":null}`
// 		// request response
// 		expected += "Content-Length: 60\r\n\r\n" + `{"jsonrpc":"2.0","id":3,"result":{"foo":"bar"},"error":null}`
// 		// shutdown response
// 		expected += "Content-Length: 51\r\n\r\n" + `{"jsonrpc":"2.0","id":4,"result":null,"error":null}`
// 		if writer.String() != expected {
// 			t.Errorf("expected:\n%v\ngot:\n%v", expected, writer.String())
// 		}
// 	})

// 	t.Run("server responds to requests in any order", func(t *testing.T) {
// 		initParams := json.RawMessage(`{"rootPath": "/foo/bar"}`)
// 		reader := newTestReader([]protocol.Request{
// 			{
// 				Version: "2.0",
// 				Id:      1,
// 				Method:  protocol.RequestInitialize,
// 				Params:  &initParams,
// 			},
// 			{
// 				Version: "2.0",
// 				Id:      2,
// 				Method:  protocol.NotificationInitialized,
// 				Params:  nil,
// 			},
// 			{
// 				Version: "2.0",
// 				Id:      3,
// 				Method:  "foo", // foo sent first
// 				Params:  nil,
// 			},
// 			{
// 				Version: "2.0",
// 				Id:      4,
// 				Method:  "bar", // bar sent second
// 				Params:  nil,
// 			},
// 		})
// 		writer := new(bytes.Buffer)
// 		conn := NewConn(reader, writer)
// 		s := testServer(conn)
// 		s.HandleRequest("foo", func(ctx context.Context, rq chan protocol.Response, r protocol.Request) {
// 			time.Sleep(100 * time.Millisecond)
// 			rq <- protocol.NewResponse(r.Id, json.RawMessage(`{"foo": "response"}`))
// 		})
// 		s.HandleRequest("bar", func(ctx context.Context, rq chan protocol.Response, r protocol.Request) {
// 			rq <- protocol.NewResponse(r.Id, json.RawMessage(`{"bar": "response"}`))
// 		})
// 		go func() {
// 			// give the handlers time to respond
// 			time.Sleep(200 * time.Millisecond)
// 			reader.addReq(protocol.Request{
// 				Version: "2.0",
// 				Id:      5,
// 				Method:  protocol.RequestShutdown,
// 				Params:  nil,
// 			})
// 			reader.addReq(protocol.Request{
// 				Version: "2.0",
// 				Id:      6,
// 				Method:  protocol.NotificationExit,
// 				Params:  nil,
// 			})
// 		}()
// 		err := s.ListenAndServe()
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		// init response
// 		expected := "Content-Length: 206\r\n\r\n" + `{"jsonrpc":"2.0","id":1,"result":{"capabilities":{"textDocumentSync":0,"hoverProvider":false,"diagnosticProvider":{"identifier":"","interFileDependencies":false,"workspaceDiagnostics":false}}},"error":null}`
// 		// bar first
// 		expected += "Content-Length: 65\r\n\r\n" + `{"jsonrpc":"2.0","id":4,"result":{"bar":"response"},"error":null}`
// 		// foo second
// 		expected += "Content-Length: 65\r\n\r\n" + `{"jsonrpc":"2.0","id":3,"result":{"foo":"response"},"error":null}`
// 		// shutdown response
// 		expected += "Content-Length: 51\r\n\r\n" + `{"jsonrpc":"2.0","id":5,"result":null,"error":null}`

// 		if writer.String() != expected {
// 			t.Errorf("expected:\n%v\ngot:\n%v", expected, writer.String())
// 		}
// 	})

// 	t.Run("server cancels request", func(t *testing.T) {
// 		initParams := json.RawMessage(`{"rootPath": "/foo/bar"}`)
// 		cancelParams := json.RawMessage(`{"id": 3}`)
// 		reader := newTestReader([]protocol.Request{
// 			{
// 				Version: "2.0",
// 				Id:      1,
// 				Method:  protocol.RequestInitialize,
// 				Params:  &initParams,
// 			},
// 			{
// 				Version: "2.0",
// 				Id:      2,
// 				Method:  protocol.NotificationInitialized,
// 				Params:  nil,
// 			},
// 			{
// 				Version: "2.0",
// 				Id:      3,
// 				Method:  "foo",
// 				Params:  nil,
// 			},
// 			{
// 				Version: "2.0",
// 				Id:      4,
// 				Method:  "$/cancelRequest",
// 				Params:  &cancelParams,
// 			},
// 		})
// 		writer := new(bytes.Buffer)
// 		conn := NewConn(reader, writer)
// 		s := testServer(conn)
// 		cancelled := false
// 		s.HandleRequest("foo", func(ctx context.Context, rq chan protocol.Response, r protocol.Request) {
// 			// give the cancel request time to be processed
// 			time.Sleep(100 * time.Millisecond)
// 			select {
// 			case <-ctx.Done():
// 				cancelled = true
// 			default:
// 				rq <- protocol.NewResponse(r.Id, json.RawMessage(`{"foo": "response"}`))
// 			}
// 		})
// 		go func() {
// 			// give the handlers time to respond
// 			time.Sleep(200 * time.Millisecond)
// 			reader.addReq(protocol.Request{
// 				Version: "2.0",
// 				Id:      5,
// 				Method:  protocol.RequestShutdown,
// 				Params:  nil,
// 			})
// 			reader.addReq(protocol.Request{
// 				Version: "2.0",
// 				Id:      6,
// 				Method:  protocol.NotificationExit,
// 				Params:  nil,
// 			})
// 		}()
// 		err := s.ListenAndServe()
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		if !cancelled {
// 			t.Errorf("expected request to be cancelled but want not")
// 		}
// 		isRunning, _ := s.isRunning(3)
// 		if isRunning {
// 			t.Errorf("expected request to be cancelled but was not")
// 		}
// 	})
// }
