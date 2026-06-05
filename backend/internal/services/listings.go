package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"team-finder/backend/internal/domain"
	"team-finder/backend/internal/ports"
	"team-finder/backend/internal/validation"

	"github.com/google/uuid"
)

type ListingService struct {
	listings     ports.ListingRepository
	applications ports.ApplicationRepository
}

type ListingInput struct {
	GameID        string   `json:"game_id"`
	Title         string   `json:"title"`
	Mode          string   `json:"mode"`
	RequiredRoles []string `json:"required_roles"`
	RankMin       string   `json:"rank_min"`
	RankMax       string   `json:"rank_max"`
	Region        string   `json:"region"`
	Description   string   `json:"description"`
}

type ApplicationInput struct {
	Message string `json:"message"`
}

func NewListingService(listings ports.ListingRepository, applications ports.ApplicationRepository) *ListingService {
	return &ListingService{listings: listings, applications: applications}
}

func (s *ListingService) List(ctx context.Context, filters domain.ListingFilters) ([]domain.ListingDetails, error) {
	if filters.Status == "" {
		filters.Status = "open"
	}
	if !validation.ListingStatus(filters.Status) {
		return nil, domain.ErrInvalidInput
	}
	return s.listings.List(ctx, filters)
}

func (s *ListingService) Get(ctx context.Context, id uuid.UUID) (domain.ListingDetails, error) {
	return s.listings.GetByID(ctx, id)
}

func (s *ListingService) Create(ctx context.Context, ownerID uuid.UUID, input ListingInput) (domain.Listing, error) {
	gameID, err := uuid.Parse(input.GameID)
	if err != nil {
		return domain.Listing{}, domain.ErrInvalidInput
	}
	if !validation.Required(input.Title) || !validation.MaxLen(input.Description, 2000) {
		return domain.Listing{}, domain.ErrInvalidInput
	}
	now := time.Now().UTC()
	listing := domain.Listing{
		ID:            uuid.New(),
		OwnerID:       ownerID,
		GameID:        gameID,
		Title:         strings.TrimSpace(input.Title),
		Mode:          strings.TrimSpace(input.Mode),
		RequiredRoles: normalizeStrings(input.RequiredRoles),
		RankMin:       strings.TrimSpace(input.RankMin),
		RankMax:       strings.TrimSpace(input.RankMax),
		Region:        strings.TrimSpace(input.Region),
		Description:   strings.TrimSpace(input.Description),
		Status:        "open",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.listings.Create(ctx, listing); err != nil {
		return domain.Listing{}, err
	}
	return listing, nil
}

func (s *ListingService) Update(ctx context.Context, actorID uuid.UUID, actorRole string, listingID uuid.UUID, input ListingInput) (domain.Listing, error) {
	details, err := s.listings.GetByID(ctx, listingID)
	if err != nil {
		return domain.Listing{}, err
	}
	if !canManage(actorID, actorRole, details.OwnerID) {
		return domain.Listing{}, domain.ErrForbidden
	}
	gameID, err := uuid.Parse(input.GameID)
	if err != nil {
		return domain.Listing{}, domain.ErrInvalidInput
	}
	if !validation.Required(input.Title) || !validation.MaxLen(input.Description, 2000) {
		return domain.Listing{}, domain.ErrInvalidInput
	}
	listing := details.Listing
	listing.GameID = gameID
	listing.Title = strings.TrimSpace(input.Title)
	listing.Mode = strings.TrimSpace(input.Mode)
	listing.RequiredRoles = normalizeStrings(input.RequiredRoles)
	listing.RankMin = strings.TrimSpace(input.RankMin)
	listing.RankMax = strings.TrimSpace(input.RankMax)
	listing.Region = strings.TrimSpace(input.Region)
	listing.Description = strings.TrimSpace(input.Description)
	listing.UpdatedAt = time.Now().UTC()
	return s.listings.Update(ctx, listing)
}

func (s *ListingService) Close(ctx context.Context, actorID uuid.UUID, actorRole string, listingID uuid.UUID) (domain.Listing, error) {
	details, err := s.listings.GetByID(ctx, listingID)
	if err != nil {
		return domain.Listing{}, err
	}
	if !canManage(actorID, actorRole, details.OwnerID) {
		return domain.Listing{}, domain.ErrForbidden
	}
	return s.listings.Close(ctx, listingID)
}

func (s *ListingService) Delete(ctx context.Context, actorID uuid.UUID, actorRole string, listingID uuid.UUID) error {
	details, err := s.listings.GetByID(ctx, listingID)
	if err != nil {
		return err
	}
	if !canManage(actorID, actorRole, details.OwnerID) {
		return domain.ErrForbidden
	}
	return s.listings.Delete(ctx, listingID)
}

func (s *ListingService) Apply(ctx context.Context, applicantID uuid.UUID, listingID uuid.UUID, input ApplicationInput) (domain.Application, error) {
	details, err := s.listings.GetByID(ctx, listingID)
	if err != nil {
		return domain.Application{}, err
	}
	if details.Status != "open" {
		return domain.Application{}, domain.ErrClosedListing
	}
	if details.OwnerID == applicantID {
		return domain.Application{}, domain.ErrOwnListing
	}
	if _, err := s.applications.GetByListingAndApplicant(ctx, listingID, applicantID); err == nil {
		return domain.Application{}, domain.ErrDuplicate
	} else if !errors.Is(err, domain.ErrNotFound) {
		return domain.Application{}, err
	}
	now := time.Now().UTC()
	application := domain.Application{
		ID:          uuid.New(),
		ListingID:   listingID,
		ApplicantID: applicantID,
		Message:     strings.TrimSpace(input.Message),
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.applications.Create(ctx, application); err != nil {
		return domain.Application{}, err
	}
	return application, nil
}

func (s *ListingService) Outgoing(ctx context.Context, applicantID uuid.UUID) ([]domain.ApplicationDetails, error) {
	return s.applications.ListOutgoing(ctx, applicantID)
}

func (s *ListingService) Incoming(ctx context.Context, ownerID uuid.UUID) ([]domain.ApplicationDetails, error) {
	return s.applications.ListIncoming(ctx, ownerID)
}

func (s *ListingService) UpdateApplicationStatus(ctx context.Context, actorID uuid.UUID, actorRole string, applicationID uuid.UUID, status string) (domain.Application, error) {
	if !validation.ApplicationStatus(status) {
		return domain.Application{}, domain.ErrInvalidInput
	}
	details, err := s.applications.GetByID(ctx, applicationID)
	if err != nil {
		return domain.Application{}, err
	}
	if !canManage(actorID, actorRole, details.Listing.OwnerID) {
		return domain.Application{}, domain.ErrForbidden
	}
	return s.applications.UpdateStatus(ctx, applicationID, status)
}

func canManage(actorID uuid.UUID, actorRole string, ownerID uuid.UUID) bool {
	return actorRole == "admin" || actorID == ownerID
}
