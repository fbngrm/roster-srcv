package database

import (
	"context"
	"database/sql"
	"time"
)

// DB represents a database with a timeout context for individual requests.
type DB struct {
	db         *sql.DB
	name       string
	reqTimeout time.Duration
}

func New(db *sql.DB, name string, timeout time.Duration) *DB {
	return &DB{
		db:         db,
		name:       name,
		reqTimeout: timeout,
	}
}

// Connect creates a connection to a database
func Connect(driverName, dsn, name string, reqTimeout time.Duration) (*DB, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	d := &DB{
		db:   db,
		name: name,
	}
	ctx, cancel := d.RequestContext(context.Background())
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	return d, nil
}

func (db *DB) GetDB() *sql.DB {
	return db.db
}

// RequestContext returns context for use with queries that can be called for
// individual requests.
func (db *DB) RequestContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if db.reqTimeout == 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, db.reqTimeout)
}

// Close closes a DB, and returns any error generated during closing the connection.
func (db *DB) Close() error {
	return db.db.Close()
}
