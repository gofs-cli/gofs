package protocol

import (
	"encoding/json"
	"testing"
)

func TestBaseMessage(t *testing.T) {
	t.Parallel()

	t.Run("response result", func(t *testing.T) {
		resp := NewResponse(1, json.RawMessage(`{"foo": "bar"}`))
		msg, err := BaseMessage(resp)
		if err != nil {
			t.Error(err)
		}
		expected := "Content-Length: 60\r\n\r\n" + `{"jsonrpc":"2.0","id":1,"result":{"foo":"bar"},"error":null}`
		if string(msg) != expected {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, string(msg))
		}
	})

	t.Run("response error", func(t *testing.T) {
		resp := NewResponseError(1, ResponseError{Code: 1, Message: "error"})
		msg, err := BaseMessage(resp)
		if err != nil {
			t.Error(err)
		}
		expected := "Content-Length: 75\r\n\r\n" + `{"jsonrpc":"2.0","id":1,"result":null,"error":{"code":1,"message":"error"}}`
		if string(msg) != expected {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, string(msg))
		}
	})
}
