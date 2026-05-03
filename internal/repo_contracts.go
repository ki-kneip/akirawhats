package internal

import (
	"context"
	"kikneip.com/akirawhats/internal/model"
)

type UserRepo interface {
	GetAllUsers(ctx context.Context) ([]*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
}
