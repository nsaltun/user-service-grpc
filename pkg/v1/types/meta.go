package types

import (
	"time"

	pb "github.com/nsaltun/user-service-grpc/proto/gen/go/shared/types/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Meta struct {
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
	Version   int32     `bson:"version" json:"version"`
}

func (m Meta) ToProto() *pb.Meta {
	return &pb.Meta{
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
		Version:   m.Version,
	}
}

func NewMeta() Meta {
	now := time.Now().UTC()
	return Meta{
		CreatedAt: now,
		UpdatedAt: now,
		Version:   0,
	}
}

func (m *Meta) Update() {
	m.UpdatedAt = time.Now().UTC()
}
