package postgres

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"subservice/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxClient interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

func New(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.Username, url.QueryEscape(cfg.Password), cfg.Host, cfg.Port, cfg.Name)

	attemptsCount := cfg.MaxConnectionAttempts

	var pool *pgxpool.Pool
	var err error

	for attemptsCount > 0 {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		pool, err = pgxpool.New(ctx, connString)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				return pool, nil
			}
		}

		time.Sleep(cfg.RetryDelay)
		attemptsCount--
	}

	return nil, fmt.Errorf("failed to connect to postgres after %d attemps: %w", cfg.MaxConnectionAttempts, err)
}
