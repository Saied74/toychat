package main

import (
	"database/sql"
	"flag"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	nats "github.com/nats-io/nats.go"
	"github.com/saied74/toychat/pkg/centerr"
)

//dbmgr is composed of two files, main and helper.  Main (this file), interfaces
//to nats server in the main function and hands the nats.Msg and nats.Conn opbjects
//to the processing function (processDBRequests).  Once the request is hnande
//off, the listenter goes back to listending.  Once the database action is
//complete, processDBRequests returns the results through the nats mailbox.

//App for inseertion of variables into functions
type App struct {
	users *userModel
}

func main() {

	pw := flag.String("pw", "password", "database password is always required")
	flag.Parse()

	var err error
	dsn := "toy:password@/toychat?parseTime=true"
	dsn = strings.Replace(dsn, "password", *pw, 1)

	db, err := openDB(dsn)
	if err != nil {
		centerr.ErrorLog.Fatal(err)
	}
	defer db.Close()

	//the function of app is dpenendency injection.
	app := App{
		users: &userModel{dB: db},
	}

	nc1, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		centerr.ErrorLog.Fatal("Error from onnection", err)
	}
	defer nc1.Close()

	sub, _ := nc1.SubscribeSync("forDB")
	for {
		msg, err := sub.NextMsg(10 * time.Hour)
		if err != nil {
			centerr.ErrorLog.Printf("Error from Sub Sync %v: ", err)
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
