package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"kikneip.com/akirawhats/internal"
	"kikneip.com/akirawhats/internal/model"
)

type UserService struct {
	repo internal.UserRepo
}

func NewUserService(repo internal.UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, req model.UserDTOPost) (*model.UserDTO, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user := &model.User{
		ID:        newID(),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  string(hash),
	}
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return toDTO(user), nil
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*model.UserDTO, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDTO(user), nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]model.UserDTO, error) {
	users, err := s.repo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]model.UserDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, *toDTO(u))
	}
	return dtos, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, req model.UserDTOPut) (*model.UserDTO, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return toDTO(user), nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(ctx, id)
}

func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*model.UserDTO, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			return nil, internal.ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, internal.ErrInvalidCredentials
	}
	return toDTO(user), nil
}

func toDTO(u *model.User) *model.UserDTO {
	return &model.UserDTO{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
	}
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand unavailable: %v", err))
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
