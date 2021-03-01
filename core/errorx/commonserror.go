package errorx

import "fmt"

type Errorable interface {
	GetCode() int
	GetMessage() string
}

type rpcError struct {
	Code    int
	Message string
}

func (r *rpcError) GetCode() int {
	return r.Code
}

func (r *rpcError) GetMessage() string {
	return r.Message
}

func (r *rpcError) Error() string {
	return fmt.Sprintf("rpc调用异常，code=%d,message=%s", r.Code, r.Message)
}

func NewRpcErrorWithMessage(message string) error {
	return &rpcError{
		Code:    10091,
		Message: message,
	}
}

func NewRpcErrorWithError(error error) error {
	return &rpcError{
		Code:    10091,
		Message: error.Error(),
	}
}

type BasicError struct {
	Code    int
	Message string
}

func (b *BasicError) GetCode() int {
	return b.Code
}

func (b *BasicError) GetMessage() string {
	return b.Message
}

func NewBasicError(code int, message string) error {
	return &rpcError{
		Code:    code,
		Message: message,
	}
}
