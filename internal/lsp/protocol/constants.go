package protocol

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/

const (
	RequestInitialize           = "initialize"
	RequestShutdown             = "shutdown"
	MethodSetTrace              = "$/setTrace" // set trace notification: 'off' | 'messages' | 'verbose'
	MethodTextDocumentDidChange = "textDocument/didChange"
	MethodTextDocumentDidClose  = "textDocument/didClose"
	MethodTextDocumentDidSave   = "textDocument/didSave"
	MethodTextDocumentHover     = "textDocument/hover"
)

const (
	NotificationInitialized = "initialized"
	NotificationExit        = "exit"
)

const (
	ErrorCodeParseError     = -32700
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInternalError  = -32603
)
