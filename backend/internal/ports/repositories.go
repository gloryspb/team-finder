package ports

import (
	"context"

	"team-finder/backend/internal/domain"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}

type ProfileRepository interface {
	Create(ctx context.Context, profile domain.PlayerProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (domain.PlayerProfile, error)
	Update(ctx context.Context, profile domain.PlayerProfile) (domain.PlayerProfile, error)
}

type GameRepository interface {
	List(ctx context.Context) ([]domain.Game, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Game, error)
	Create(ctx context.Context, game domain.Game) error
	Update(ctx context.Context, game domain.Game) (domain.Game, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ListingRepository interface {
	List(ctx context.Context, filters domain.ListingFilters) ([]domain.ListingDetails, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.ListingDetails, error)
	Create(ctx context.Context, listing domain.Listing) error
	Update(ctx context.Context, listing domain.Listing) (domain.Listing, error)
	Close(ctx context.Context, id uuid.UUID) (domain.Listing, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ApplicationRepository interface {
	Create(ctx context.Context, application domain.Application) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.ApplicationDetails, error)
	GetByListingAndApplicant(ctx context.Context, listingID, applicantID uuid.UUID) (domain.Application, error)
	ListOutgoing(ctx context.Context, applicantID uuid.UUID) ([]domain.ApplicationDetails, error)
	ListIncoming(ctx context.Context, ownerID uuid.UUID) ([]domain.ApplicationDetails, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (domain.Application, error)
}
