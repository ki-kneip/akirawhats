package repo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	internal_pkg "kikneip.com/akirawhats/internal"
	"kikneip.com/akirawhats/internal/db"
	"kikneip.com/akirawhats/internal/model"
)

const usersCollection = "users"

type userDoc struct {
	ID        string `bson:"_id"`
	FirstName string `bson:"first_name"`
	LastName  string `bson:"last_name"`
	Email     string `bson:"email"`
	Password  string `bson:"password"`
}

type UserImpl struct{}

func (repo *UserImpl) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	col := db.Collection(usersCollection)
	cursor, err := col.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []userDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	users := make([]*model.User, 0, len(docs))
	for i := range docs {
		users = append(users, docToUser(&docs[i]))
	}
	return users, nil
}

func (repo *UserImpl) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	col := db.Collection(usersCollection)
	var doc userDoc
	if err := col.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, internal_pkg.ErrNotFound
		}
		return nil, err
	}
	return docToUser(&doc), nil
}

func (repo *UserImpl) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	col := db.Collection(usersCollection)
	var doc userDoc
	if err := col.FindOne(ctx, bson.M{"email": email}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, internal_pkg.ErrNotFound
		}
		return nil, err
	}
	return docToUser(&doc), nil
}

func (repo *UserImpl) UpdateUser(ctx context.Context, u *model.User) error {
	col := db.Collection(usersCollection)
	opts := options.Replace().SetUpsert(true)
	_, err := col.ReplaceOne(ctx, bson.M{"_id": u.ID}, userToDoc(u), opts)
	return err
}

func (repo *UserImpl) DeleteUser(ctx context.Context, id string) error {
	col := db.Collection(usersCollection)
	result, err := col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return internal_pkg.ErrNotFound
	}
	return nil
}

func userToDoc(u *model.User) *userDoc {
	return &userDoc{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Password:  u.Password,
	}
}

func docToUser(d *userDoc) *model.User {
	return &model.User{
		ID:        d.ID,
		FirstName: d.FirstName,
		LastName:  d.LastName,
		Email:     d.Email,
		Password:  d.Password,
	}
}
