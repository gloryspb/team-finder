package services

import (
	"context"
	"strings"

	"team-finder/backend/internal/domain"
	"team-finder/backend/internal/ports"

	"github.com/google/uuid"
)

type GameService struct {
	games ports.GameRepository
}

type GameInput struct {
	Name  string   `json:"name"`
	Modes []string `json:"modes"`
	Roles []string `json:"roles"`
}

func NewGameService(games ports.GameRepository) *GameService {
	return &GameService{games: games}
}

func (s *GameService) List(ctx context.Context) ([]domain.Game, error) {
	return s.games.List(ctx)
}

func (s *GameService) Create(ctx context.Context, input GameInput) (domain.Game, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.Game{}, domain.ErrInvalidInput
	}
	game := domain.Game{ID: uuid.New(), Name: name, Modes: normalizeStrings(input.Modes), Roles: normalizeStrings(input.Roles)}
	if err := s.games.Create(ctx, game); err != nil {
		return domain.Game{}, err
	}
	return game, nil
}

func (s *GameService) Update(ctx context.Context, id uuid.UUID, input GameInput) (domain.Game, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.Game{}, domain.ErrInvalidInput
	}
	return s.games.Update(ctx, domain.Game{ID: id, Name: name, Modes: normalizeStrings(input.Modes), Roles: normalizeStrings(input.Roles)})
}

func (s *GameService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.games.Delete(ctx, id)
}
