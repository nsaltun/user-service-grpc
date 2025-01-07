package errwrap

import (
	"fmt"

	"google.golang.org/grpc/codes"
)

type IError interface {
	error
	SetMessage(msg string) IError
	SetHttpCode(code int) IError
	SetGrpcCode(code codes.Code) IError
	SetOriginError(err error) IError
	HttpCode() int
	GrpcCode() codes.Code
	Message() string
	ErrorResp() ErrorResponse
	OriginErr() error
}

type errorWrapper struct {
	message   string
	code      string
	httpCode  int
	grpcCode  codes.Code
	originErr error
}

type ErrorResponse struct {
	Message string
	Code    string
}

func NewError(msg string, code string) IError {
	return &errorWrapper{
		message: msg,
		code:    code,
	}
}

func (e *errorWrapper) SetMessage(msg string) IError {
	newErr := e.clone()
	newErr.message = msg
	return newErr
}

func (e *errorWrapper) SetHttpCode(code int) IError {
	newErr := e.clone()
	newErr.httpCode = code
	return newErr
}

func (e *errorWrapper) SetGrpcCode(code codes.Code) IError {
	newErr := e.clone()
	newErr.grpcCode = code
	return newErr
}

func (e *errorWrapper) SetOriginError(err error) IError {
	newErr := e.clone()
	newErr.originErr = err
	return newErr
}

func (e *errorWrapper) HttpCode() int {
	return e.httpCode
}

func (e *errorWrapper) GrpcCode() codes.Code {
	return e.grpcCode
}

func (e *errorWrapper) Message() string {
	return e.message
}

func (e *errorWrapper) ErrorResp() ErrorResponse {
	return ErrorResponse{
		Code:    e.code,
		Message: e.message,
	}
}

func (e *errorWrapper) Error() string {
	return fmt.Sprintf("%s code:%s. originErr:%v", e.message, e.code, e.originErr)
}

func (e *errorWrapper) OriginErr() error {
	return e.originErr
}

func (e *errorWrapper) clone() *errorWrapper {
	if e == nil {
		return nil
	}
	return &errorWrapper{
		code:      e.code,
		httpCode:  e.httpCode,
		grpcCode:  e.grpcCode,
		message:   e.message,
		originErr: e.originErr,
	}
}
