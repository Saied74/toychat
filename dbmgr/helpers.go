package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
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
		exchData.Err = fmt.Errorf("gob did not decode %v", err)
		return &exchData, err
	}

	switch exchData.Action {
	case "insert":
		name := exchData.Name
		email := exchData.Email
		password := exchData.Password
		err := app.users.insertUser(name, email, password)
		if err != nil {
			exchData.Err = fmt.Errorf("user did not insert %v", err)
			return &exchData, err
		}
	case "authenticate":
		email := exchData.Email
		password := exchData.Password
		id, err := app.users.authenticateUser(email, password)
		if err != nil {
			exchData.Err = fmt.Errorf("user %s did not authenticate %v", email, err)
			return &exchData, err
		}
		exchData.ID = id
		return &exchData, nil
	case "getuser":
		id := exchData.ID
		user, err := app.users.getUser(id)
		if err != nil {
			exchData.Err = fmt.Errorf("could not get user %d because %v", id, err)
			return &exchData, err
		}
		exchData.pushUser(user)
	default:
		exchData.Err = fmt.Errorf("command not implemented")
		return &exchData, exchData.Err
	}
	return &exchData, fmt.Errorf("fell off the bottom")
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
	dec := gob.NewDecoder(b)
	err := dec.Decode(e)
	if err != nil {
		return fmt.Errorf("failed screen gob decode %v", err)
	}
	return nil
}

func (e *ExchData) pullUser() *user {
	newUser := user{
		ID:             e.ID,
		Name:           e.Name,
		Email:          e.Email,
		HashedPassword: e.HashedPassword,
		Created:        e.Created,
		Active:         e.Active,
	}
	return &newUser
}

func (e *ExchData) pushUser(u *user) {
	e.ID = u.ID
	e.Name = u.Name
	e.Email = u.Email
	e.HashedPassword = u.HashedPassword
	e.Created = u.Created
	e.Active = u.Active
}
