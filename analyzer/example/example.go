package example

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/farislr/commoneer/rdbx"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	_rsyncpool "github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

func initDB() rdbx.DBTX {
	username := "root"
	password := ""
	host := "localhost"
	port := "3306"
	dbname := "test"
	parsetime := true
	loc := "Asia%2FJakarta"

	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=%t&loc=%s", username, password, host, port, dbname, parsetime, loc)

	db, err := sql.Open("mysql", connection)
	if err != nil {
		panic(err)
	}

	rClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	pool := _rsyncpool.NewPool(rClient)
	rsync := redsync.New(pool)

	return rdbx.NewDBx(db, rClient, rsync)
}

func queryx() {
	db := initDB()

	var model struct {
		ID        string
		Name      string
		CreatedAt sql.NullTime
	}

	if err := db.Queryx(context.Background(), "SELECT * FROM model_table", model); err != nil {
		log.Println(err)
	}

	fmt.Printf("model: %v\n", model)
}
