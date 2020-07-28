//Copyright (c) 2020 Saied Seghatoleslami
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

//dbmgr application is the database manager interface for toychat.  It supports
//admin, message and dialog tables.  Session managers directly interface with
//the database manager.  In the future, when and if the front end and the back
//end are run on seperate servers or in a more scalable solution, each will
//have its own mysql database for the session data.
//
//The dbmgr expects a password with the -pw flag for the database at startup.
//It also expects the nats server to be up and running.
//
//The interface to the dbmgr is through nats.  It listens on the nats.DefaultURL
//looking for messages addressed to "forDB".  It blocks until it recieves the
//message with a return mailbox.  Once it recieves that, it fires off a goroutine
//to process the request and goes back to listening for the next request.
//
//When firing off the go routine,  it hands it pointers to the connection and
//to the message.  The goroutine uses connections.Publish to reply to the
//message once the database action is completed.
//
//Objects and methods for the communication are defined in the broker pkg so
//they are usable on both sides of the interface.  See the broker package
//documentation for details.  Serialization of the data is accomplished by gob
//encoding.  Since errors do not encode into gob, a set of error encoding and
//decoding are also provided by the broker package.
//
//The exchange object is the main vehicle for communication to the dbmgr.
//Its field "Action" defines the request to the dbmgr.  the processDBRequests
//(the method run as a go routine) switches on this field and invokes the
//method for processing the request.  Currently, insert, get, and put are
//supported.  "agent" for selecting an agent for a new dialog is not abstracted
//or generalized like the other three.  If it turns out that multiple insances
//of this function is needed, I will try and see.  Agent selection ia a
//transaction so two go routines cannot grab the same agent.
//
//For all three methods, the SQL statement are generated on the fly from the
//Exchage object fields (see broker documentation for the details).  This allows
//for the requeser to request specific fields to be returned and specific
//conditions to be met.  Also, for differnet tables, it allows the same
//method to be used.
//
//The sql library Execute and Quarty statements take the SQL statement as thier
//first variable and that is built as described above.  The next set of arugments
//to both function is variadic arguments of the type interface{}.  They are
//pointers to the variables that accept the database search results from the
//select statement or are the conditions for the select as indicated by ? symbol
//for executing the prepared statements.  To make the
//insert, get and put functions able to handle different tables with different
//schema, a slice of []interface{} is supplied by the caller over nats (again
//see broker documentation as how these are built).  This is not the case for
//rows.Scan method since the number of rows returned by the Query statement
//is not known in advance.
//
//The error handling for duplicat rows and no rows found are straight out of
//Alex Edwards Let's go book.
//
//the MySQL database, in addition to the session tables as indicated above,
//has the users, admins (which includes agents) dialogs and messages tables.

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
