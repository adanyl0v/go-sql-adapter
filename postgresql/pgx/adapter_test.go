package pgxadapt

import (
	"context"
	"errors"
	"testing"

	adapter "github.com/adanyl0v/go-sql-adapter"
	"github.com/adanyl0v/go-sql-adapter/postgresql/pgx/driver"
	mock_driver "github.com/adanyl0v/go-sql-adapter/postgresql/pgx/driver/mock"
	"github.com/adanyl0v/go-sql-adapter/postgresql/pgx/trace"
	mock_trace "github.com/adanyl0v/go-sql-adapter/postgresql/pgx/trace/mock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Result
// ------

func TestResult_LastInsertId(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockResult := mock_driver.NewMockResult(ctrl)
	mockResult.
		EXPECT().
		RowsAffected().
		Return(0)

	result := NewResult(mockResult)

	id, err := result.LastInsertId()
	require.EqualValues(t, id, 0)
	require.EqualError(t, err, adapter.ErrUnsupportedLastInsertId.Error())
}

func TestResult_RowsAffected(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	const rowsAffected = 10

	mockResult := mock_driver.NewMockResult(ctrl)
	mockResult.
		EXPECT().
		RowsAffected().
		Return(rowsAffected)

	result := NewResult(mockResult)

	affected, err := result.RowsAffected()
	require.EqualValues(t, affected, rowsAffected)
	require.NoError(t, err)
}

// Row
// ---

func TestRow_Err(t *testing.T) {
	t.Parallel()

	row := NewRow(nil, nil)
	require.NoError(t, row.Err())
}

func TestRow_Scan(t *testing.T) {
	t.Parallel()

	type commandType func(row *Row) error

	command := func(row *Row) error {
		return row.Scan(nil)
	}

	testCases := map[string]struct {
		Expect  func(mockRow *mock_driver.MockRow, mockTracer *mock_trace.MockLogger)
		Command commandType
		Check   func(err error)
	}{
		"success": {
			Expect: func(
				mockRow *mock_driver.MockRow,
				mockTracer *mock_trace.MockLogger,
			) {
				mockRow.
					EXPECT().
					Scan(nil).Return(nil)

				mockTracer.
					EXPECT().
					Log(trace.TraceLevel, "scanned a row", nil)
			},
			Command: command,
			Check: func(err error) {
				require.NoError(t, err)
			},
		},
		"failure": {
			Expect: func(mockRow *mock_driver.MockRow, mockTracer *mock_trace.MockLogger) {
				mockRow.
					EXPECT().
					Scan(nil).
					Return(errors.New(""))

				mockTracer.
					EXPECT().
					Log(trace.ErrorLevel, "failed to scan a row", gomock.Any())
			},
			Command: command,
			Check: func(err error) {
				require.Error(t, err)
			},
		},
		"no_rows": {
			Expect: func(mockRow *mock_driver.MockRow, mockTracer *mock_trace.MockLogger) {
				mockRow.
					EXPECT().
					Scan(nil).
					Return(pgx.ErrNoRows)

				mockTracer.
					EXPECT().
					Log(trace.ErrorLevel, "failed to scan a row", gomock.Any())
			},
			Command: command,
			Check: func(err error) {
				require.EqualError(t, err, adapter.ErrNoRows.Error())
			},
		},
		"too_many_rows": {
			Expect: func(mockRow *mock_driver.MockRow, mockTracer *mock_trace.MockLogger) {
				mockRow.
					EXPECT().
					Scan(nil).
					Return(pgx.ErrTooManyRows)

				mockTracer.
					EXPECT().
					Log(trace.ErrorLevel, "failed to scan a row", gomock.Any())
			},
			Command: command,
			Check: func(err error) {
				require.EqualError(t, err, adapter.ErrTooManyRows.Error())
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockRow := mock_driver.NewMockRow(ctrl)
			mockTracer := mock_trace.NewMockLogger(ctrl)

			testCase.Expect(mockRow, mockTracer)

			row := NewRow(mockRow, mockTracer)

			err := testCase.Command(&row)
			testCase.Check(err)
		})
	}
}

// Rows
// ----

func TestRows_Err(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	expectedErr := errors.New("error")

	mockRows := mock_driver.NewMockRows(ctrl)
	mockRows.
		EXPECT().
		Err().
		Return(expectedErr)

	rows := NewRows(mockRows, nil)

	err := rows.Err()
	require.EqualError(t, err, expectedErr.Error())
}

func TestRows_Next(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	expectedNext := true

	mockRows := mock_driver.NewMockRows(ctrl)
	mockRows.
		EXPECT().
		Next().
		Return(expectedNext)

	rows := NewRows(mockRows, nil)

	next := rows.Next()
	require.EqualValues(t, next, expectedNext)
}

func TestRows_Close(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockRows := mock_driver.NewMockRows(ctrl)
	mockRows.
		EXPECT().
		Close()

	rows := NewRows(mockRows, nil)

	err := rows.Close()
	require.NoError(t, err)
}

func TestRows_Scan(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRows := mock_driver.NewMockRows(ctrl)
		mockRows.
			EXPECT().
			Scan(nil).Return(nil)

		mockTracer := mock_trace.NewMockLogger(ctrl)
		mockTracer.
			EXPECT().
			Log(trace.TraceLevel, "scanned a row", nil)

		rows := NewRows(mockRows, mockTracer)

		err := rows.Scan(nil)
		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRows := mock_driver.NewMockRows(ctrl)
		mockRows.
			EXPECT().
			Scan(nil).
			Return(errors.New(""))

		mockTracer := mock_trace.NewMockLogger(ctrl)
		mockTracer.
			EXPECT().
			Log(trace.ErrorLevel, "failed to scan a row", gomock.Any())

		rows := NewRows(mockRows, mockTracer)

		err := rows.Scan(nil)
		require.Error(t, err)
	})

	t.Run("no_rows", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRows := mock_driver.NewMockRows(ctrl)
		mockRows.
			EXPECT().
			Scan(nil).
			Return(pgx.ErrNoRows)

		mockTracer := mock_trace.NewMockLogger(ctrl)
		mockTracer.
			EXPECT().
			Log(trace.ErrorLevel, "failed to scan a row", gomock.Any())

		rows := NewRows(mockRows, mockTracer)

		err := rows.Scan(nil)
		require.EqualError(t, err, adapter.ErrNoRows.Error())
	})
}

// Stmt
// ----

func TestStmt_Exec(t *testing.T) {
}

func TestStmt_Query(t *testing.T) {
}

func TestStmt_QueryRow(t *testing.T) {
}

func TestStmt_Close(t *testing.T) {
	t.Parallel()

	stmt := NewStmt(nil, nil, nil, "")
	require.NoError(t, stmt.Close())
}

// Conn
// ----

func TestConn_Exec(t *testing.T) {
}

func TestConn_Query(t *testing.T) {
}

func TestConn_QueryRow(t *testing.T) {
}

func TestConn_Prepare(t *testing.T) {
}

func TestConn_Begin(t *testing.T) {
}

func TestConn_Ping(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockConn := mock_driver.NewMockConn(ctrl)
		mockConn.
			EXPECT().
			Ping(gomock.Any()).
			Return(nil)

		mockTracer := mock_trace.NewMockLogger(ctrl)
		mockTracer.
			EXPECT().
			Log(trace.TraceLevel, "pinged the connection", nil)

		conn := NewConn(mockConn, mockTracer)
		err := conn.Ping(context.Background())
		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockConn := mock_driver.NewMockConn(ctrl)
		mockConn.
			EXPECT().
			Ping(gomock.Any()).
			Return(errors.New(""))

		mockTracer := mock_trace.NewMockLogger(ctrl)
		mockTracer.
			EXPECT().
			Log(trace.ErrorLevel, "failed to ping the connection", gomock.Any())

		conn := NewConn(mockConn, mockTracer)
		err := conn.Ping(context.Background())
		require.Error(t, err)
	})
}

// Tx
// --

func TestTx_Exec(t *testing.T) {
}

func TestTx_Query(t *testing.T) {
}

func TestTx_QueryRow(t *testing.T) {
}

func TestTx_Prepare(t *testing.T) {
}

func TestTx_Begin(t *testing.T) {
}

func TestTx_Commit(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockTx := mock_driver.NewMockTx(ctrl)
		mockTx.
			EXPECT().
			Commit(gomock.Any()).
			Return(nil)

		mockTracer := mock_trace.NewMockLogger(ctrl)
		mockTracer.
			EXPECT().
			Log(trace.TraceLevel, "committed a transaction", gomock.Any())

		tx := NewTx(mockTx, mockTracer)
		err := tx.Commit(context.Background())
		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockTx := mock_driver.NewMockTx(ctrl)
		mockTx.
			EXPECT().
			Commit(gomock.Any()).
			Return(errors.New(""))

		mockTracer := mock_trace.NewMockLogger(ctrl)
		mockTracer.
			EXPECT().
			Log(trace.ErrorLevel, "failed to commit a transaction", gomock.Any())

		tx := NewTx(mockTx, mockTracer)
		err := tx.Commit(context.Background())
		require.Error(t, err)
	})
}

func TestTx_Rollback(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockTx := mock_driver.NewMockTx(ctrl)
		mockTx.
			EXPECT().
			Rollback(gomock.Any()).
			Return(nil)

		mockTracer := mock_trace.NewMockLogger(ctrl)
		mockTracer.
			EXPECT().
			Log(trace.TraceLevel, "rolled back a transaction", gomock.Any())

		tx := NewTx(mockTx, mockTracer)
		err := tx.Rollback(context.Background())
		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockTx := mock_driver.NewMockTx(ctrl)
		mockTx.
			EXPECT().
			Rollback(gomock.Any()).
			Return(errors.New(""))

		mockTracer := mock_trace.NewMockLogger(ctrl)
		mockTracer.
			EXPECT().
			Log(trace.ErrorLevel, "failed to rollback a transaction", gomock.Any())

		tx := NewTx(mockTx, mockTracer)
		err := tx.Rollback(context.Background())
		require.Error(t, err)
	})
}

// Helpers
// -------

// ATTENTION!!!
// This is the creepiest part of the package...

func TestRunExec(t *testing.T) {
	t.Parallel()

	type commandType func(
		execer driver.Execer,
		tracer trace.Logger,
		runExecFn func(
			execer driver.Execer,
			tracer trace.Logger,
			ctx context.Context,
			query string,
			args ...any,
		) (adapter.Result, error),
	) (adapter.Result, error)

	command := func(
		execer driver.Execer,
		tracer trace.Logger,
		runExecFn func(
			execer driver.Execer,
			tracer trace.Logger,
			ctx context.Context,
			query string,
			args ...any,
		) (adapter.Result, error),
	) (adapter.Result, error) {
		return runExecFn(execer, tracer, context.Background(), "")
	}

	testCases := map[string]struct {
		Expect func(
			ctrl *gomock.Controller,
			mockExecer *mock_driver.MockExecer,
			mockTracer *mock_trace.MockLogger,
		)
		Command commandType
		Check   func(result adapter.Result, err error)
	}{
		"success": {
			Expect: func(
				ctrl *gomock.Controller,
				mockExecer *mock_driver.MockExecer,
				mockTracer *mock_trace.MockLogger,
			) {
				mockExecer.
					EXPECT().
					Exec(gomock.Any(), "").
					Return(pgconn.NewCommandTag(""), nil)

				mockTracer.
					EXPECT().
					Log(trace.TraceLevel, "executed", gomock.Any())
			},
			Command: command,
			Check: func(_ adapter.Result, err error) {
				require.NoError(t, err)
			},
		},
		"failure": {
			Expect: func(
				ctrl *gomock.Controller,
				mockExecer *mock_driver.MockExecer,
				mockTracer *mock_trace.MockLogger,
			) {
				mockExecer.
					EXPECT().
					Exec(gomock.Any(), "").
					Return(pgconn.CommandTag{}, errors.New(""))

				mockTracer.
					EXPECT().
					Log(trace.ErrorLevel, "failed to execute", gomock.Any())
			},
			Command: command,
			Check: func(_ adapter.Result, err error) {
				require.Error(t, err)
			},
		},
		"check_violation": {
			Expect: func(
				ctrl *gomock.Controller,
				mockExecer *mock_driver.MockExecer,
				mockTracer *mock_trace.MockLogger,
			) {
				mockExecer.
					EXPECT().
					Exec(gomock.Any(), "").
					Return(pgconn.CommandTag{}, &pgconn.PgError{
						Code: pgerrcode.CheckViolation,
					})

				mockTracer.
					EXPECT().
					Log(trace.ErrorLevel, "failed to execute", gomock.Any())
			},
			Command: command,
			Check: func(_ adapter.Result, err error) {
				require.EqualError(t, err, adapter.ErrCheckViolation.Error())
			},
		},
		"unique_violation": {
			Expect: func(
				ctrl *gomock.Controller,
				mockExecer *mock_driver.MockExecer,
				mockTracer *mock_trace.MockLogger,
			) {
				mockExecer.
					EXPECT().
					Exec(gomock.Any(), "").
					Return(pgconn.CommandTag{}, &pgconn.PgError{
						Code: pgerrcode.UniqueViolation,
					})

				mockTracer.
					EXPECT().
					Log(trace.ErrorLevel, "failed to execute", gomock.Any())
			},
			Command: command,
			Check: func(_ adapter.Result, err error) {
				require.EqualError(t, err, adapter.ErrUniqueViolation.Error())
			},
		},
		"not_null_violation": {
			Expect: func(
				ctrl *gomock.Controller,
				mockExecer *mock_driver.MockExecer,
				mockTracer *mock_trace.MockLogger,
			) {
				mockExecer.
					EXPECT().
					Exec(gomock.Any(), "").
					Return(pgconn.CommandTag{}, &pgconn.PgError{
						Code: pgerrcode.NotNullViolation,
					})

				mockTracer.
					EXPECT().
					Log(trace.ErrorLevel, "failed to execute", gomock.Any())
			},
			Command: command,
			Check: func(_ adapter.Result, err error) {
				require.EqualError(t, err, adapter.ErrNotNullViolation.Error())
			},
		},
		"foreign_key_violation": {
			Expect: func(
				ctrl *gomock.Controller,
				mockExecer *mock_driver.MockExecer,
				mockTracer *mock_trace.MockLogger,
			) {
				mockExecer.
					EXPECT().
					Exec(gomock.Any(), "").
					Return(pgconn.CommandTag{}, &pgconn.PgError{
						Code: pgerrcode.ForeignKeyViolation,
					})

				mockTracer.
					EXPECT().
					Log(trace.ErrorLevel, "failed to execute", gomock.Any())
			},
			Command: command,
			Check: func(_ adapter.Result, err error) {
				require.EqualError(
					t,
					err,
					adapter.ErrForeignKeyViolation.Error(),
				)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockExecer := mock_driver.NewMockExecer(ctrl)

			mockTracer := mock_trace.NewMockLogger(ctrl)
			mockTracer.
				EXPECT().
				WithCallerSkip(gomock.Any()).
				Return(mockTracer)
			mockTracer.
				EXPECT().
				With(gomock.Any()).
				Return(mockTracer)

			testCase.Expect(ctrl, mockExecer, mockTracer)
			result, err := testCase.Command(mockExecer, mockTracer, runExec)
			testCase.Check(result, err)
		})
	}
}

func TestRunQuery(t *testing.T) {
	t.Parallel()

	type commandType func(
		querier driver.Querier,
		tracer trace.Logger,
		runExecFn func(
			querier driver.Querier,
			tracer trace.Logger,
			ctx context.Context,
			query string,
			args ...any,
		) (rows adapter.Rows, err error),
	) (rows adapter.Rows, err error)

	command := func(
		querier driver.Querier,
		tracer trace.Logger,
		runQueryFn func(
			querier driver.Querier,
			tracer trace.Logger,
			ctx context.Context,
			query string,
			args ...any,
		) (rows adapter.Rows, err error),
	) (rows adapter.Rows, err error) {
		return runQueryFn(querier, tracer, context.Background(), "")
	}

	testCases := map[string]struct {
		Expect func(
			ctrl *gomock.Controller,
			mockQuerier *mock_driver.MockQuerier,
			mockTracer *mock_trace.MockLogger,
		)
		Command commandType
		Check   func(rows adapter.Rows, err error)
	}{
		"success": {
			Expect: func(
				ctrl *gomock.Controller,
				mockQuerier *mock_driver.MockQuerier,
				mockTracer *mock_trace.MockLogger,
			) {
				mockQuerier.
					EXPECT().
					Query(gomock.Any(), "").
					Return(nil, nil)

				mockTracer.
					EXPECT().
					Log(trace.TraceLevel, "executed", gomock.Any())
			},
			Command: command,
			Check: func(_ adapter.Rows, err error) {
				require.NoError(t, err)
			},
		},
		"failure": {
			Expect: func(
				ctrl *gomock.Controller,
				mockQuerier *mock_driver.MockQuerier,
				mockTracer *mock_trace.MockLogger,
			) {
				mockQuerier.
					EXPECT().
					Query(gomock.Any(), "").
					Return(nil, errors.New(""))

				mockTracer.
					EXPECT().
					Log(trace.ErrorLevel, "failed to execute", gomock.Any())
			},
			Command: command,
			Check: func(_ adapter.Rows, err error) {
				require.Error(t, err)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockQuerier := mock_driver.NewMockQuerier(ctrl)

			mockTracer := mock_trace.NewMockLogger(ctrl)
			mockTracer.
				EXPECT().
				WithCallerSkip(gomock.Any()).
				Return(mockTracer)
			mockTracer.
				EXPECT().
				With(gomock.Any()).
				Return(mockTracer)

			testCase.Expect(ctrl, mockQuerier, mockTracer)
			rows, err := testCase.Command(mockQuerier, mockTracer, runQuery)
			testCase.Check(rows, err)
		})
	}
}

func TestRunQueryRow(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockRowQuerier := mock_driver.NewMockRowQuerier(ctrl)

	mockTracer := mock_trace.NewMockLogger(ctrl)
	mockTracer.
		EXPECT().
		WithCallerSkip(gomock.Any()).
		Return(mockTracer)
	mockTracer.
		EXPECT().
		Log(trace.TraceLevel, "executed", gomock.Any())

	mockRowQuerier.
		EXPECT().
		QueryRow(gomock.Any(), "")

	_ = runQueryRow(mockRowQuerier, mockTracer, context.Background(), "")
}

func TestRunPrepare(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockConn := mock_driver.NewMockConn(ctrl)

	mockTracer := mock_trace.NewMockLogger(ctrl)
	mockTracer.
		EXPECT().
		WithCallerSkip(gomock.Any()).
		Return(mockTracer)
	mockTracer.
		EXPECT().
		Log(trace.TraceLevel, "prepared a statement", gomock.Any())

	_, err := runPrepare(mockConn, mockTracer, context.Background(), "")
	require.NoError(t, err)
}

func TestRunBegin(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockBeginner := mock_driver.NewMockBeginner(ctrl)
		mockBeginner.
			EXPECT().
			Begin(gomock.Any()).
			Return(nil, nil)

		mockTracer := mock_trace.NewMockLogger(ctrl)
		mockTracer.
			EXPECT().
			WithCallerSkip(gomock.Any()).
			Return(mockTracer)
		mockTracer.
			EXPECT().
			Log(trace.TraceLevel, "began a transaction", gomock.Any())

		_, err := runBegin(mockBeginner, mockTracer, context.Background())
		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockBeginner := mock_driver.NewMockBeginner(ctrl)
		mockBeginner.
			EXPECT().
			Begin(gomock.Any()).
			Return(nil, errors.New(""))

		mockTracer := mock_trace.NewMockLogger(ctrl)
		mockTracer.
			EXPECT().
			WithCallerSkip(gomock.Any()).
			Return(mockTracer)
		mockTracer.
			EXPECT().
			Log(trace.ErrorLevel, "failed to begin a transaction", gomock.Any())

		_, err := runBegin(mockBeginner, mockTracer, context.Background())
		require.Error(t, err)
	})
}
