package model

import (
	"github.com/nsaltun/user-service-grpc/pkg/v1/types"
	pbuser "github.com/nsaltun/user-service-grpc/proto/gen/go/core/user/v1"
	pbtypes "github.com/nsaltun/user-service-grpc/proto/gen/go/shared/types/v1"
	"go.mongodb.org/mongo-driver/bson"
)

type UserStatus int

const (
	UserStatus_Unspecified UserStatus = 0 //Default
	UserStatus_Active      UserStatus = 1 //Active
	UserStatus_Inactive    UserStatus = 2 //Inactive
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

type UserFilter struct {
	Status     UserStatus          `bson:"status" json:"status"`
	Email      string              `bson:"email" json:"email"`
	NickName   string              `bson:"nick_name" json:"nick_name"`
	FirstName  string              `bson:"first_name" json:"first_name"`
	LastName   string              `bson:"last_name" json:"last_name"`
	Country    string              `bson:"country" json:"country"`
	Pagination types.PaginationReq `json:"pagination"` //bson tag is not used for pagination
}

func (u *User) UserToProto() *pbuser.User {
	return &pbuser.User{
		Id:        u.Id,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		NickName:  u.NickName,
		//Please notice that password is not included in the proto
		Country: u.Country,
		Status:  pbuser.UserStatus(u.Status),
		Meta:    u.Meta.ToProto(),
	}
}

// ParseUserFilter converts a UserFilter into a MongoDB filter
func (f *UserFilter) ToBson() bson.M {
	//TODO: Sorting might be added as well
	mongoFilter := bson.M{}

	// Use exact matches for fields to utilize indexes
	if f.FirstName != "" {
		// Use prefix match instead of full regex if possible
		mongoFilter["first_name"] = bson.M{"$regex": "^" + f.FirstName, "$options": "i"} // Prefix matching
	}
	if f.LastName != "" {
		mongoFilter["last_name"] = bson.M{"$regex": "^" + f.LastName, "$options": "i"}
	}
	if f.NickName != "" {
		mongoFilter["nick_name"] = f.NickName // Exact match
	}
	if f.Email != "" {
		mongoFilter["email"] = f.Email // Exact match for email
	}
	if f.Country != "" {
		mongoFilter["country"] = f.Country // Exact match for country code
	}
	if f.Status == 0 {
		mongoFilter["status"] = UserStatus_Active // Set active as default
	} else if f.Status > 0 {
		mongoFilter["status"] = f.Status //Set value coming from userFilter
	}

	return mongoFilter
}

func (u *User) UserFromProto(pbUser *pbuser.User) {
	u.Id = pbUser.Id
	u.FirstName = pbUser.FirstName
	u.LastName = pbUser.LastName
	u.Email = pbUser.Email
	u.NickName = pbUser.NickName
	u.Country = pbUser.Country
	u.Password = pbUser.Password
	u.Status = UserStatus(pbUser.Status)
}

func (u *UserFilter) UserFilterFromProto(pbFilter *pbuser.UserFilter, pbPagination *pbtypes.List) {
	u.Status = UserStatus(pbFilter.Status)
	u.Email = pbFilter.Email
	u.NickName = pbFilter.NickName
	u.FirstName = pbFilter.FirstName
	u.LastName = pbFilter.LastName
	u.Country = pbFilter.Country
	u.Pagination = types.NewPaginationReq(pbPagination.GetOffset(), pbPagination.GetLimit())
}
