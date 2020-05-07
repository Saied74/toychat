package main

import (
	"fmt"
	"io"
	"log"

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

func (app *App) processDBRequest(data []byte) (*broker.ExchData, error) {
	exchData := broker.ExchData{}
	err := exchData.FromGob(data)
	if err != nil {
		exchData.EncodeErr(err)
		return &exchData, err
	}

	switch exchData.Action {
	case "insert":
		name := exchData.Name
		email := exchData.Email
		password := exchData.Password
		err := app.users.insertUser(name, email, password)
		exchData.EncodeErr(err)
		return &exchData, err
	case "authenticate":
		email := exchData.Email
		password := exchData.Password
		id, err := app.users.authenticateUser(email, password)
		if err != nil {
			app.errorLog.Printf("in authenticate case after authentiateUser call %v",
				err)
			exchData.EncodeErr(err)
			return &exchData, err
		}
		exchData.ID = id
		exchData.EncodeErr(err)
		return &exchData, nil
	case "getuser":
		id := exchData.ID
		exchData, err := app.users.getUser(id)
		if err != nil {
			exchData.EncodeErr(err)
			return exchData, err
		}
		// exchData.PushUser(user)
		exchData.EncodeErr(err)
		return exchData, nil
	default:
		exchData.EncodeErr(err)
		return &exchData, fmt.Errorf("command not implemented")
	}
}
