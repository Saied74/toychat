package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	nats "github.com/nats-io/nats.go"
)

//App for inseertion of variables into functions
type App struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	users    *userModel
}

func main() {

	pw := flag.String("pw", "password", "database password is always required")
	flag.Parse()

	var err error
	dsn := "toy:f00lish@/toychat?parseTime=true"
	dsn = strings.Replace(dsn, "password", *pw, 1)
	infoLog := getInfoLogger(os.Stdout)
	errorLog := getErrorLogger(os.Stdout)

	db, err := openDB(dsn)
	if err != nil {
		errorLog().Fatal(err)
	}
	defer db.Close()

	app := App{
		infoLog:  infoLog(),
		errorLog: errorLog(),
		users:    &userModel{dB: db},
	}

	nc1, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		app.errorLog.Fatal("Error from onnection", err)
	}
	defer nc1.Close()

	sub, _ := nc1.SubscribeSync("forDB")
	for {
		msg, err := sub.NextMsg(10 * time.Hour)
		if err != nil {
			app.errorLog.Printf("Error from Sub Sync %v: ", err)
		}
		go app.processDBRequests(msg, nc1)
	}
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
