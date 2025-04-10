package main

import (
	"context"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mengdu/fmtx"
	"github.com/mengdu/logdb"
)

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	// db, err := sql.Open("mysql", dsn)
	db, err := logdb.Open("mysql", dsn, func(ctx context.Context, log logdb.Record) {
		fmtx.Println("[db]", log.Query, log.Args, log.Ts.String(), log.Preparing, log.Err)
	})
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	// rows, err := db.Query(`select 1 as a, "2" as b, ? as c`, 123)
	rows, err := db.Query(`select 1 as a, "2" as b, 123 as c`)
	fmtx.Println(err)
	for rows.Next() {
		var a int
		var b string
		var c any
		err := rows.Scan(&a, &b, &c)
		fmtx.Println(err, a, b, c)
	}
	res, err := db.Exec(`insert into test set str = ?`, time.Now().String())
	fmtx.Println(err)
	fmtx.Println(res.RowsAffected())
	db.QueryRow("select str from test where id = ?", 1)
	stmt, err := db.Prepare("select str from test where id = ?")
	fmtx.Println(err)
	stmt.Query(2)
	tx, err := db.Begin()
	fmtx.Println(err)
	rows, err = tx.Query("select str from test where id = ?", 4)
	for rows.Next() {
	}
	fmtx.Println(err)
	res, err = tx.Exec("insert into test set str = ?", time.Now().Format(time.RFC3339))
	fmtx.Println(err)
	// fmtx.Println(res.RowsAffected())
	tx.Commit()
}
