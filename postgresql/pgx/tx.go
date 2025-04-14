package pgxadapt

import (
	"context"

	sqladapt "github.com/adanyl0v/go-sql-adapter"
	driver "github.com/adanyl0v/go-sql-adapter/postgresql/pgx/driver"
)

type Tx struct {
	driverTx driver.Tx
}

func NewTx(driverTx driver.Tx) Tx {
	return Tx{
		driverTx: driverTx,
	}
}

func (t Tx) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (sqladapt.Result, error) {

	driverResult, err := t.driverTx.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	result := NewResult(driverResult)
	return result, nil
}

func (t Tx) Query(
	ctx context.Context,
	query string,
	args ...any,
) (sqladapt.Rows, error) {

	//nolint:rowserrcheck,sqlclosecheck
	driverRows, err := t.driverTx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	rows := NewRows(driverRows)
	return rows, nil
}

func (t Tx) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) sqladapt.Row {
	driverRow := t.driverTx.QueryRow(ctx, query, args...)
	row := NewRow(driverRow)
	return row
}

// Prepare always returns no errors.
func (t Tx) Prepare(ctx context.Context, query string) (sqladapt.Stmt, error) {
	stmt := NewStmt(t.driverTx, ctx, query)
	return stmt, nil
}

func (t Tx) Begin(ctx context.Context) (sqladapt.Tx, error) {
	driverTx, err := t.driverTx.Begin(ctx)
	if err != nil {
		return nil, err
	}

	tx := NewTx(driverTx)
	return tx, nil
}

func (t Tx) Commit(ctx context.Context) error {
	err := t.driverTx.Commit(ctx)
	return err
}

func (t Tx) Rollback(ctx context.Context) error {
	err := t.driverTx.Rollback(ctx)
	return err
}
