package main

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mengdu/fmtx"
	"github.com/mengdu/sqllog"
)

// implements sqllog.Logger
type SqlLog struct{}

func (l SqlLog) Log(ctx context.Context, log sqllog.Record) {
	args := []any{}
	for _, arg := range log.Args {
		args = append(args, arg.Value)
	}
	errMsg := ""
	if log.Err != nil {
		errMsg = log.Err.Error()
	}
	fmtx.Println(fmt.Sprintf("[db:%s]", log.At.Format("2006-01-02T15:04:05.000Z07:00")), log.Query, args, log.Ts.String(), log.Preparing, errMsg)
}

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	// db, err := sql.Open("mysql", dsn)
	db, err := sqllog.Open("mysql", dsn, SqlLog{})
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	rows, err := db.Query(`select ?, ?, ?, ?, ?, ?, ?`, 123, 3.14, true, nil, "abc", 'a', []byte("abc"))
	if err != nil {
		panic(err)
	}
	for rows.Next() {
	}

	rows, err = db.Query(`select 1 as a, "2" as b, 123 as c`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
	}

	// rows, err = db.Query(`select str from test where id = :id`, sql.Named("id", 1))
	// if err != nil {
	// 	panic(err)
	// }
	// for rows.Next() {
	// }

	_, err = db.Exec(`insert into test set str = ?`, time.Now().String())
	if err != nil {
		panic(err)
	}

	db.QueryRow("select str from test where id = ?", 1)

	stmt, err := db.Prepare("select str from test where id = ?")
	if err != nil {
		panic(err)
	}
	rows, err = stmt.Query(2)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
	}

	rows, err = stmt.Query(3)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
	}

	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		panic(err)
	}

	rows, err = tx.Query("select str from test where id = ?", 4)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
	}

	_, err = tx.Exec("insert into test set str = ?", time.Now().Format(time.RFC3339))
	if err != nil {
		panic(err)
	}
	tx.Commit()
}
