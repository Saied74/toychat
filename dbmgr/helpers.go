package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	// "toychat/pkg/broker"
	// "toychat/pkg/broker"
	// "github.com/Saied74/toychat/pkg/broker"
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

func (app *App) processDBRequest(data []byte) (*ExchData, error) {
	exchData := ExchData{}
	err := exchData.fromGob(data)
	if err != nil {
		exchData.encodeErr(err)
		return &exchData, err
	}

	switch exchData.Action {
	case "insert":
		name := exchData.Name
		email := exchData.Email
		password := exchData.Password
		err := app.users.insertUser(name, email, password)
		// if err != nil {
		exchData.encodeErr(err)
		return &exchData, err
		// }
	case "authenticate":
		email := exchData.Email
		password := exchData.Password
		id, err := app.users.authenticateUser(email, password)
		if err != nil {
			app.errorLog.Printf("in authenticate case after authentiateUser call %v",
				err)
			exchData.encodeErr(err)
			return &exchData, err
		}
		exchData.ID = id
		exchData.encodeErr(err)
		return &exchData, nil
	case "getuser":
		id := exchData.ID
		user, err := app.users.getUser(id)
		if err != nil {
			exchData.encodeErr(err)
			return &exchData, err
		}
		exchData.pushUser(user)
		exchData.encodeErr(err)
		return &exchData, nil
	default:
		exchData.encodeErr(err)
		return &exchData, fmt.Errorf("command not implemented")
	}
	// return &exchData, fmt.Errorf("fell off the bottom")
}

func (e *ExchData) toGob() ([]byte, error) {
	b := &bytes.Buffer{}
	enc := gob.NewEncoder(b)
	err := enc.Encode(*e)
	if err != nil {
		return []byte{}, fmt.Errorf("failed gob Encode %v", err)
	}
	return b.Bytes(), nil
}

func (e *ExchData) fromGob(g []byte) error {
	b := &bytes.Buffer{}
	b.Write(g)
	dec := gob.NewDecoder(b)
	err := dec.Decode(e)
	if err != nil {
		return fmt.Errorf("failed in dbmgr to gob decode %v", err)
	}
	return nil
}

// func (e *ExchData) pullUser() *user {
// 	newUser := user{
// 		ID:             e.ID,
// 		Name:           e.Name,
// 		Email:          e.Email,
// 		HashedPassword: e.HashedPassword,
// 		Created:        e.Created,
// 		Active:         e.Active,
// 	}
// 	return &newUser
// }
//
func (e *ExchData) pushUser(u *user) {
	e.ID = u.ID
	e.Name = u.Name
	e.Email = u.Email
	e.HashedPassword = u.HashedPassword
	e.Created = u.Created
	e.Active = u.Active
}

func (e *ExchData) encodeErr(err error) {
	if err == nil {
		e.ErrType = noErr
		return
	}
	if errors.Is(err, errNoRecord) {
		e.ErrType = noRecord
		return
	}
	if errors.Is(err, errInvalidCredentials) {
		e.ErrType = invalidCreds
		return
	}
	if errors.Is(err, errDuplicateEmail) {
		return
	}
	e.ErrType = errZero
	e.Err = fmt.Sprintf("%v", err)
}
