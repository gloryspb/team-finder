package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"team-finder/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct{ db *pgxpool.Pool }
type ProfileRepository struct{ db *pgxpool.Pool }
type GameRepository struct{ db *pgxpool.Pool }
type ListingRepository struct{ db *pgxpool.Pool }
type ApplicationRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) *UserRepository       { return &UserRepository{db: db} }
func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository { return &ProfileRepository{db: db} }
func NewGameRepository(db *pgxpool.Pool) *GameRepository       { return &GameRepository{db: db} }
func NewListingRepository(db *pgxpool.Pool) *ListingRepository { return &ListingRepository{db: db} }
func NewApplicationRepository(db *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, user.ID, user.Email, user.PasswordHash, user.Role, user.CreatedAt)
	return mapErr(err)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, role, created_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	return user, mapErr(err)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, role, created_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	return user, mapErr(err)
}

func (r *ProfileRepository) Create(ctx context.Context, profile domain.PlayerProfile) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO player_profiles (id, user_id, nickname, region, languages, voice_chat, bio, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, profile.ID, profile.UserID, profile.Nickname, profile.Region, profile.Languages, profile.VoiceChat, profile.Bio, profile.CreatedAt, profile.UpdatedAt)
	return mapErr(err)
}

func (r *ProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (domain.PlayerProfile, error) {
	var profile domain.PlayerProfile
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, nickname, region, languages, voice_chat, bio, created_at, updated_at
		FROM player_profiles WHERE user_id = $1
	`, userID).Scan(&profile.ID, &profile.UserID, &profile.Nickname, &profile.Region, &profile.Languages, &profile.VoiceChat, &profile.Bio, &profile.CreatedAt, &profile.UpdatedAt)
	return profile, mapErr(err)
}

func (r *ProfileRepository) Update(ctx context.Context, profile domain.PlayerProfile) (domain.PlayerProfile, error) {
	err := r.db.QueryRow(ctx, `
		UPDATE player_profiles
		SET nickname = $2, region = $3, languages = $4, voice_chat = $5, bio = $6, updated_at = $7
		WHERE user_id = $1
		RETURNING id, user_id, nickname, region, languages, voice_chat, bio, created_at, updated_at
	`, profile.UserID, profile.Nickname, profile.Region, profile.Languages, profile.VoiceChat, profile.Bio, profile.UpdatedAt).
		Scan(&profile.ID, &profile.UserID, &profile.Nickname, &profile.Region, &profile.Languages, &profile.VoiceChat, &profile.Bio, &profile.CreatedAt, &profile.UpdatedAt)
	return profile, mapErr(err)
}

func (r *GameRepository) List(ctx context.Context) ([]domain.Game, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, modes, roles FROM games ORDER BY name`)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()

	games := []domain.Game{}
	for rows.Next() {
		var game domain.Game
		if err := rows.Scan(&game.ID, &game.Name, &game.Modes, &game.Roles); err != nil {
			return nil, mapErr(err)
		}
		games = append(games, game)
	}
	return games, mapErr(rows.Err())
}

func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Game, error) {
	var game domain.Game
	err := r.db.QueryRow(ctx, `SELECT id, name, modes, roles FROM games WHERE id = $1`, id).
		Scan(&game.ID, &game.Name, &game.Modes, &game.Roles)
	return game, mapErr(err)
}

func (r *GameRepository) Create(ctx context.Context, game domain.Game) error {
	_, err := r.db.Exec(ctx, `INSERT INTO games (id, name, modes, roles) VALUES ($1, $2, $3, $4)`, game.ID, game.Name, game.Modes, game.Roles)
	return mapErr(err)
}

func (r *GameRepository) Update(ctx context.Context, game domain.Game) (domain.Game, error) {
	err := r.db.QueryRow(ctx, `
		UPDATE games SET name = $2, modes = $3, roles = $4
		WHERE id = $1
		RETURNING id, name, modes, roles
	`, game.ID, game.Name, game.Modes, game.Roles).Scan(&game.ID, &game.Name, &game.Modes, &game.Roles)
	return game, mapErr(err)
}

func (r *GameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM games WHERE id = $1`, id)
	if err := mapErr(err); err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ListingRepository) List(ctx context.Context, filters domain.ListingFilters) ([]domain.ListingDetails, error) {
	query := baseListingSelect()
	conditions := []string{}
	args := []any{}
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if filters.GameID != "" {
		add("l.game_id = $%d", filters.GameID)
	}
	if filters.Role != "" {
		add("$%d = ANY(l.required_roles)", filters.Role)
	}
	if filters.Region != "" {
		add("LOWER(l.region) = LOWER($%d)", filters.Region)
	}
	if filters.Mode != "" {
		add("LOWER(l.mode) = LOWER($%d)", filters.Mode)
	}
	if filters.Status != "" {
		add("l.status = $%d", filters.Status)
	}
	if filters.Search != "" {
		args = append(args, "%"+filters.Search+"%")
		index := len(args)
		conditions = append(conditions, fmt.Sprintf("(l.title ILIKE $%d OR l.description ILIKE $%d)", index, index))
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY l.created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()

	listings := []domain.ListingDetails{}
	for rows.Next() {
		item, err := scanListingDetails(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, item)
	}
	return listings, mapErr(rows.Err())
}

func (r *ListingRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.ListingDetails, error) {
	item, err := scanListingDetails(r.db.QueryRow(ctx, baseListingSelect()+" WHERE l.id = $1", id))
	return item, mapErr(err)
}

func (r *ListingRepository) Create(ctx context.Context, listing domain.Listing) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO listings (id, owner_id, game_id, title, mode, required_roles, rank_min, rank_max, region, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, listing.ID, listing.OwnerID, listing.GameID, listing.Title, listing.Mode, listing.RequiredRoles, listing.RankMin, listing.RankMax, listing.Region, listing.Description, listing.Status, listing.CreatedAt, listing.UpdatedAt)
	return mapErr(err)
}

func (r *ListingRepository) Update(ctx context.Context, listing domain.Listing) (domain.Listing, error) {
	err := r.db.QueryRow(ctx, `
		UPDATE listings
		SET game_id = $2, title = $3, mode = $4, required_roles = $5, rank_min = $6, rank_max = $7,
			region = $8, description = $9, updated_at = $10
		WHERE id = $1
		RETURNING id, owner_id, game_id, title, mode, required_roles, rank_min, rank_max, region, description, status, created_at, updated_at
	`, listing.ID, listing.GameID, listing.Title, listing.Mode, listing.RequiredRoles, listing.RankMin, listing.RankMax, listing.Region, listing.Description, listing.UpdatedAt).
		Scan(&listing.ID, &listing.OwnerID, &listing.GameID, &listing.Title, &listing.Mode, &listing.RequiredRoles, &listing.RankMin, &listing.RankMax, &listing.Region, &listing.Description, &listing.Status, &listing.CreatedAt, &listing.UpdatedAt)
	return listing, mapErr(err)
}

func (r *ListingRepository) Close(ctx context.Context, id uuid.UUID) (domain.Listing, error) {
	var listing domain.Listing
	err := r.db.QueryRow(ctx, `
		UPDATE listings SET status = 'closed', updated_at = NOW()
		WHERE id = $1
		RETURNING id, owner_id, game_id, title, mode, required_roles, rank_min, rank_max, region, description, status, created_at, updated_at
	`, id).Scan(&listing.ID, &listing.OwnerID, &listing.GameID, &listing.Title, &listing.Mode, &listing.RequiredRoles, &listing.RankMin, &listing.RankMax, &listing.Region, &listing.Description, &listing.Status, &listing.CreatedAt, &listing.UpdatedAt)
	return listing, mapErr(err)
}

func (r *ListingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM listings WHERE id = $1`, id)
	if err := mapErr(err); err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ApplicationRepository) Create(ctx context.Context, application domain.Application) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO applications (id, listing_id, applicant_id, message, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, application.ID, application.ListingID, application.ApplicantID, application.Message, application.Status, application.CreatedAt, application.UpdatedAt)
	return mapErr(err)
}

func (r *ApplicationRepository) GetByListingAndApplicant(ctx context.Context, listingID, applicantID uuid.UUID) (domain.Application, error) {
	var application domain.Application
	err := r.db.QueryRow(ctx, `
		SELECT id, listing_id, applicant_id, message, status, created_at, updated_at
		FROM applications WHERE listing_id = $1 AND applicant_id = $2
	`, listingID, applicantID).Scan(&application.ID, &application.ListingID, &application.ApplicantID, &application.Message, &application.Status, &application.CreatedAt, &application.UpdatedAt)
	return application, mapErr(err)
}

func (r *ApplicationRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.ApplicationDetails, error) {
	item, err := scanApplicationDetails(r.db.QueryRow(ctx, baseApplicationSelect()+" WHERE a.id = $1", id))
	return item, mapErr(err)
}

func (r *ApplicationRepository) ListOutgoing(ctx context.Context, applicantID uuid.UUID) ([]domain.ApplicationDetails, error) {
	rows, err := r.db.Query(ctx, baseApplicationSelect()+" WHERE a.applicant_id = $1 ORDER BY a.created_at DESC", applicantID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	return scanApplications(rows)
}

func (r *ApplicationRepository) ListIncoming(ctx context.Context, ownerID uuid.UUID) ([]domain.ApplicationDetails, error) {
	rows, err := r.db.Query(ctx, baseApplicationSelect()+" WHERE l.owner_id = $1 ORDER BY a.created_at DESC", ownerID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	return scanApplications(rows)
}

func (r *ApplicationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (domain.Application, error) {
	var application domain.Application
	err := r.db.QueryRow(ctx, `
		UPDATE applications SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, listing_id, applicant_id, message, status, created_at, updated_at
	`, id, status).Scan(&application.ID, &application.ListingID, &application.ApplicantID, &application.Message, &application.Status, &application.CreatedAt, &application.UpdatedAt)
	return application, mapErr(err)
}

func scanApplications(rows pgx.Rows) ([]domain.ApplicationDetails, error) {
	items := []domain.ApplicationDetails{}
	for rows.Next() {
		item, err := scanApplicationDetails(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, mapErr(rows.Err())
}

type scanner interface {
	Scan(dest ...any) error
}

func scanListingDetails(row scanner) (domain.ListingDetails, error) {
	var item domain.ListingDetails
	err := row.Scan(
		&item.ID, &item.OwnerID, &item.GameID, &item.Title, &item.Mode, &item.RequiredRoles, &item.RankMin, &item.RankMax,
		&item.Region, &item.Description, &item.Status, &item.CreatedAt, &item.UpdatedAt,
		&item.Game.ID, &item.Game.Name, &item.Game.Modes, &item.Game.Roles,
		&item.Owner.ID, &item.Owner.Email, &item.Owner.Role, &item.Owner.CreatedAt,
		&item.OwnerProfile.ID, &item.OwnerProfile.UserID, &item.OwnerProfile.Nickname, &item.OwnerProfile.Region,
		&item.OwnerProfile.Languages, &item.OwnerProfile.VoiceChat, &item.OwnerProfile.Bio, &item.OwnerProfile.CreatedAt, &item.OwnerProfile.UpdatedAt,
		&item.ApplicationsCount,
	)
	return item, err
}

func scanApplicationDetails(row scanner) (domain.ApplicationDetails, error) {
	var item domain.ApplicationDetails
	err := row.Scan(
		&item.ID, &item.ListingID, &item.ApplicantID, &item.Message, &item.Status, &item.CreatedAt, &item.UpdatedAt,
		&item.Listing.ID, &item.Listing.OwnerID, &item.Listing.GameID, &item.Listing.Title, &item.Listing.Mode, &item.Listing.RequiredRoles,
		&item.Listing.RankMin, &item.Listing.RankMax, &item.Listing.Region, &item.Listing.Description, &item.Listing.Status,
		&item.Listing.CreatedAt, &item.Listing.UpdatedAt,
		&item.Game.ID, &item.Game.Name, &item.Game.Modes, &item.Game.Roles,
		&item.Applicant.ID, &item.Applicant.Email, &item.Applicant.Role, &item.Applicant.CreatedAt,
		&item.ApplicantProfile.ID, &item.ApplicantProfile.UserID, &item.ApplicantProfile.Nickname, &item.ApplicantProfile.Region,
		&item.ApplicantProfile.Languages, &item.ApplicantProfile.VoiceChat, &item.ApplicantProfile.Bio, &item.ApplicantProfile.CreatedAt, &item.ApplicantProfile.UpdatedAt,
	)
	return item, err
}

func baseListingSelect() string {
	return `
		SELECT
			l.id, l.owner_id, l.game_id, l.title, l.mode, l.required_roles, l.rank_min, l.rank_max,
			l.region, l.description, l.status, l.created_at, l.updated_at,
			g.id, g.name, g.modes, g.roles,
			u.id, u.email, u.role, u.created_at,
			p.id, p.user_id, p.nickname, p.region, p.languages, p.voice_chat, p.bio, p.created_at, p.updated_at,
			(SELECT COUNT(*)::int FROM applications a WHERE a.listing_id = l.id) AS applications_count
		FROM listings l
		JOIN games g ON g.id = l.game_id
		JOIN users u ON u.id = l.owner_id
		JOIN player_profiles p ON p.user_id = u.id
	`
}

func baseApplicationSelect() string {
	return `
		SELECT
			a.id, a.listing_id, a.applicant_id, a.message, a.status, a.created_at, a.updated_at,
			l.id, l.owner_id, l.game_id, l.title, l.mode, l.required_roles, l.rank_min, l.rank_max,
			l.region, l.description, l.status, l.created_at, l.updated_at,
			g.id, g.name, g.modes, g.roles,
			u.id, u.email, u.role, u.created_at,
			p.id, p.user_id, p.nickname, p.region, p.languages, p.voice_chat, p.bio, p.created_at, p.updated_at
		FROM applications a
		JOIN listings l ON l.id = a.listing_id
		JOIN games g ON g.id = l.game_id
		JOIN users u ON u.id = a.applicant_id
		JOIN player_profiles p ON p.user_id = u.id
	`
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return domain.ErrConflict
		case "23503", "22P02":
			return domain.ErrInvalidInput
		}
	}
	return err
}
