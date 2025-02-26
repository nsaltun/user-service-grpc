// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: core/user/v1/auth_api.proto

package userv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/nsaltun/user-service-grpc/proto/gen/go/core/user/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// AuthAPIName is the fully-qualified name of the AuthAPI service.
	AuthAPIName = "core.user.v1.AuthAPI"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// AuthAPILoginProcedure is the fully-qualified name of the AuthAPI's Login RPC.
	AuthAPILoginProcedure = "/core.user.v1.AuthAPI/Login"
	// AuthAPIRefreshProcedure is the fully-qualified name of the AuthAPI's Refresh RPC.
	AuthAPIRefreshProcedure = "/core.user.v1.AuthAPI/Refresh"
	// AuthAPILogoutProcedure is the fully-qualified name of the AuthAPI's Logout RPC.
	AuthAPILogoutProcedure = "/core.user.v1.AuthAPI/Logout"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	authAPIServiceDescriptor       = v1.File_core_user_v1_auth_api_proto.Services().ByName("AuthAPI")
	authAPILoginMethodDescriptor   = authAPIServiceDescriptor.Methods().ByName("Login")
	authAPIRefreshMethodDescriptor = authAPIServiceDescriptor.Methods().ByName("Refresh")
	authAPILogoutMethodDescriptor  = authAPIServiceDescriptor.Methods().ByName("Logout")
)

// AuthAPIClient is a client for the core.user.v1.AuthAPI service.
type AuthAPIClient interface {
	// Login authenticates a user with email and password
	Login(context.Context, *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error)
	// Refresh generates new access token using refresh token
	Refresh(context.Context, *connect.Request[v1.RefreshRequest]) (*connect.Response[v1.RefreshResponse], error)
	// Logout invalidates the current session
	Logout(context.Context, *connect.Request[v1.LogoutRequest]) (*connect.Response[v1.LogoutResponse], error)
}

// NewAuthAPIClient constructs a client for the core.user.v1.AuthAPI service. By default, it uses
// the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewAuthAPIClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) AuthAPIClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &authAPIClient{
		login: connect.NewClient[v1.LoginRequest, v1.LoginResponse](
			httpClient,
			baseURL+AuthAPILoginProcedure,
			connect.WithSchema(authAPILoginMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		refresh: connect.NewClient[v1.RefreshRequest, v1.RefreshResponse](
			httpClient,
			baseURL+AuthAPIRefreshProcedure,
			connect.WithSchema(authAPIRefreshMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		logout: connect.NewClient[v1.LogoutRequest, v1.LogoutResponse](
			httpClient,
			baseURL+AuthAPILogoutProcedure,
			connect.WithSchema(authAPILogoutMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// authAPIClient implements AuthAPIClient.
type authAPIClient struct {
	login   *connect.Client[v1.LoginRequest, v1.LoginResponse]
	refresh *connect.Client[v1.RefreshRequest, v1.RefreshResponse]
	logout  *connect.Client[v1.LogoutRequest, v1.LogoutResponse]
}

// Login calls core.user.v1.AuthAPI.Login.
func (c *authAPIClient) Login(ctx context.Context, req *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {
	return c.login.CallUnary(ctx, req)
}

// Refresh calls core.user.v1.AuthAPI.Refresh.
func (c *authAPIClient) Refresh(ctx context.Context, req *connect.Request[v1.RefreshRequest]) (*connect.Response[v1.RefreshResponse], error) {
	return c.refresh.CallUnary(ctx, req)
}

// Logout calls core.user.v1.AuthAPI.Logout.
func (c *authAPIClient) Logout(ctx context.Context, req *connect.Request[v1.LogoutRequest]) (*connect.Response[v1.LogoutResponse], error) {
	return c.logout.CallUnary(ctx, req)
}

// AuthAPIHandler is an implementation of the core.user.v1.AuthAPI service.
type AuthAPIHandler interface {
	// Login authenticates a user with email and password
	Login(context.Context, *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error)
	// Refresh generates new access token using refresh token
	Refresh(context.Context, *connect.Request[v1.RefreshRequest]) (*connect.Response[v1.RefreshResponse], error)
	// Logout invalidates the current session
	Logout(context.Context, *connect.Request[v1.LogoutRequest]) (*connect.Response[v1.LogoutResponse], error)
}

// NewAuthAPIHandler builds an HTTP handler from the service implementation. It returns the path on
// which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewAuthAPIHandler(svc AuthAPIHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	authAPILoginHandler := connect.NewUnaryHandler(
		AuthAPILoginProcedure,
		svc.Login,
		connect.WithSchema(authAPILoginMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	authAPIRefreshHandler := connect.NewUnaryHandler(
		AuthAPIRefreshProcedure,
		svc.Refresh,
		connect.WithSchema(authAPIRefreshMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	authAPILogoutHandler := connect.NewUnaryHandler(
		AuthAPILogoutProcedure,
		svc.Logout,
		connect.WithSchema(authAPILogoutMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/core.user.v1.AuthAPI/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case AuthAPILoginProcedure:
			authAPILoginHandler.ServeHTTP(w, r)
		case AuthAPIRefreshProcedure:
			authAPIRefreshHandler.ServeHTTP(w, r)
		case AuthAPILogoutProcedure:
			authAPILogoutHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedAuthAPIHandler returns CodeUnimplemented from all methods.
type UnimplementedAuthAPIHandler struct{}

func (UnimplementedAuthAPIHandler) Login(context.Context, *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("core.user.v1.AuthAPI.Login is not implemented"))
}

func (UnimplementedAuthAPIHandler) Refresh(context.Context, *connect.Request[v1.RefreshRequest]) (*connect.Response[v1.RefreshResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("core.user.v1.AuthAPI.Refresh is not implemented"))
}

func (UnimplementedAuthAPIHandler) Logout(context.Context, *connect.Request[v1.LogoutRequest]) (*connect.Response[v1.LogoutResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("core.user.v1.AuthAPI.Logout is not implemented"))
}
