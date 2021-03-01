package httpx

import (
	"encoding/json"
	"github.com/tal-tech/go-zero/core/errorx"
	"github.com/tal-tech/go-zero/core/logx"
	"net/http"
	"sync"
)

var (
	errorHandler func(error) (int, interface{})
	lock         sync.RWMutex
)

// Error writes err into w.
const (
	ResponseWrapMessageKey = "message"
	ResponseWrapCodeKey    = "code"
	ResponseWrapDataKey    = "data"
	ResponseWrapSuccessKey = "success"
)

func Error(w http.ResponseWriter, err error) {
	//lock.RLock()
	//handler := errorHandler
	//lock.RUnlock()
	//
	//if handler == nil {
	//	WriteJson(w, http.StatusOK, map[string]interface{}{
	//		ResponseWrapCodeKey:    http.StatusBadRequest,
	//		ResponseWrapMessageKey: "error handler is null",
	//		ResponseWrapSuccessKey: false,
	//	})
	//	return
	//}
	//
	//code, body := errorHandler(err)
	e, ok := err.(errorx.Errorable)
	//e, ok := body.(error)
	if ok {
		WriteJson(w, http.StatusOK, map[string]interface{}{
			ResponseWrapMessageKey: e.GetMessage(),
			ResponseWrapCodeKey:    e.GetCode(),
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

// Ok writes HTTP 200 OK into w.
func Ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// OkJson writes v into w with 200 OK.
func OkJson(w http.ResponseWriter, v interface{}) {
	WriteJson(w, http.StatusOK, map[string]interface{}{
		ResponseWrapMessageKey: "success",
		ResponseWrapCodeKey:    10000,
		ResponseWrapDataKey:    v,
		ResponseWrapSuccessKey: true,
	})
}

// SetErrorHandler sets the error handler, which is called on calling Error.
func SetErrorHandler(handler func(error) (int, interface{})) {
	lock.Lock()
	defer lock.Unlock()
	errorHandler = handler
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
