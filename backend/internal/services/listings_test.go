package services_test

import (
	"context"
	"errors"
	"testing"

	"team-finder/backend/internal/domain"
	"team-finder/backend/internal/services"

	"github.com/google/uuid"
)

func TestListingServiceCreate(t *testing.T) {
	ctx := context.Background()
	listings := newFakeListings()
	service := services.NewListingService(listings, newFakeApplications())
	ownerID := uuid.New()
	gameID := uuid.New()

	listing, err := service.Create(ctx, ownerID, services.ListingInput{
		GameID: gameID.String(), Title: "Ищем саппорта", Mode: "Ranked", RequiredRoles: []string{"Support"}, Region: "EU",
	})
	if err != nil {
		t.Fatalf("create listing failed: %v", err)
	}
	if listing.OwnerID != ownerID || listing.GameID != gameID || listing.Status != "open" {
		t.Fatalf("unexpected listing: %+v", listing)
	}
}

func TestListingServiceApplyRejectsOwnListing(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	listingID := uuid.New()
	listings := newFakeListings()
	listings.details[listingID] = listingDetails(listingID, ownerID, "open")
	service := services.NewListingService(listings, newFakeApplications())

	_, err := service.Apply(ctx, ownerID, listingID, services.ApplicationInput{Message: "Возьмите меня"})
	if !errors.Is(err, domain.ErrOwnListing) {
		t.Fatalf("expected own listing error, got %v", err)
	}
}

func TestListingServiceApplyRejectsDuplicate(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	applicantID := uuid.New()
	listingID := uuid.New()
	listings := newFakeListings()
	listings.details[listingID] = listingDetails(listingID, ownerID, "open")
	applications := newFakeApplications()
	service := services.NewListingService(listings, applications)

	if _, err := service.Apply(ctx, applicantID, listingID, services.ApplicationInput{}); err != nil {
		t.Fatalf("first apply failed: %v", err)
	}
	_, err := service.Apply(ctx, applicantID, listingID, services.ApplicationInput{})
	if !errors.Is(err, domain.ErrDuplicate) {
		t.Fatalf("expected duplicate, got %v", err)
	}
}

func TestListingServiceUpdateApplicationStatus(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	applicantID := uuid.New()
	listingID := uuid.New()
	applicationID := uuid.New()
	applications := newFakeApplications()
	applications.details[applicationID] = domain.ApplicationDetails{
		Application: domain.Application{ID: applicationID, ListingID: listingID, ApplicantID: applicantID, Status: "pending"},
		Listing:     domain.Listing{ID: listingID, OwnerID: ownerID},
	}
	service := services.NewListingService(newFakeListings(), applications)

	updated, err := service.UpdateApplicationStatus(ctx, ownerID, "user", applicationID, "accepted")
	if err != nil {
		t.Fatalf("update status failed: %v", err)
	}
	if updated.Status != "accepted" {
		t.Fatalf("expected accepted, got %s", updated.Status)
	}
}

func FuzzCreateListing(f *testing.F) {
	f.Add("Ищем игрока", "Описание", "EU", "Ranked")
	f.Add("", "", "", "")
	f.Fuzz(func(t *testing.T, title, description, region, mode string) {
		service := services.NewListingService(newFakeListings(), newFakeApplications())
		_, _ = service.Create(context.Background(), uuid.New(), services.ListingInput{
			GameID: uuid.New().String(), Title: title, Description: description, Region: region, Mode: mode,
		})
	})
}

func FuzzApplicationStatus(f *testing.F) {
	f.Add("accepted")
	f.Add("rejected")
	f.Add("pending")
	f.Fuzz(func(t *testing.T, status string) {
		ownerID := uuid.New()
		applicationID := uuid.New()
		applications := newFakeApplications()
		applications.details[applicationID] = domain.ApplicationDetails{
			Application: domain.Application{ID: applicationID, Status: "pending"},
			Listing:     domain.Listing{OwnerID: ownerID},
		}
		service := services.NewListingService(newFakeListings(), applications)
		_, _ = service.UpdateApplicationStatus(context.Background(), ownerID, "user", applicationID, status)
	})
}

type fakeListings struct {
	details map[uuid.UUID]domain.ListingDetails
}

func newFakeListings() *fakeListings {
	return &fakeListings{details: map[uuid.UUID]domain.ListingDetails{}}
}

func (r *fakeListings) List(_ context.Context, _ domain.ListingFilters) ([]domain.ListingDetails, error) {
	items := make([]domain.ListingDetails, 0, len(r.details))
	for _, item := range r.details {
		items = append(items, item)
	}
	return items, nil
}

func (r *fakeListings) GetByID(_ context.Context, id uuid.UUID) (domain.ListingDetails, error) {
	item, ok := r.details[id]
	if !ok {
		return domain.ListingDetails{}, domain.ErrNotFound
	}
	return item, nil
}

func (r *fakeListings) Create(_ context.Context, listing domain.Listing) error {
	r.details[listing.ID] = domain.ListingDetails{Listing: listing}
	return nil
}

func (r *fakeListings) Update(_ context.Context, listing domain.Listing) (domain.Listing, error) {
	r.details[listing.ID] = domain.ListingDetails{Listing: listing}
	return listing, nil
}

func (r *fakeListings) Close(_ context.Context, id uuid.UUID) (domain.Listing, error) {
	item, ok := r.details[id]
	if !ok {
		return domain.Listing{}, domain.ErrNotFound
	}
	item.Status = "closed"
	r.details[id] = item
	return item.Listing, nil
}

func (r *fakeListings) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.details, id)
	return nil
}

type fakeApplications struct {
	details map[uuid.UUID]domain.ApplicationDetails
}

func newFakeApplications() *fakeApplications {
	return &fakeApplications{details: map[uuid.UUID]domain.ApplicationDetails{}}
}

func (r *fakeApplications) Create(_ context.Context, application domain.Application) error {
	for _, item := range r.details {
		if item.ListingID == application.ListingID && item.ApplicantID == application.ApplicantID {
			return domain.ErrConflict
		}
	}
	r.details[application.ID] = domain.ApplicationDetails{Application: application}
	return nil
}

func (r *fakeApplications) GetByID(_ context.Context, id uuid.UUID) (domain.ApplicationDetails, error) {
	item, ok := r.details[id]
	if !ok {
		return domain.ApplicationDetails{}, domain.ErrNotFound
	}
	return item, nil
}

func (r *fakeApplications) GetByListingAndApplicant(_ context.Context, listingID, applicantID uuid.UUID) (domain.Application, error) {
	for _, item := range r.details {
		if item.ListingID == listingID && item.ApplicantID == applicantID {
			return item.Application, nil
		}
	}
	return domain.Application{}, domain.ErrNotFound
}

func (r *fakeApplications) ListOutgoing(_ context.Context, applicantID uuid.UUID) ([]domain.ApplicationDetails, error) {
	var items []domain.ApplicationDetails
	for _, item := range r.details {
		if item.ApplicantID == applicantID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *fakeApplications) ListIncoming(_ context.Context, ownerID uuid.UUID) ([]domain.ApplicationDetails, error) {
	var items []domain.ApplicationDetails
	for _, item := range r.details {
		if item.Listing.OwnerID == ownerID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *fakeApplications) UpdateStatus(_ context.Context, id uuid.UUID, status string) (domain.Application, error) {
	item, ok := r.details[id]
	if !ok {
		return domain.Application{}, domain.ErrNotFound
	}
	item.Status = status
	item.Application.Status = status
	r.details[id] = item
	return item.Application, nil
}

func listingDetails(listingID, ownerID uuid.UUID, status string) domain.ListingDetails {
	return domain.ListingDetails{Listing: domain.Listing{ID: listingID, OwnerID: ownerID, Status: status}}
}
