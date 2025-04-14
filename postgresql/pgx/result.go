package pgxadapt

import (
	sqladapt "github.com/adanyl0v/go-sql-adapter"
	driver "github.com/adanyl0v/go-sql-adapter/postgresql/pgx/driver"
)

type Result struct {
	rowsAffected int64
}

func NewResult(driverResult driver.Result) Result {
	return Result{
		rowsAffected: driverResult.RowsAffected(),
	}
}

// LastInsertId always returns 0, because PostgreSQL
// does not return the last insert id.
func (r Result) LastInsertId() (int64, error) {
	return 0, sqladapt.ErrUnsupportedLastInsertId
}

// RowsAffected always returns no errors.
func (r Result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}
