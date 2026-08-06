package mcp

import "fmt"

const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
	codeAuthRequired   = -32001
	codeToolError      = -32002
)

type Error struct {
	RPCCode int
	Code    string
	Message string
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Code == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func rpcErr(rpcCode int, code string, message string) *Error {
	return &Error{RPCCode: rpcCode, Code: code, Message: message}
}

func authRequiredError() *Error {
	return rpcErr(codeAuthRequired, "auth_required", "PatchNote Agent login is required")
}
