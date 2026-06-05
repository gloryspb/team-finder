package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type PlayerProfile struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Region    string    `json:"region"`
	Languages []string  `json:"languages"`
	VoiceChat bool      `json:"voice_chat"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Game struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Modes []string  `json:"modes"`
	Roles []string  `json:"roles"`
}

type Listing struct {
	ID            uuid.UUID `json:"id"`
	OwnerID       uuid.UUID `json:"owner_id"`
	GameID        uuid.UUID `json:"game_id"`
	Title         string    `json:"title"`
	Mode          string    `json:"mode"`
	RequiredRoles []string  `json:"required_roles"`
	RankMin       string    `json:"rank_min"`
	RankMax       string    `json:"rank_max"`
	Region        string    `json:"region"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ListingDetails struct {
	Listing
	Game              Game          `json:"game"`
	Owner             User          `json:"owner"`
	OwnerProfile      PlayerProfile `json:"owner_profile"`
	ApplicationsCount int           `json:"applications_count"`
}

type ListingFilters struct {
	GameID string
	Role   string
	Region string
	Mode   string
	Status string
	Search string
}

type Application struct {
	ID          uuid.UUID `json:"id"`
	ListingID   uuid.UUID `json:"listing_id"`
	ApplicantID uuid.UUID `json:"applicant_id"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ApplicationDetails struct {
	Application
	Listing          Listing       `json:"listing"`
	Game             Game          `json:"game"`
	Applicant        User          `json:"applicant"`
	ApplicantProfile PlayerProfile `json:"applicant_profile"`
}
