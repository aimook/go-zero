package errorx

import (
	"encoding/json"
)

type (
	Data             map[string]interface{}
	BasicErrorOption func(*BasicError)
	BasicError       struct {
		Code    int
		Message string
		Err     error
		Data    Data
	}
)

func (b BasicError) Error() string {
	bs, err := json.Marshal(b)
	if err != nil {
		return ""
	}
	return string(bs)
}

func WithCode(code int) BasicErrorOption {
	return func(err *BasicError) {
		err.Code = code
	}
}

func WithError(err error) BasicErrorOption {
	return func(err *BasicError) {
		err.Err = err
	}
}

func WithData(data Data) BasicErrorOption {
	return func(err *BasicError) {
		err.Data = data
	}
}

func WithDataItem(key string, value interface{}) BasicErrorOption {
	return func(err *BasicError) {
		if err.Data == nil {
			err.Data = Data(map[string]interface{}{
				key: value,
			})
		} else {
			err.Data[key] = value
		}
	}
}

func NewBasicError(message string, options ...BasicErrorOption) BasicError {
	basic := BasicError{
		Message: message,
	}
	for _, o := range options {
		o(&basic)
	}
	return basic
}

func (b *BasicError) GetCode() int {
	return b.Code
}

func (b *BasicError) GetMessage() string {
	return b.Message
}

func (b *BasicError) GetData() map[string]interface{} {
	return b.Data
}
