package model

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

// 预定义错误码
const (
	CodeSuccess       = 0
	CodeBadRequest    = 400
	CodeRateLimited   = 429
	CodeInternalError = 500
)

// Response 统一响应结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success 成功响应
func Success(ctx *app.RequestContext, data any) {
	ctx.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "ok",
		Data:    data,
	})
}

// Error 错误响应
func Error(ctx *app.RequestContext, code int, msg string) {
	httpStatus := http.StatusOK
	if code == CodeInternalError {
		httpStatus = http.StatusInternalServerError
	}
	ctx.JSON(httpStatus, Response{
		Code:    code,
		Message: msg,
	})
}

// Abort 中断并返回错误响应
func Abort(ctx *app.RequestContext, code int, msg string) {
	ctx.Abort()
	Error(ctx, code, msg)
}
