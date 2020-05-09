package main

import (
	"io"
	"log"

	nats "github.com/nats-io/nats.go"
	"github.com/saied74/toychat/pkg/broker"
)

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

func (app *App) processDBRequests(msg *nats.Msg, conn *nats.Conn) {
	var err error
	var exchData = &broker.ExchData{}
	err = exchData.FromGob(msg.Data)
	if err != nil {
		exchData.EncodeErr(err)
	}
	if err == nil {
		switch exchData.Action {
		case "insert":
			name := exchData.Name
			email := exchData.Email
			password := exchData.Password
			err := app.users.insertUser(name, email, password)
			exchData.EncodeErr(err)
		case "authenticate":
			email := exchData.Email
			password := exchData.Password
			id, err := app.users.authenticateUser(email, password)
			if err != nil {
				app.errorLog.Printf("in authenticate case after authentiateUser call %v",
					err)
				exchData.EncodeErr(err)
			} else {
				exchData.ID = id
				exchData.EncodeErr(err)
			}
		case "getuser":
			id := exchData.ID
			exchData, err = app.users.getUser(id)
			if err != nil {
				exchData.EncodeErr(err)
			} else {
				exchData.EncodeErr(err)
			}
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
