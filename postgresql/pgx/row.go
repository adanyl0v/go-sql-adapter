package pgxadapt

import driver "github.com/adanyl0v/go-sql-adapter/postgresql/pgx/driver"

type Row struct {
	driverRow driver.Row
}

func NewRow(driverRow driver.Row) Row {
	return Row{
		driverRow: driverRow,
	}
}

// Err always returns no errors.
func (r Row) Err() error {
	return nil
}

func (r Row) Scan(dest ...any) error {
	err := r.driverRow.Scan(dest...)
	return err
}
