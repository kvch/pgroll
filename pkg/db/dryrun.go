// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"
	"database/sql"
)

// DryRunDB is a database connection wrapper that logs SQL statements
// instead of executing them when in dry-run mode.
type DryRunDB struct {
	DB
	logger func(stmt string, args []any)
}

// NewDryRunDB creates a new dry-run database wrapper
func NewDryRunDB(db DB, logger func(stmt string, args []any)) DB {
	return &DryRunDB{
		DB:     db,
		logger: logger,
	}
}

// ExecContext logs the SQL statement instead of executing it
func (d *DryRunDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	d.logger(query, args)
	return &dryRunResult{}, nil
}

// QueryContext executes the query normally (read operations are allowed in dry-run)
func (d *DryRunDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.DB.QueryContext(ctx, query, args...)
}

// dryRunResult is a mock sql.Result for dry-run mode
type dryRunResult struct{}

func (r *dryRunResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r *dryRunResult) RowsAffected() (int64, error) {
	return 0, nil
}
