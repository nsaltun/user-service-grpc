package model

import (
	"github.com/nsaltun/user-service-grpc/pkg/v1/types"
	pb "github.com/nsaltun/user-service-grpc/proto/gen/go/core/user/v1"
)

type UserStatus int

const (
	UserStatus_Active   UserStatus = 1 //Default
	UserStatus_Inactive UserStatus = 2
)

type User struct {
	Id         string           `bson:"_id" json:"id"`
	FirstName  string           `bson:"first_name" json:"first_name"`
	LastName   string           `bson:"last_name" json:"last_name"`
	Email      string           `bson:"email" json:"email"`
	NickName   string           `bson:"nick_name" json:"nick_name"`
	Password   string           `bson:"password" json:"password"`
	Country    string           `bson:"country" json:"country"`
	Status     UserStatus       `bson:"status" json:"status"`
	types.Meta `bson:",inline"` // Embed Meta fields directly
}

func (u *User) ToProto() *pb.User {
	return &pb.User{
		Id:        u.Id,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		NickName:  u.NickName,
		//Please notice that password is not included in the proto
		Country: u.Country,
		Status:  pb.UserStatus(u.Status),
		Meta:    u.Meta.ToProto(),
	}
}

func (u *User) FromProto(pbUser *pb.User, hashedPwd string) {
	u.Id = pbUser.Id
	u.FirstName = pbUser.FirstName
	u.LastName = pbUser.LastName
	u.Email = pbUser.Email
	u.NickName = pbUser.NickName
	u.Password = hashedPwd
	u.Country = pbUser.Country
	u.Status = UserStatus(pbUser.Status)
}
