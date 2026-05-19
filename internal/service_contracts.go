package internal

import (
	"context"

	"kikneip.com/akirawhats/internal/model"
)

type UserService interface {
	GetAllUsers(ctx context.Context) ([]model.UserDTO, error)
	GetUserByID(ctx context.Context, id string) (*model.UserDTO, error)
	CreateUser(ctx context.Context, u model.UserDTOPost) (*model.UserDTO, error)
	UpdateUser(ctx context.Context, id string, u model.UserDTOPut) (*model.UserDTO, error)
	ChangePassword(ctx context.Context, id, currentPassword, newPassword string) error
	DeleteUser(ctx context.Context, id string) error
	AuthenticateUser(ctx context.Context, email, password string) (*model.UserDTO, error)
}
