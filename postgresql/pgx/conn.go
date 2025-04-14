package pgxadapt

import (
	"context"

	sqladapt "github.com/adanyl0v/go-sql-adapter"
	driver "github.com/adanyl0v/go-sql-adapter/postgresql/pgx/driver"
)

type Conn struct {
	driverConn driver.Conn
}

func NewConn(driverConn driver.Conn) Conn {
	return Conn{
		driverConn: driverConn,
	}
}

func (c Conn) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (sqladapt.Result, error) {

	driverResult, err := c.driverConn.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	result := NewResult(driverResult)
	return result, nil
}

func (c Conn) Query(
	ctx context.Context,
	query string,
	args ...any,
) (sqladapt.Rows, error) {

	//nolint:rowserrcheck,sqlclosecheck
	driverRows, err := c.driverConn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	rows := NewRows(driverRows)
	return rows, nil
}

func (c Conn) QueryRow(ctx context.Context, query string, args ...any) sqladapt.Row {
	driverRow := c.driverConn.QueryRow(ctx, query, args...)
	row := NewRow(driverRow)
	return row
}

// Prepare always returns no errors.
func (c Conn) Prepare(ctx context.Context, query string) (sqladapt.Stmt, error) {
	stmt := NewStmt(c.driverConn, ctx, query)
	return stmt, nil
}

func (c Conn) Begin(ctx context.Context) (sqladapt.Tx, error) {
	driverTx, err := c.driverConn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	tx := NewTx(driverTx)
	return tx, nil
}

func (c Conn) Ping(ctx context.Context) error {
	err := c.driverConn.Ping(ctx)
	return err
}

// Close always returns no errors.
func (c Conn) Close() error {
	c.driverConn.Close()
	return nil
}
