package driver

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Result interface {
	RowsAffected() int64
}

type Row interface {
	Scan(dest ...any) error
}

type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
}

type Execer interface {
	Exec(
		ctx context.Context,
		query string,
		args ...any,
	) (pgconn.CommandTag, error)
}

type Querier interface {
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
}

type RowQuerier interface {
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
}

type Beginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Conn interface {
	Execer
	Querier
	RowQuerier
	Beginner
	Ping(ctx context.Context) error
	Close()
}

type Tx interface {
	Execer
	Querier
	RowQuerier
	Beginner
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
