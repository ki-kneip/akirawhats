package service_test

import (
	"context"
	"testing"

	"kikneip.com/akirawhats/internal"
	"kikneip.com/akirawhats/internal/model"
	"kikneip.com/akirawhats/internal/service"
)

// mockUserRepo implementa internal.UserRepo em memória.
type mockUserRepo struct {
	byID    map[string]*model.User
	byEmail map[string]*model.User
}

func newMockRepo() *mockUserRepo {
	return &mockUserRepo{
		byID:    make(map[string]*model.User),
		byEmail: make(map[string]*model.User),
	}
}

func (m *mockUserRepo) GetAllUsers(_ context.Context) ([]*model.User, error) {
	users := make([]*model.User, 0, len(m.byID))
	for _, u := range m.byID {
		users = append(users, u)
	}
	return users, nil
}

func (m *mockUserRepo) GetUserByID(_ context.Context, id string) (*model.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, internal.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetUserByEmail(_ context.Context, email string) (*model.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, internal.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) UpdateUser(_ context.Context, u *model.User) error {
	m.byID[u.ID] = u
	m.byEmail[u.Email] = u
	return nil
}

func (m *mockUserRepo) DeleteUser(_ context.Context, id string) error {
	u, ok := m.byID[id]
	if !ok {
		return internal.ErrNotFound
	}
	delete(m.byEmail, u.Email)
	delete(m.byID, id)
	return nil
}

func TestCreateUser(t *testing.T) {
	svc := service.NewUserService(newMockRepo())
	ctx := context.Background()

	dto, err := svc.CreateUser(ctx, model.UserDTOPost{
		FirstName: "João",
		LastName:  "Silva",
		Email:     "joao@example.com",
		Password:  "senha123",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if dto.ID == "" {
		t.Error("ID não deve ser vazio")
	}
	if dto.Email != "joao@example.com" {
		t.Errorf("email: got %q, want %q", dto.Email, "joao@example.com")
	}
}

func TestAuthenticateUser_Success(t *testing.T) {
	svc := service.NewUserService(newMockRepo())
	ctx := context.Background()

	if _, err := svc.CreateUser(ctx, model.UserDTOPost{
		Email:    "user@example.com",
		Password: "secreto",
	}); err != nil {
		t.Fatalf("setup CreateUser: %v", err)
	}

	dto, err := svc.AuthenticateUser(ctx, "user@example.com", "secreto")
	if err != nil {
		t.Fatalf("AuthenticateUser: %v", err)
	}
	if dto.Email != "user@example.com" {
		t.Errorf("email: got %q", dto.Email)
	}
}

func TestAuthenticateUser_WrongPassword(t *testing.T) {
	svc := service.NewUserService(newMockRepo())
	ctx := context.Background()

	if _, err := svc.CreateUser(ctx, model.UserDTOPost{
		Email:    "user@example.com",
		Password: "correta",
	}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := svc.AuthenticateUser(ctx, "user@example.com", "errada")
	if err == nil {
		t.Fatal("esperava erro, obteve nil")
	}
	if err != internal.ErrInvalidCredentials {
		t.Errorf("erro esperado ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticateUser_NotFound(t *testing.T) {
	svc := service.NewUserService(newMockRepo())
	ctx := context.Background()

	_, err := svc.AuthenticateUser(ctx, "naoexiste@example.com", "qualquer")
	if err != internal.ErrInvalidCredentials {
		t.Errorf("erro esperado ErrInvalidCredentials, got %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	svc := service.NewUserService(newMockRepo())
	ctx := context.Background()

	dto, _ := svc.CreateUser(ctx, model.UserDTOPost{Email: "del@example.com", Password: "123456"})

	if err := svc.DeleteUser(ctx, dto.ID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	if err := svc.DeleteUser(ctx, dto.ID); err != internal.ErrNotFound {
		t.Errorf("segundo delete: esperava ErrNotFound, got %v", err)
	}
}

func TestUpdateUser_Partial(t *testing.T) {
	svc := service.NewUserService(newMockRepo())
	ctx := context.Background()

	dto, _ := svc.CreateUser(ctx, model.UserDTOPost{
		FirstName: "Ana",
		Email:     "ana@example.com",
		Password:  "123456",
	})

	updated, err := svc.UpdateUser(ctx, dto.ID, model.UserDTOPut{FirstName: "Maria"})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if updated.FirstName != "Maria" {
		t.Errorf("FirstName: got %q, want Maria", updated.FirstName)
	}
	if updated.Email != "ana@example.com" {
		t.Error("email não deveria ter mudado")
	}
}
