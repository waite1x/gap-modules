package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waite1x/gap/log"
)

type HttpResult struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

const ErrCodeUnkown string = "UnkonwError"

type HttpError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func NewHttpError(err error) HttpError {
	return HttpError{
		Message: err.Error(),
		Code:    ErrCodeUnkown,
	}
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("http error: code: %s, msg: %s", e.Code, e.Message)
}

type ErrorHandFunc func(*gin.Context)

const ErrHandlersKey = "ErrorHandlers"

type ErrorHandlers struct {
	handlers []ErrorHandFunc
}

func NewErrorHandlers() *ErrorHandlers {
	return &ErrorHandlers{
		handlers: make([]ErrorHandFunc, 0),
	}
}

func (eh *ErrorHandlers) Run(c *gin.Context) {
	for i := range eh.handlers {
		eh.handlers[i](c)
	}
}
func (eh *ErrorHandlers) Add(h ErrorHandFunc) {
	eh.handlers = append(eh.handlers, h)
}

// ErrorMiddleware request panic error handler
func ErrorMiddleware(c *gin.Context) {
	handlers := &ErrorHandlers{}
	c.Set(ErrHandlersKey, handlers)
	defer func() {
		if r := recover(); r != nil {
			handlers.Run(c)
			if err, ok := r.(error); ok {
				log.Error("request error", err)
				c.JSON(http.StatusInternalServerError, NewHttpError(err))
			} else {
				log.Error("request error", err)
				c.JSON(http.StatusInternalServerError, r)
			}
			c.Abort()
		}
	}()
	c.Next()
}

// 添加请求生命周期中出现错误时的处理方法。
// 注意: 该方法中不要使用panic
func OnError(ctx context.Context, h ErrorHandFunc) {
	v := ctx.Value(ErrHandlersKey)
	if v != nil {
		v.(*ErrorHandlers).Add(h)
	}
}
