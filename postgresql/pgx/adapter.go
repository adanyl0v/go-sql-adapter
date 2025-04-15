package pgxadapt

import (
	"context"
	"errors"
	"time"

	adapter "github.com/adanyl0v/go-sql-adapter"
	"github.com/adanyl0v/go-sql-adapter/postgresql/pgx/driver"
	"github.com/adanyl0v/go-sql-adapter/postgresql/pgx/errs"
	"github.com/adanyl0v/go-sql-adapter/postgresql/pgx/trace"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Result
// ------

type Result struct {
	rowsAffected int64
}

func NewResult(driverResult driver.Result) Result {
	return Result{
		rowsAffected: driverResult.RowsAffected(),
	}
}

// LastInsertId always returns 0 and a non-nil error,
// because PostgreSQL does not return the last insert id.
func (r Result) LastInsertId() (int64, error) {
	return 0, adapter.ErrUnsupportedLastInsertId
}

// RowsAffected always returns nil.
func (r Result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

// Row
// ---

type Row struct {
	driverRow driver.Row
	tracer    trace.Logger
}

func NewRow(driverRow driver.Row, tracer trace.Logger) Row {
	return Row{
		driverRow: driverRow,
		tracer:    tracer,
	}
}

// Err always returns nil.
func (r Row) Err() error {
	return nil
}

func (r Row) Scan(dest ...any) error {
	err := r.driverRow.Scan(dest...)
	if err != nil {
		r.tracer.Log(trace.ErrorLevel, "failed to scan a row", map[string]any{
			trace.ErrorKey: err,
		})
		return err
	}

	r.tracer.Log(trace.TraceLevel, "scanned a row", nil)
	return nil
}

// Rows
// ----

type Rows struct {
	driverRows driver.Rows
	tracer     trace.Logger
}

func NewRows(driverRows driver.Rows, tracer trace.Logger) Rows {
	return Rows{
		driverRows: driverRows,
		tracer:     tracer,
	}
}

func (r Rows) Err() error {
	return r.driverRows.Err()
}

func (r Rows) Next() bool {
	return r.driverRows.Next()
}

// Close always returns nil.
func (r Rows) Close() error {
	r.driverRows.Close()
	return nil
}

func (r Rows) Scan(dest ...any) error {
	err := r.driverRows.Scan(dest...)
	if err != nil {
		r.tracer.Log(trace.ErrorLevel, "failed to scan a row", map[string]any{
			trace.ErrorKey: err,
		})
		return err
	}

	r.tracer.Log(trace.TraceLevel, "scanned a row", nil)
	return nil
}

// Stmt
// ----

type StmtConn interface {
	driver.Execer
	driver.Querier
	driver.RowQuerier
}

type Stmt struct {
	conn   StmtConn
	tracer trace.Logger
	ctx    context.Context
	query  string
}

func NewStmt(
	conn StmtConn,
	tracer trace.Logger,
	ctx context.Context,
	query string,
) Stmt {
	return Stmt{
		conn:   conn,
		tracer: tracer,
		ctx:    ctx,
		query:  query,
	}
}

func (s Stmt) Exec(args ...any) (adapter.Result, error) {
	return runExec(s.conn, s.tracer, s.ctx, s.query, args...)
}

func (s Stmt) Query(args ...any) (adapter.Rows, error) {
	return runQuery(s.conn, s.tracer, s.ctx, s.query, args...)
}

func (s Stmt) QueryRow(args ...any) adapter.Row {
	return runQueryRow(s.conn, s.tracer, s.ctx, s.query, args...)
}

// Close does nothing and always returns nil.
func (s Stmt) Close() error {
	return nil
}

// Conn
// ----

type Conn struct {
	driverConn driver.Conn
	tracer     trace.Logger
}

func NewConn(driverConn driver.Conn, tracer trace.Logger) Conn {
	return Conn{
		driverConn: driverConn,
		tracer:     tracer,
	}
}

func (c Conn) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (adapter.Result, error) {
	return runExec(c.driverConn, c.tracer, ctx, query, args...)
}

func (c Conn) Query(
	ctx context.Context,
	query string,
	args ...any,
) (adapter.Rows, error) {
	return runQuery(c.driverConn, c.tracer, ctx, query, args...)
}

func (c Conn) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) adapter.Row {
	return runQueryRow(c.driverConn, c.tracer, ctx, query, args...)
}

func (c Conn) Prepare(ctx context.Context, query string) (adapter.Stmt, error) {
	return runPrepare(c.driverConn, c.tracer, ctx, query)
}

func (c Conn) Begin(ctx context.Context) (adapter.Tx, error) {
	return runBegin(c.driverConn, c.tracer, ctx)
}

func (c Conn) Ping(ctx context.Context) error {
	err := c.driverConn.Ping(ctx)
	if err != nil {
		c.tracer.Log(
			trace.ErrorLevel,
			"failed to ping the connection",
			map[string]any{
				trace.ErrorKey: err,
			},
		)
		return err
	}

	c.tracer.Log(trace.TraceLevel, "pinged the connection", nil)
	return nil
}

// Close always returns nil.
func (c Conn) Close() error {
	c.driverConn.Close()
	return nil
}

// Tx
// --

type Tx struct {
	driverTx driver.Tx
	tracer   trace.Logger
}

func NewTx(driverTx driver.Tx, tracer trace.Logger) Tx {
	return Tx{
		driverTx: driverTx,
		tracer:   tracer,
	}
}

func (t Tx) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (adapter.Result, error) {
	return runExec(t.driverTx, t.tracer, ctx, query, args...)
}

func (t Tx) Query(
	ctx context.Context,
	query string,
	args ...any,
) (adapter.Rows, error) {
	return runQuery(t.driverTx, t.tracer, ctx, query, args...)
}

func (t Tx) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) adapter.Row {
	return runQueryRow(t.driverTx, t.tracer, ctx, query, args...)
}

func (t Tx) Prepare(ctx context.Context, query string) (adapter.Stmt, error) {
	return runPrepare(t.driverTx, t.tracer, ctx, query)
}

func (t Tx) Begin(ctx context.Context) (adapter.Tx, error) {
	return runBegin(t.driverTx, t.tracer, ctx)
}

func (t Tx) Commit(ctx context.Context) error {
	err := t.driverTx.Commit(ctx)
	if err != nil {
		t.tracer.Log(
			trace.ErrorLevel,
			"failed to commit a transaction",
			map[string]any{
				trace.ErrorKey: err,
			},
		)
		return err
	}

	t.tracer.Log(trace.TraceLevel, "committed a transaction", nil)
	return nil
}

func (t Tx) Rollback(ctx context.Context) error {
	err := t.driverTx.Rollback(ctx)
	if err != nil {
		t.tracer.Log(
			trace.ErrorLevel,
			"failed to rollback a transaction",
			map[string]any{
				trace.ErrorKey: err,
			},
		)
		return err
	}

	t.tracer.Log(trace.TraceLevel, "rolled back a transaction", nil)
	return nil
}

// Helpers
// -------

func runExec(
	execer driver.Execer,
	tracer trace.Logger,
	ctx context.Context,
	query string,
	args ...any,
) (adapter.Result, error) {

	tracer = tracer.With(map[string]any{
		trace.QueryKey: query,
	})

	start := time.Now()
	driverResult, err := execer.Exec(ctx, query, args...)
	dur := time.Since(start)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.CheckViolation:
				err = errs.New(adapter.ErrCheckViolation.Error(), err)
			case pgerrcode.UniqueViolation:
				err = errs.New(adapter.ErrUniqueViolation.Error(), err)
			case pgerrcode.NotNullViolation:
				err = errs.New(adapter.ErrNotNullViolation.Error(), err)
			case pgerrcode.ForeignKeyViolation:
				err = errs.New(adapter.ErrForeignKeyViolation.Error(), err)
			}
		}

		tracer.Log(trace.ErrorLevel, "failed to execute", map[string]any{
			trace.ErrorKey: err,
		})
		return nil, err
	}

	result := NewResult(driverResult)
	tracer.Log(trace.ErrorLevel, "executed", map[string]any{
		trace.ResultKey:   result,
		trace.DurationKey: dur,
	})

	return result, nil
}

func runQuery(
	querier driver.Querier,
	tracer trace.Logger,
	ctx context.Context,
	query string,
	args ...any,
) (adapter.Rows, error) {

	tracer = tracer.With(map[string]any{
		trace.QueryKey: query,
	})

	start := time.Now()
	//nolint:rowserrcheck,sqlclosecheck
	driverRows, err := querier.Query(ctx, query, args...)
	dur := time.Since(start)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = errs.New(adapter.ErrNoRows.Error(), err)
		}

		tracer.Log(trace.ErrorLevel, "failed to execute", map[string]any{
			trace.ErrorKey: err,
		})
		return nil, err
	}

	tracer.Log(trace.ErrorLevel, "executed", map[string]any{
		trace.DurationKey: dur,
	})

	rows := NewRows(driverRows, tracer)
	return rows, nil
}

func runQueryRow(
	rowQuerier driver.RowQuerier,
	tracer trace.Logger,
	ctx context.Context,
	query string,
	args ...any,
) adapter.Row {

	start := time.Now()
	//nolint:rowserrcheck,sqlclosecheck
	driverRow := rowQuerier.QueryRow(ctx, query, args...)
	dur := time.Since(start)

	tracer.Log(trace.ErrorLevel, "executed", map[string]any{
		trace.QueryKey:    query,
		trace.DurationKey: dur,
	})

	row := NewRow(driverRow, tracer)
	return row
}

// runPrepare always returns no errors.
func runPrepare(
	conn StmtConn,
	tracer trace.Logger,
	ctx context.Context,
	query string,
) (adapter.Stmt, error) {

	tracer.Log(trace.TraceLevel, "prepared a statement", map[string]any{
		trace.QueryKey: query,
	})

	stmt := NewStmt(conn, tracer, ctx, query)
	return stmt, nil
}

func runBegin(
	beginner driver.Beginner,
	tracer trace.Logger,
	ctx context.Context,
) (adapter.Tx, error) {

	driverTx, err := beginner.Begin(ctx)
	if err != nil {
		tracer.Log(
			trace.ErrorLevel,
			"failed to begin a transaction",
			map[string]any{
				trace.ErrorKey: err,
			},
		)
		return nil, err
	}

	tracer.Log(trace.TraceLevel, "began a transaction", nil)

	tx := NewTx(driverTx, tracer)
	return tx, nil
}
