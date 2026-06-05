package http

import (
	"errors"
	stdhttp "net/http"

	"team-finder/backend/internal/domain"
	"team-finder/backend/internal/middleware"
	"team-finder/backend/internal/services"
	"team-finder/backend/internal/validation"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	auth     *services.AuthService
	games    *services.GameService
	listings *services.ListingService
}

func New(auth *services.AuthService, games *services.GameService, listings *services.ListingService) *Handler {
	return &Handler{auth: auth, games: games, listings: listings}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, jwtSecret string) {
	e.HTTPErrorHandler = errorHandler

	api := e.Group("/api")
	api.POST("/auth/register", h.register)
	api.POST("/auth/login", h.login)
	api.GET("/games", h.listGames)
	api.GET("/listings", h.listListings)
	api.GET("/listings/:id", h.getListing)

	protected := api.Group("", middleware.JWT(jwtSecret))
	protected.GET("/me", h.me)
	protected.GET("/me/profile", h.profile)
	protected.PUT("/me/profile", h.updateProfile)
	protected.POST("/listings", h.createListing)
	protected.PUT("/listings/:id", h.updateListing)
	protected.PATCH("/listings/:id/close", h.closeListing)
	protected.DELETE("/listings/:id", h.deleteListing)
	protected.POST("/listings/:id/applications", h.apply)
	protected.GET("/applications/outgoing", h.outgoing)
	protected.GET("/applications/incoming", h.incoming)
	protected.PATCH("/applications/:id/status", h.updateApplicationStatus)

	admin := protected.Group("", middleware.AdminOnly)
	admin.POST("/games", h.createGame)
	admin.PUT("/games/:id", h.updateGame)
	admin.DELETE("/games/:id", h.deleteGame)
}

func (h *Handler) register(c echo.Context) error {
	var input services.RegisterInput
	if err := c.Bind(&input); err != nil {
		return domain.ErrInvalidInput
	}
	response, err := h.auth.Register(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusCreated, response)
}

func (h *Handler) login(c echo.Context) error {
	var input services.LoginInput
	if err := c.Bind(&input); err != nil {
		return domain.ErrInvalidInput
	}
	response, err := h.auth.Login(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, response)
}

func (h *Handler) me(c echo.Context) error {
	user, err := h.auth.CurrentUser(c.Request().Context(), middleware.UserID(c))
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, user)
}

func (h *Handler) profile(c echo.Context) error {
	profile, err := h.auth.Profile(c.Request().Context(), middleware.UserID(c))
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, profile)
}

func (h *Handler) updateProfile(c echo.Context) error {
	var input services.UpdateProfileInput
	if err := c.Bind(&input); err != nil {
		return domain.ErrInvalidInput
	}
	profile, err := h.auth.UpdateProfile(c.Request().Context(), middleware.UserID(c), input)
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, profile)
}

func (h *Handler) listGames(c echo.Context) error {
	games, err := h.games.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, games)
}

func (h *Handler) createGame(c echo.Context) error {
	var input services.GameInput
	if err := c.Bind(&input); err != nil {
		return domain.ErrInvalidInput
	}
	game, err := h.games.Create(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusCreated, game)
}

func (h *Handler) updateGame(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var input services.GameInput
	if err := c.Bind(&input); err != nil {
		return domain.ErrInvalidInput
	}
	game, err := h.games.Update(c.Request().Context(), id, input)
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, game)
}

func (h *Handler) deleteGame(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := h.games.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(stdhttp.StatusOK)
}

func (h *Handler) listListings(c echo.Context) error {
	listings, err := h.listings.List(c.Request().Context(), domain.ListingFilters{
		GameID: c.QueryParam("game_id"),
		Role:   c.QueryParam("role"),
		Region: c.QueryParam("region"),
		Mode:   c.QueryParam("mode"),
		Status: c.QueryParam("status"),
		Search: c.QueryParam("search"),
	})
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, listings)
}

func (h *Handler) getListing(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	listing, err := h.listings.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, listing)
}

func (h *Handler) createListing(c echo.Context) error {
	var input services.ListingInput
	if err := c.Bind(&input); err != nil {
		return domain.ErrInvalidInput
	}
	listing, err := h.listings.Create(c.Request().Context(), middleware.UserID(c), input)
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusCreated, listing)
}

func (h *Handler) updateListing(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var input services.ListingInput
	if err := c.Bind(&input); err != nil {
		return domain.ErrInvalidInput
	}
	listing, err := h.listings.Update(c.Request().Context(), middleware.UserID(c), middleware.Role(c), id, input)
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, listing)
}

func (h *Handler) closeListing(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	listing, err := h.listings.Close(c.Request().Context(), middleware.UserID(c), middleware.Role(c), id)
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, listing)
}

func (h *Handler) deleteListing(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := h.listings.Delete(c.Request().Context(), middleware.UserID(c), middleware.Role(c), id); err != nil {
		return err
	}
	return c.NoContent(stdhttp.StatusOK)
}

func (h *Handler) apply(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var input services.ApplicationInput
	if err := c.Bind(&input); err != nil {
		return domain.ErrInvalidInput
	}
	application, err := h.listings.Apply(c.Request().Context(), middleware.UserID(c), id, input)
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusCreated, application)
}

func (h *Handler) outgoing(c echo.Context) error {
	items, err := h.listings.Outgoing(c.Request().Context(), middleware.UserID(c))
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, items)
}

func (h *Handler) incoming(c echo.Context) error {
	items, err := h.listings.Incoming(c.Request().Context(), middleware.UserID(c))
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, items)
}

func (h *Handler) updateApplicationStatus(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&body); err != nil || !validation.ApplicationStatus(body.Status) {
		return domain.ErrInvalidInput
	}
	application, err := h.listings.UpdateApplicationStatus(c.Request().Context(), middleware.UserID(c), middleware.Role(c), id, body.Status)
	if err != nil {
		return err
	}
	return c.JSON(stdhttp.StatusOK, application)
}

func parseID(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, domain.ErrInvalidInput
	}
	return id, nil
}

func errorHandler(err error, c echo.Context) {
	status := stdhttp.StatusInternalServerError
	message := "internal server error"
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		status, message = stdhttp.StatusBadRequest, "invalid request"
	case errors.Is(err, domain.ErrUnauthorized):
		status, message = stdhttp.StatusUnauthorized, "unauthorized"
	case errors.Is(err, domain.ErrForbidden):
		status, message = stdhttp.StatusForbidden, "forbidden"
	case errors.Is(err, domain.ErrNotFound):
		status, message = stdhttp.StatusNotFound, "not found"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrDuplicate), errors.Is(err, domain.ErrOwnListing), errors.Is(err, domain.ErrClosedListing):
		status, message = stdhttp.StatusConflict, err.Error()
	}
	if echoErr, ok := err.(*echo.HTTPError); ok {
		status = echoErr.Code
		message = "request error"
	}
	_ = c.JSON(status, map[string]string{"error": message})
}
