package internal

import (
	"context"

	"kikneip.com/akirawhats/internal/model"
)

type UserService interface {
	GetAllUsers(ctx context.Context) ([]model.UserDTO, error)
	GetUserByID(ctx context.Context, id string) (*model.UserDTO, error)
	CreateUser(ctx context.Context, u model.UserDTOPost) (*model.UserDTO, error)
	DeleteUser(ctx context.Context, id string) error
}
