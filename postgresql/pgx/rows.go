package pgxadapt

import driver "github.com/adanyl0v/go-sql-adapter/postgresql/pgx/driver"

type Rows struct {
	driverRows driver.Rows
}

func NewRows(driverRows driver.Rows) Rows {
	return Rows{
		driverRows: driverRows,
	}
}

func (r Rows) Err() error {
	return r.driverRows.Err()
}

func (r Rows) Next() bool {
	return r.driverRows.Next()
}

// Close always returns no errors.
func (r Rows) Close() error {
	r.driverRows.Close()
	return nil
}

func (r Rows) Scan(dest ...any) error {
	err := r.driverRows.Scan(dest...)
	return err
}
