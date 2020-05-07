package main

import (
	"database/sql"
	"log"
	"os"
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

//ExchData is the structure for exchanging data over the nats message broker
type ExchData struct {
	ID             int
	Name           string
	Email          string
	Password       string
	Created        time.Time
	Active         bool
	HashedPassword []byte
	Authenticated  bool
	Action         string //authenticate, insert, and getuser are permitted actions
	ErrType        errMsg
	Err            string
}

func (e *ExchData) String(msg string) {
	log.Printf("Start: %s\n", msg)
	log.Printf("ID: %d\n", e.ID)
	log.Printf("Name: %s\n", e.Name)
	log.Printf("Email: %s\n", e.Email)
	log.Printf("Active: %v\n", e.Active)
	log.Printf("Action: %s\n", e.Action)
	log.Printf("ErrType: %v\n", e.ErrType)
	log.Printf("Err: %s\n", e.Err)
}

func main() {
	var err error
	dsn := "toy:f00lish@/toychat?parseTime=true"

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

	for {
		sub, _ := nc1.SubscribeSync("forDB") //, func(matMsg *nats.Msg) { //playMatHandler)
		msg, err := sub.NextMsg(10 * time.Hour)
		if err != nil {
			app.errorLog.Printf("Error from Sub Sync %v: ", err)
		}
		exchData, err := app.processDBRequest(msg.Data)
		if err != nil {
			app.errorLog.Printf("processing DB request failed %v", err)
		}
		g, err := exchData.toGob()
		if err != nil {
			app.errorLog.Printf("did not go to Gob %v", err)
			nc1.Publish(msg.Reply, []byte{})
		} else {

			nc1.Publish(msg.Reply, g)
		}
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
