package jsonrpc2

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/textproto"
	"strconv"
	"sync"

	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

type Conn struct {
	In  *bufio.Reader
	Out *bufio.Writer
	mu  sync.Mutex
}

func NewConn(in io.Reader, out io.Writer) *Conn {
	return &Conn{In: bufio.NewReader(in), Out: bufio.NewWriter(out)}
}

func (c *Conn) Read() (*protocol.Request, error) {
	header, err := textproto.NewReader(c.In).ReadMIMEHeader()
	if err == io.EOF {
		log.Fatal("read io EOF. terminating...\n")
	}
	if err != nil {
		return nil, fmt.Errorf("message MIME header error: %s", err)
	}

	l, err := strconv.ParseInt(header.Get("Content-Length"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("message content-length error: %s", err)
	}

	b := make([]byte, l)
	i, err := io.ReadFull(c.In, b)
	if err == io.EOF {
		log.Fatal("io EOF. terminating...\n")
	}
	if err != nil {
		return nil, fmt.Errorf("message body read error: %s", err)
	}
	if int64(i) != l {
		return nil, errors.New("message body length does not match header content-length")
	}

	var req protocol.Request
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&req)
	if err != nil {
		return nil, fmt.Errorf("message body json decode error: %s", err)
	}

	return &req, nil
}

func (c *Conn) Write(r protocol.Response) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	b, err := protocol.BaseMessage(r)
	if err != nil {
		return err
	}

	_, err = c.Out.Write(b)
	if err == io.ErrClosedPipe {
		log.Fatal("write io EOF. terminating...\n")
	}
	if err != nil {
		return fmt.Errorf("message write header error: %s", err)
	}

	err = c.Out.Flush()
	if err == io.ErrClosedPipe {
		log.Fatal("write io EOF. terminating...\n")
	}
	if err != nil {
		return fmt.Errorf("write buffer flush error: %s", err)
	}

	return nil
}
