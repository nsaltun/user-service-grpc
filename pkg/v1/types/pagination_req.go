package types

import (
	"math"

	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	pb "github.com/nsaltun/user-service-grpc/proto/gen/go/shared/types/v1"
	"google.golang.org/grpc/codes"
)

const (
	DefaultOffset = 0
	DefaultLimit  = 20
	MinLimit      = 1
	MaxLimit      = 100
	MinOffset     = 0
	MaxOffset     = math.MaxInt64
)

type PaginationReq struct {
	Offset int64 `bson:"offset" json:"offset"`
	Limit  int64 `bson:"limit" json:"limit"`
}

func NewPaginationReq(offset int64, limit int64) PaginationReq {
	if offset == 0 {
		offset = DefaultOffset
	}
	if limit == 0 {
		limit = DefaultLimit
	}
	return PaginationReq{
		Offset: offset,
		Limit:  limit,
	}
}

func (p PaginationReq) FromProto(proto *pb.List) PaginationReq {
	return PaginationReq{
		Offset: proto.Offset,
		Limit:  proto.Limit,
	}
}

// validatePaginationParams validates
func ValidatePaginationParams(limit, offset int64) error {
	if limit < MinLimit || limit > MaxLimit {
		return errwrap.NewError("limit must be between 1 and 100", codes.InvalidArgument.String()).SetGrpcCode(codes.InvalidArgument)
	}

	if offset < MinOffset || offset > MaxOffset {
		return errwrap.NewError("offset must be non-negative", codes.InvalidArgument.String()).SetGrpcCode(codes.InvalidArgument)
	}

	return nil
}
