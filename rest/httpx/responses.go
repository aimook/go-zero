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
	content[ResponseWrapSuccessKey] = false

	//error传入情况下，尝试从error中获取相关值
	e, ok := r.error.(errorx.BasicError)
	if ok {
		//未输入状态码
		if 0 == r.code && e.Code != 0 {
			r.code = e.Code
		}
		//data数据为空
		if nil == r.data && e.Data != nil {
			r.data = e.Data
		}
		//消息为空
		if "" == r.message && e.Message != "" {
			r.message = e.Message
		}
	}

	//再次检查，使用默认消息兜底
	//状态码
	if 0 == r.code {
		r.code = 5000
	}
	content[ResponseWrapCodeKey] = r.code
	//消息
	if "" == r.message {
		r.message = "fail"
	}
	content[ResponseWrapMessageKey] = r.message
	//data数据
	if nil == r.data && e.Data != nil {
		content[ResponseWrapDataKey] = e.Data
	}
	//开发模式
	if 1 == Model {
		if nil != r.error {
			if ok { //传入error，并且是BasicError类型
				content[ResponseWrapErrorKey] = e.Err
			} else { //其他类型error取Error()
				content[ResponseWrapErrorKey] = r.error.Error()
			}
		}
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
func Ok(w http.ResponseWriter) { w.WriteHeader(http.StatusOK) }

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
