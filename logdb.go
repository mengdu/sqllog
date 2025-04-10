package logdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"
)

type Record struct {
	Query     string
	Args      []driver.NamedValue
	Effect    bool // whether the query has side effects
	Preparing bool
	Err       error
	At        time.Time
	Ts        time.Duration
}
type Logger func(ctx context.Context, log Record)

func Open(driverName string, dataSourceName string, log Logger) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	if log == nil {
		return db, nil
	}

	return sql.OpenDB(&connector{
		driver: db.Driver(),
		dsn:    dataSourceName,
		log:    log,
	}), nil
}

// connector implements driver.Connector
type connector struct {
	driver driver.Driver
	dsn    string
	log    Logger
}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.driver.Open(c.dsn)
	return &connection{Conn: conn, log: c.log}, err
}

func (c *connector) Driver() driver.Driver {
	return c.driver
}

type connection struct {
	driver.Conn
	log Logger
}

// PrepareContext implements driver.ConnPrepareContext
func (c *connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	record := Record{
		Query: query,
		At:    time.Now(),
	}

	st, err := c.Conn.(driver.ConnPrepareContext).PrepareContext(ctx, query)
	if err != nil {
		record.Err = err
		record.Ts = time.Since(record.At)
		record.Preparing = true
		if c.log != nil {
			c.log(ctx, record)
		}
		return st, err
	}
	return &stmt{Stmt: st, record: record, log: c.log}, err
}

type stmt struct {
	driver.Stmt
	record Record
	log    Logger
}

// QueryContext implements driver.StmtQueryContext
func (s *stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	defer func() {
		s.record.Effect = false
		s.record.Args = args
		s.record.Ts = time.Since(s.record.At)
		s.record.Err = err
		if s.log != nil {
			s.log(ctx, s.record)
		}
	}()
	return s.Stmt.(driver.StmtQueryContext).QueryContext(ctx, args)
}

// ExecContext implements driver.StmtExecContext
func (s *stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (result driver.Result, err error) {
	defer func() {
		s.record.Effect = true
		s.record.Args = args
		s.record.Ts = time.Since(s.record.At)
		s.record.Err = err
		if s.log != nil {
			s.log(ctx, s.record)
		}
	}()
	return s.Stmt.(driver.StmtExecContext).ExecContext(ctx, args)
}
