package protocol

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/

type Request struct {
	Version string           `json:"jsonrpc"`
	Id      int              `json:"id"` // vscode uses int, but the spec says it can be int or string
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params"`
}

type Response struct {
	Version string           `json:"jsonrpc"`
	Id      int              `json:"id"` // vscode uses int, but the spec says it can be int or string or null
	Result  *json.RawMessage `json:"result"`
	Error   *ResponseError   `json:"error"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CancelRequest struct {
	Id int `json:"id"` // vscode uses int, but the spec says it can be int or string
}

func NewResponse(id int, result json.RawMessage) Response {
	return Response{
		Version: "2.0",
		Id:      id,
		Result:  &result,
	}
}

func NewResponseError(id int, err ResponseError) Response {
	return Response{
		Version: "2.0",
		Id:      id,
		Error:   &err,
	}
}

func NewEmptyResponse[T any](id int, r T) Response {
	b, err := json.Marshal(r)
	if err != nil {
		return Response{}
	}
	return NewResponse(id, json.RawMessage(b))
}

// raw base message with MIME header
func BaseMessage(r any) ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("json marshal error: %s", err)
	}

	msg := []byte("Content-Length: " + strconv.Itoa(len(b)) + "\r\n\r\n")
	msg = append(msg, b...)
	return msg, nil
}
