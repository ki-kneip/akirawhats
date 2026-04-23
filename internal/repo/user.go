package repo

import (
	"errors"

	"kikneip.com/akirawhats/internal/db"
	"kikneip.com/akirawhats/internal/model"
)

type UserImpl struct {
}

func (repo *UserImpl) GetUserByID(id string) (*model.User, error) {
	res := &model.User{}
	if err := db.Get(id, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (repo *UserImpl) UpdateUser(u *model.User) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	return db.Set(u.ID, u)
}

func (repo *UserImpl) DeleteUser(id string) error {
	return db.Delete(id)
}
