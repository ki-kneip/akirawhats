//go:build integration

package repo_test

import (
	"context"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"kikneip.com/akirawhats/internal"
	"kikneip.com/akirawhats/internal/db"
	"kikneip.com/akirawhats/internal/model"
	"kikneip.com/akirawhats/internal/repo"
)

func TestMain(m *testing.M) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	// Usa um banco isolado para testes.
	os.Setenv("MONGO_DB", "akirawhats_test")
	db.Connect(context.Background())

	code := m.Run()

	// Limpa o banco de teste após os testes.
	client.Database("akirawhats_test").Drop(context.Background())
	client.Disconnect(context.Background())
	os.Exit(code)
}

func TestUserRepo_CreateAndGet(t *testing.T) {
	r := &repo.UserImpl{}
	ctx := context.Background()

	u := &model.User{
		ID:        "test-id-1",
		FirstName: "Test",
		Email:     "test@example.com",
		Password:  "hashed",
	}
	if err := r.UpdateUser(ctx, u); err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}

	got, err := r.GetUserByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.Email != u.Email {
		t.Errorf("email: got %q, want %q", got.Email, u.Email)
	}
}

func TestUserRepo_GetByEmail(t *testing.T) {
	r := &repo.UserImpl{}
	ctx := context.Background()

	u := &model.User{
		ID:    "test-id-2",
		Email: "byemail@example.com",
	}
	if err := r.UpdateUser(ctx, u); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := r.GetUserByEmail(ctx, u.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if got.ID != u.ID {
		t.Errorf("id: got %q, want %q", got.ID, u.ID)
	}
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	r := &repo.UserImpl{}
	ctx := context.Background()

	_, err := r.GetUserByEmail(ctx, "naoexiste@example.com")
	if err != internal.ErrNotFound {
		t.Errorf("esperava ErrNotFound, got %v", err)
	}
}

func TestUserRepo_Delete(t *testing.T) {
	r := &repo.UserImpl{}
	ctx := context.Background()

	u := &model.User{ID: "test-id-3", Email: "del@example.com"}
	r.UpdateUser(ctx, u)

	if err := r.DeleteUser(ctx, u.ID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	if err := r.DeleteUser(ctx, u.ID); err != internal.ErrNotFound {
		t.Errorf("segundo delete: esperava ErrNotFound, got %v", err)
	}
}
