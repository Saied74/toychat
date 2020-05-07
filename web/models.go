package main

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"time"
)

type userModel struct {
	dB *sql.DB
}

var (
	errNoRecord = errors.New("models: no matching record found")
	// Add a new ErrInvalidCredentials error. We'll use this later if a user
	// tries to login with an incorrect email address or password.
	errInvalidCredentials = errors.New("models: invalid credentials")
	// Add a new ErrDuplicateEmail error. We'll use this later if a user
	// tries to signup with an email address that's already in use.
	errDuplicateEmail = errors.New("models: duplicate email")
)

type user struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
	Active         bool
}

//ExchData for data exchange in gob format
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

func (st *sT) insertUserR(name, email, password string) error {
	// var err error
	exchData := ExchData{
		Name:     name,
		Email:    email,
		Password: password,
		Action:   "insert",
	}
	sendData, err := exchData.toGob()
	if err != nil {
		return err
	}
	answer := st.chatConnection(string(sendData), "forDB", "")
	exchData.fromGob(answer)

	return exchData.decodeErr()

}

func (st *sT) authenticateUserR(email, password string) (int, error) {
	exchData := ExchData{
		Email:    email,
		Password: password,
		Action:   "authenticate",
	}
	sendData, err := exchData.toGob()
	if err != nil {
		return 0, err
	}
	answer := st.chatConnection(string(sendData), "forDB", "")
	exchData.fromGob(answer)
	return exchData.ID, exchData.decodeErr()
}

func (st *sT) getUserR(id int) (*user, error) {
	exchData := ExchData{
		ID:     id,
		Action: "getuser",
	}

	sendData, err := exchData.toGob()
	if err != nil {
		return &user{}, err
	}
	answer := st.chatConnection(string(sendData), "forDB", "")
	exchData.fromGob(answer)

	return exchData.pullUser(), exchData.decodeErr()
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

func (e *ExchData) decodeErr() error {
	switch e.ErrType {
	case noErr:
		return nil
	case errZero:
		if e.Err == "" {
			return errNoRecord
		}
		return fmt.Errorf(e.Err)
	case noRecord:
		return errNoRecord
	case invalidCreds:
		return errInvalidCredentials
	case duplicateMail:
		return errDuplicateEmail
	}
	return fmt.Errorf("error decoder failed %d", int(e.ErrType))
}
