package httpx

import (
	"encoding/json"
	"github.com/tal-tech/go-zero/core/errorx"
	"github.com/tal-tech/go-zero/core/logx"
	"net/http"
	"sync"
)

// Error writes err into w.
const (
	ResponseWrapMessageKey = "message"
	ResponseWrapCodeKey    = "code"
	ResponseWrapDataKey    = "data"
	ResponseWrapSuccessKey = "success"
	ResponseWrapErrorKey   = "error"

	Model = 1
)

var (
	errorHandler func(error) (int, interface{})
	lock         sync.RWMutex
	model        int
)

type reply struct {
	writer  http.ResponseWriter
	message string
	code    int
	data    interface{}
	error   error
}

func Create(w http.ResponseWriter) *reply {
	return &reply{writer: w}
}

func (r *reply) Data(data interface{}) *reply {
	r.data = data
	return r
}

func (r *reply) Message(message string) *reply {
	r.message = message
	return r
}

func (r *reply) Code(code int) *reply {
	r.code = code
	return r
}

func (r *reply) Success() {
	content := make(map[string]interface{}, 0)
	if "" == r.message {
		r.message = "success"
	}
	content[ResponseWrapMessageKey] = r.message
	content[ResponseWrapCodeKey] = 10000
	if nil != r.data {
		content[ResponseWrapDataKey] = r.data
	}
	content[ResponseWrapSuccessKey] = true
	WriteJson(r.writer, http.StatusOK, content)
	return
}

func (r *reply) Error(err error) {
	r.error = err
	r.Fail()
	return
}

func (r *reply) Fail() {
	content := make(map[string]interface{}, 0)
	//如果未定义消息，则尝试从error中获取
	if "" == r.message {
		e, ok := r.error.(errorx.BasicError)
		if ok {
			//如果未输入状态码，则尝试从error中获取
			if 0 == r.code {
				r.code = e.Code
			}
			r.message = e.Message
		} else {
			r.message = r.error.Error()
		}
	}
	//默认消息
	if "" == r.message {
		r.message = "fail"
	}
	content[ResponseWrapMessageKey] = r.message
	//默认状态码
	if 0 == r.code {
		r.code = 50000
	}
	content[ResponseWrapCodeKey] = r.code
	if nil != r.data {
		content[ResponseWrapDataKey] = r.data
	}
	content[ResponseWrapSuccessKey] = false
	if 1 == Model {
		content[ResponseWrapErrorKey] = r.error
	}
	WriteJson(r.writer, http.StatusOK, content)
	return
}

// Deprecated: use Create() instead
func Error(w http.ResponseWriter, err error) {
	e, ok := err.(errorx.BasicError)
	if ok {
		WriteJson(w, http.StatusOK, map[string]interface{}{
			ResponseWrapMessageKey: e.GetMessage(),
			ResponseWrapCodeKey:    e.GetCode(),
			ResponseWrapDataKey:    e.GetData(),
			ResponseWrapSuccessKey: false,
		})
	} else {
		WriteJson(w, http.StatusOK, map[string]interface{}{
			ResponseWrapMessageKey: err.Error(),
			ResponseWrapCodeKey:    http.StatusInternalServerError,
			ResponseWrapSuccessKey: false,
		})
	}
}

//  Deprecated:  use Create() instead
// Ok writes HTTP 200 OK into w.
func Ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

//  Deprecated:  use Create() instead
// OkJson writes v into w with 200 OK.
func OkJson(w http.ResponseWriter, v interface{}) {
	WriteJson(w, http.StatusOK, map[string]interface{}{
		ResponseWrapMessageKey: "success",
		ResponseWrapCodeKey:    10000,
		ResponseWrapDataKey:    v,
		ResponseWrapSuccessKey: true,
	})
}

//  Deprecated:  use Create() instead
// SetErrorHandler sets the error handler, which is called on calling Error.
func SetErrorHandler(handler func(error) (int, interface{})) {
	lock.Lock()
	defer lock.Unlock()
	errorHandler = handler
}

func SetModel(model int) {
	lock.Lock()
	defer lock.Unlock()
	model = model
}

// WriteJson writes v as json string into w with code.
func WriteJson(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set(ContentType, ApplicationJson)
	w.WriteHeader(code)

	if bs, err := json.Marshal(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if n, err := w.Write(bs); err != nil {
		// http.ErrHandlerTimeout has been handled by http.TimeoutHandler,
		// so it's ignored here.
		if err != http.ErrHandlerTimeout {
			logx.Errorf("write response failed, error: %s", err)
		}
	} else if n < len(bs) {
		logx.Errorf("actual bytes: %d, written bytes: %d", len(bs), n)
	}
}
