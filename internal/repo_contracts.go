package internal

import "kikneip.com/akirawhats/internal/model"

type UserRepo interface {
	GetUserByID(id string) (*model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(id string) error
}
