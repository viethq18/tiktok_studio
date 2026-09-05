package httpx

import (
	"errors"
	"fmt"
	"net/http"
)

// Error is the only shape that ever reaches the client. Internal detail stays
// in Internal and is logged, never serialized (§90).
type Error struct {
	Status   int    `json:"-"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Internal error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Code, e.Internal)
	}
	return e.Code
}

func (e *Error) Unwrap() error { return e.Internal }

func NewError(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

func Wrap(status int, code, message string, err error) *Error {
	return &Error{Status: status, Code: code, Message: message, Internal: err}
}

var (
	ErrUnauthorized = NewError(http.StatusUnauthorized, "unauthorized", "Bạn cần đăng nhập.")
	ErrForbidden    = NewError(http.StatusForbidden, "forbidden", "Bạn không có quyền truy cập tài nguyên này.")
	ErrNotFound     = NewError(http.StatusNotFound, "not_found", "Không tìm thấy tài nguyên.")
	ErrConflict     = NewError(http.StatusConflict, "conflict", "Dữ liệu đã thay đổi ở nơi khác. Hãy tải lại.")
)

func BadRequest(message string) *Error {
	return NewError(http.StatusBadRequest, "bad_request", message)
}

func Internal(err error) *Error {
	return Wrap(http.StatusInternalServerError, "internal_error", "Đã có lỗi xảy ra. Vui lòng thử lại.", err)
}

func AsError(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return Internal(err)
}
