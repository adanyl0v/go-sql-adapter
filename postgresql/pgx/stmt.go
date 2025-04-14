package pgxadapt

import (
	"context"

	sqladapt "github.com/adanyl0v/go-sql-adapter"
	driver "github.com/adanyl0v/go-sql-adapter/postgresql/pgx/driver"
)

type Stmt struct {
	driverRawConn driver.RawConn
	ctx           context.Context
	query         string
}

func NewStmt(driverRawConn driver.RawConn, ctx context.Context, query string) Stmt {
	return Stmt{
		driverRawConn: driverRawConn,
		ctx:           ctx,
		query:         query,
	}
}

func (s Stmt) Exec(args ...any) (sqladapt.Result, error) {
	driverResult, err := s.driverRawConn.Exec(s.ctx, s.query, args...)
	if err != nil {
		return nil, err
	}

	result := NewResult(driverResult)
	return result, nil
}

func (s Stmt) Query(args ...any) (sqladapt.Rows, error) {
	//nolint:rowserrcheck,sqlclosecheck
	driverRows, err := s.driverRawConn.Query(s.ctx, s.query, args...)
	if err != nil {
		return nil, err
	}

	rows := NewRows(driverRows)
	return rows, nil
}

func (s Stmt) QueryRow(args ...any) sqladapt.Row {
	driverRow := s.driverRawConn.QueryRow(s.ctx, s.query, args...)
	row := NewRow(driverRow)
	return row
}

// Close does nothing and always returns no errors.
func (s Stmt) Close() error {
	return nil
}
