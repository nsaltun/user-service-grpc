package repository

type Repository interface {
	UserRepo
}

type repository struct {
	UserRepo
}

func New(userRepo UserRepo) Repository {
	return &repository{
		userRepo,
	}
}
