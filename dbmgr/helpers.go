package main

import (
	"io"
	"log"

	nats "github.com/nats-io/nats.go"
	"github.com/saied74/toychat/pkg/broker"
)

//these loggers will have to be moved to a package file.
func getInfoLogger(out io.Writer) func() *log.Logger {
	infoLog := log.New(out, "INFO\t", log.Ldate|log.Ltime)
	return func() *log.Logger {
		return infoLog
	}
}

func getErrorLogger(out io.Writer) func() *log.Logger {
	errorLog := log.New(out, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return func() *log.Logger {
		return errorLog
	}
}

//recieves the message from the loop in the main goroutine and runs as a
//goroutine itself.  First, it builds a copy of the struct and decode the
//gob string that came over the wire.  there are three possible commands.
//insert, authenticate, and getuser.  It performs these by calling the
//database methods in the models file.  Since error types cannot be gob
//encoded, it string encodes them so they can be sent over the wire and
//decoded on the far side.
func (app *App) processDBRequests(msg *nats.Msg, conn *nats.Conn) {
	var err error
	var exchData = &broker.ExchData{}
	err = exchData.FromGob(msg.Data)
	if err != nil {
		exchData.EncodeErr(err)
	}
	// log.Println("got to getByStatus zzzz", exchData.Action)
	if err == nil {
		switch exchData.Action {
		case "insert":
			name := exchData.Name
			email := exchData.Email
			password := exchData.Password
			table := exchData.Table
			role := exchData.Role
			err := app.users.insertUser(table, role, name, email, password)
			exchData.EncodeErr(err)

		case "authenticate":
			email := exchData.Email
			password := exchData.Password
			table := exchData.Table
			role := exchData.Role
			id, err := app.users.authenticateUser(table, role, email, password)
			if err != nil {
				// app.errorLog.Printf("in authenticate case after authentiateUser call %v",
				// 	err)
				exchData.EncodeErr(err)
			} else {
				exchData.ID = id
				exchData.EncodeErr(err)
			}
		case "getuser":
			id := exchData.ID
			table := exchData.Table
			exchData, err = app.users.getUser(table, id)
			if err != nil {
				exchData.EncodeErr(err)
			} else {
				exchData.EncodeErr(err)
			}
		case "getByStatus":
			// log.Println("got to getByStatus xxx", exchData)
			table := exchData.Table
			status := exchData.Active
			role := exchData.Role
			exchData, err = app.users.getByStatus(table, role, status)
			exchData.EncodeErr(err)
		case "doActivation":
			log.Println("in doActivation", exchData)
			table := exchData.Table
			people := exchData.People
			err = app.users.activation(table, people)
			exchData.EncodeErr(err)
		case "chgPwd":
			table := exchData.Table
			role := exchData.Role
			email := exchData.Email
			password := exchData.Password
			err = app.users.chgPwd(table, role, email, password)
			log.Printf("error from db insert %v", err)
			exchData.EncodeErr(err)
		default:
			exchData.EncodeErr(err)
		}
	}
	g, err := exchData.ToGob()
	if err != nil {
		conn.Publish(msg.Reply, []byte{})
	} else {
		conn.Publish(msg.Reply, g)
	}
}
