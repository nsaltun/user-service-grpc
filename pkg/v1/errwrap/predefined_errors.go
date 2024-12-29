package errwrap

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

var (
	ErrBadRequest       = NewError("invalid argument", "400").SetHttpCode(http.StatusBadRequest).SetGrpcCode(codes.InvalidArgument)
	ErrUnauthenticated  = NewError("token is invalid", "401").SetHttpCode(http.StatusUnauthorized).SetGrpcCode(codes.Unauthenticated)
	ErrPermissionDenied = NewError("permission denied", "403").SetHttpCode(http.StatusForbidden).SetGrpcCode(codes.PermissionDenied)
	ErrNotFound         = NewError("resource not found", "404").SetHttpCode(http.StatusNotFound).SetGrpcCode(codes.NotFound)
	ErrConflict         = NewError("already exists", "409").SetHttpCode(http.StatusConflict).SetGrpcCode(codes.AlreadyExists)
	ErrInternal         = NewError("internal server error", "500").SetHttpCode(http.StatusInternalServerError).SetGrpcCode(codes.Internal)
)
