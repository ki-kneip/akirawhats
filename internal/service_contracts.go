package internal

import "kikneip.com/akirawhats/internal/model"

type UserService interface {
	GetUserByID(id string) (*model.UserDTO, error)
	CreateUser(u model.UserDTOPost) error
	DeleteUser(id string) error
}
