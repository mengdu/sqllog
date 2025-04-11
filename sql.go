package sqllog

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"time"
)

type Record struct {
	Query     string              // sql query
	Args      []driver.NamedValue // arguments
	Effect    bool                // whether the query has side effects
	Preparing bool                // is the query preparing a statement
	Err       error               // error from executing the query
	At        time.Time           // start time
	Ts        time.Duration       // duration of the query
}
type Logger interface {
	Log(ctx context.Context, log Record)
}

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

// // CheckNamedValue implements driver.NamedValueChecker
// func (c *connection) CheckNamedValue(n *driver.NamedValue) error {
// 	fmt.Println("CheckNamedValue", n.Value)
// 	fmt.Println(string(debug.Stack()))
// 	return c.Conn.(driver.NamedValueChecker).CheckNamedValue(n)
// }

// PrepareContext implements driver.ConnPrepareContext
func (c *connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	record := Record{
		Query: query,
		At:    time.Now(),
	}

	var st driver.Stmt
	var err error
	if v, ok := c.Conn.(driver.ConnPrepareContext); ok {
		st, err = v.PrepareContext(ctx, query)
	} else {
		st, err = c.Conn.Prepare(query)
	}

	if err != nil {
		if c.log != nil {
			record.Err = err
			record.Ts = time.Since(record.At)
			record.Preparing = true
			c.log.Log(ctx, record)
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
		if s.log != nil {
			s.record.Effect = false
			s.record.Args = args
			s.record.Ts = time.Since(s.record.At)
			s.record.Err = err
			s.log.Log(ctx, s.record)
		}
	}()

	if v, ok := s.Stmt.(driver.StmtQueryContext); ok {
		return v.QueryContext(ctx, args)
	}
	values, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	return s.Stmt.Query(values)
}

// ExecContext implements driver.StmtExecContext
func (s *stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (result driver.Result, err error) {
	defer func() {
		if s.log != nil {
			s.record.Effect = true
			s.record.Args = args
			s.record.Ts = time.Since(s.record.At)
			s.record.Err = err
			s.log.Log(ctx, s.record)
		}
	}()
	if v, ok := s.Stmt.(driver.StmtExecContext); ok {
		return v.ExecContext(ctx, args)
	}
	values, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	return s.Stmt.Exec(values)
}

func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("sql: driver does not support the use of Named Parameters")
		}
		dargs[n] = param.Value
	}
	return dargs, nil
}
