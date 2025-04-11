# sqllog

Go SQL database driver that supports logging.

```sh
go get github.com/mengdu/sqllog
```

```go
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

  rows, err := db.Query(`select ?, ?, ?, ?, ?, ?, ?`, 123, 3.14, true, nil, "abc", 'a', []byte("abc"))
	if err != nil {
		panic(err)
	}
	for rows.Next() {
	}
}
```

**Output**

```log
[db:2025-04-11T15:15:39.539+08:00] select ?, ?, ?, ?, ?, ?, ? [7/8]any[123, 3.14, true, nil.(<invalid>), "abc", 97, [3/3]uint8[97, 98, 99]] 965.911µs false
```
