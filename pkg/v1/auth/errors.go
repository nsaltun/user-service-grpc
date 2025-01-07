package auth

type JwtError struct {
	error
	msg       string
	originErr error
}

var (
	ErrTokenExpired                    = NewJwtError("token has expired")
	ErrTokenJwtParse                   = NewJwtError("error parsing jwt")
	ErrInvalidToken                    = NewJwtError("invalid token")
	ErrTokenTypeNotSpecified           = NewJwtError("token type not specified")
	ErrInvalidTokenTypeExpectedAccess  = NewJwtError("invalid token type: expected access token")
	ErrInvalidTokenTypeExpectedRefresh = NewJwtError("invalid token type: expected refresh token")
	ErrTokenInvalidated                = NewJwtError("token has been invalidated")
	ErrTokenStatusVerificationFailed   = NewJwtError("failed to verify token status")
	ErrRefreshTokenValidationFailed    = NewJwtError("failed to validate refresh token")
	ErrGenerateAccessTokenFailed       = NewJwtError("failed to generate access token")
	ErrGenerateRefreshTokenFailed      = NewJwtError("failed to generate refresh token")
	ErrInvalidateTokenFailed           = NewJwtError("failed to invalidate token")
	ErrInvalidateDeviceTokenFailed     = NewJwtError("failed to invalidate device token")
)

func NewJwtError(msg string) *JwtError {
	return &JwtError{
		msg: msg,
	}
}

func (e *JwtError) SetOriginErr(err error) *JwtError {
	e.clone().originErr = err
	return e
}

func (e *JwtError) clone() *JwtError {
	return &JwtError{
		msg:       e.msg,
		originErr: e.originErr,
	}
}

func (e *JwtError) Error() string {
	return e.msg
}
