package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"team-finder/backend/internal/config"
	apphttp "team-finder/backend/internal/handlers/http"
	"team-finder/backend/internal/repositories/postgres"
	"team-finder/backend/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	pool, err := connectPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
	}
	defer pool.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("redis is unavailable, continuing without cache: %v", err)
	}
	defer redisClient.Close()

	userRepo := postgres.NewUserRepository(pool)
	profileRepo := postgres.NewProfileRepository(pool)
	gameRepo := postgres.NewGameRepository(pool)
	listingRepo := postgres.NewListingRepository(pool)
	applicationRepo := postgres.NewApplicationRepository(pool)

	authService := services.NewAuthService(userRepo, profileRepo, cfg.JWTSecret, cfg.TokenTTL)
	gameService := services.NewGameService(gameRepo)
	listingService := services.NewListingService(listingRepo, applicationRepo)

	e := echo.New()
	e.HideBanner = true
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: strings.Split(cfg.AllowOrigins, ","),
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	apphttp.New(authService, gameService, listingService).RegisterRoutes(e, cfg.JWTSecret)

	log.Printf("Team Finder API listening on :%s", cfg.ServerPort)
	if err := e.Start(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}

func connectPostgres(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	var lastErr error
	for attempt := 0; attempt < 20; attempt++ {
		pool, err := pgxpool.New(ctx, databaseURL)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				return pool, nil
			} else {
				lastErr = pingErr
			}
			pool.Close()
		} else {
			lastErr = err
		}
		time.Sleep(time.Second)
	}
	return nil, lastErr
}

func runMigrations(databaseURL string) error {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, "migrations")
}
