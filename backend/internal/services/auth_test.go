package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"team-finder/backend/internal/domain"
	"team-finder/backend/internal/services"

	"github.com/google/uuid"
)

func TestAuthServiceRegisterAndLogin(t *testing.T) {
	ctx := context.Background()
	users := newFakeUsers()
	profiles := newFakeProfiles()
	service := services.NewAuthService(users, profiles, "test-secret", time.Hour)

	registered, err := service.Register(ctx, services.RegisterInput{
		Email: "Player@example.com", Password: "secret123", Nickname: "PlayerOne",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if registered.Token == "" || registered.User.Email != "player@example.com" || registered.Profile.Nickname != "PlayerOne" {
		t.Fatalf("unexpected register response: %+v", registered)
	}

	loggedIn, err := service.Login(ctx, services.LoginInput{Email: "player@example.com", Password: "secret123"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loggedIn.Token == "" || loggedIn.User.ID != registered.User.ID {
		t.Fatalf("unexpected login response: %+v", loggedIn)
	}
}

func TestAuthServiceWrongPassword(t *testing.T) {
	ctx := context.Background()
	users := newFakeUsers()
	profiles := newFakeProfiles()
	service := services.NewAuthService(users, profiles, "test-secret", time.Hour)
	if _, err := service.Register(ctx, services.RegisterInput{Email: "p@example.com", Password: "secret123", Nickname: "P"}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, err := service.Login(ctx, services.LoginInput{Email: "p@example.com", Password: "bad-password"})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestAuthServiceDuplicateEmail(t *testing.T) {
	ctx := context.Background()
	users := newFakeUsers()
	profiles := newFakeProfiles()
	service := services.NewAuthService(users, profiles, "test-secret", time.Hour)
	input := services.RegisterInput{Email: "dup@example.com", Password: "secret123", Nickname: "Dup"}
	if _, err := service.Register(ctx, input); err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	_, err := service.Register(ctx, input)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func FuzzRegister(f *testing.F) {
	f.Add("fuzz@example.com", "secret123", "Fuzzer")
	f.Add("bad-email", "123", "")
	f.Fuzz(func(t *testing.T, email, password, nickname string) {
		service := services.NewAuthService(newFakeUsers(), newFakeProfiles(), "test-secret", time.Hour)
		_, _ = service.Register(context.Background(), services.RegisterInput{Email: email, Password: password, Nickname: nickname})
	})
}

func FuzzLogin(f *testing.F) {
	f.Add("login@example.com", "secret123")
	f.Add("bad-email", "wrong")
	f.Fuzz(func(t *testing.T, email, password string) {
		ctx := context.Background()
		service := services.NewAuthService(newFakeUsers(), newFakeProfiles(), "test-secret", time.Hour)
		_, _ = service.Register(ctx, services.RegisterInput{
			Email:    "login@example.com",
			Password: "secret123",
			Nickname: "LoginFuzzer",
		})
		_, _ = service.Login(ctx, services.LoginInput{Email: email, Password: password})
	})
}

type fakeUsers struct {
	byID    map[uuid.UUID]domain.User
	byEmail map[string]domain.User
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{byID: map[uuid.UUID]domain.User{}, byEmail: map[string]domain.User{}}
}

func (r *fakeUsers) Create(_ context.Context, user domain.User) error {
	if _, ok := r.byEmail[user.Email]; ok {
		return domain.ErrConflict
	}
	r.byID[user.ID] = user
	r.byEmail[user.Email] = user
	return nil
}

func (r *fakeUsers) GetByEmail(_ context.Context, email string) (domain.User, error) {
	user, ok := r.byEmail[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (r *fakeUsers) GetByID(_ context.Context, id uuid.UUID) (domain.User, error) {
	user, ok := r.byID[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

type fakeProfiles struct {
	byUserID map[uuid.UUID]domain.PlayerProfile
}

func newFakeProfiles() *fakeProfiles {
	return &fakeProfiles{byUserID: map[uuid.UUID]domain.PlayerProfile{}}
}

func (r *fakeProfiles) Create(_ context.Context, profile domain.PlayerProfile) error {
	r.byUserID[profile.UserID] = profile
	return nil
}

func (r *fakeProfiles) GetByUserID(_ context.Context, userID uuid.UUID) (domain.PlayerProfile, error) {
	profile, ok := r.byUserID[userID]
	if !ok {
		return domain.PlayerProfile{}, domain.ErrNotFound
	}
	return profile, nil
}

func (r *fakeProfiles) Update(_ context.Context, profile domain.PlayerProfile) (domain.PlayerProfile, error) {
	r.byUserID[profile.UserID] = profile
	return profile, nil
}
