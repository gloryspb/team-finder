package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"team-finder/backend/internal/domain"
	"team-finder/backend/internal/ports"
	"team-finder/backend/internal/validation"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users     ports.UserRepository
	profiles  ports.ProfileRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

type AuthResponse struct {
	Token   string               `json:"token"`
	User    domain.User          `json:"user"`
	Profile domain.PlayerProfile `json:"profile"`
}

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileInput struct {
	Nickname  string   `json:"nickname"`
	Region    string   `json:"region"`
	Languages []string `json:"languages"`
	VoiceChat bool     `json:"voice_chat"`
	Bio       string   `json:"bio"`
}

func NewAuthService(users ports.UserRepository, profiles ports.ProfileRepository, jwtSecret string, tokenTTL time.Duration) *AuthService {
	return &AuthService{users: users, profiles: profiles, jwtSecret: []byte(jwtSecret), tokenTTL: tokenTTL}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	nickname := strings.TrimSpace(input.Nickname)
	if !validation.Email(email) || !validation.Password(input.Password) || !validation.Required(nickname) {
		return AuthResponse{}, domain.ErrInvalidInput
	}

	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return AuthResponse{}, domain.ErrConflict
	} else if !errors.Is(err, domain.ErrNotFound) {
		return AuthResponse{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{}, err
	}

	now := time.Now().UTC()
	user := domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user",
		CreatedAt:    now,
	}
	profile := domain.PlayerProfile{
		ID:        uuid.New(),
		UserID:    user.ID,
		Nickname:  nickname,
		Languages: []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return AuthResponse{}, err
	}
	if err := s.profiles.Create(ctx, profile); err != nil {
		return AuthResponse{}, err
	}

	token, err := s.issueToken(user)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, User: user, Profile: profile}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if !validation.Email(email) || input.Password == "" {
		return AuthResponse{}, domain.ErrInvalidInput
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return AuthResponse{}, domain.ErrUnauthorized
		}
		return AuthResponse{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return AuthResponse{}, domain.ErrUnauthorized
	}
	profile, err := s.profiles.GetByUserID(ctx, user.ID)
	if err != nil {
		return AuthResponse{}, err
	}
	token, err := s.issueToken(user)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, User: user, Profile: profile}, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	return s.users.GetByID(ctx, userID)
}

func (s *AuthService) Profile(ctx context.Context, userID uuid.UUID) (domain.PlayerProfile, error) {
	return s.profiles.GetByUserID(ctx, userID)
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (domain.PlayerProfile, error) {
	if !validation.Required(input.Nickname) || !validation.MaxLen(input.Bio, 1000) {
		return domain.PlayerProfile{}, domain.ErrInvalidInput
	}
	existing, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return domain.PlayerProfile{}, err
	}
	existing.Nickname = strings.TrimSpace(input.Nickname)
	existing.Region = strings.TrimSpace(input.Region)
	existing.Languages = normalizeStrings(input.Languages)
	existing.VoiceChat = input.VoiceChat
	existing.Bio = strings.TrimSpace(input.Bio)
	existing.UpdatedAt = time.Now().UTC()
	return s.profiles.Update(ctx, existing)
}

func (s *AuthService) issueToken(user domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().UTC().Add(s.tokenTTL).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

func normalizeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
