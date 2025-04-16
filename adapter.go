package adapter

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrUnsupportedLastInsertId = errors.New("unsupported last insert id")
	ErrUnsupportedRowsAffected = errors.New("unsupported rows affected")

	ErrNoRows              = errors.New("no rows in result set")
	ErrTooManyRows         = errors.New("too many rows in result set")
	ErrCheckViolation      = errors.New("violated the check constraint")
	ErrUniqueViolation     = errors.New("violated the unique constraint")
	ErrNotNullViolation    = errors.New("violated the not null constraint")
	ErrForeignKeyViolation = errors.New("violated the foreign key constraint")
)

type Result interface {
	sql.Result
}

type Row interface {
	Err() error
	Scan(dest ...any) error
}

type Rows interface {
	Err() error
	Next() bool
	Close() error
	Scan(dest ...any) error
}

type Stmt interface {
	Exec(args ...any) (Result, error)
	Query(args ...any) (Rows, error)
	QueryRow(args ...any) Row
	Close() error
}

type Conn interface {
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
	Prepare(ctx context.Context, query string) (Stmt, error)
	Begin(ctx context.Context) (Tx, error)
	Ping(ctx context.Context) error
	Close() error
}

type Tx interface {
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
	Prepare(ctx context.Context, query string) (Stmt, error)
	Begin(ctx context.Context) (Tx, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
