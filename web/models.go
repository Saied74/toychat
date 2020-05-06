package main

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
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
	Err            error
}

func (m *userModel) insertUser(name, email, password string) error {
	// exchData := ExchData{
	// 	Name: name,
	// 	Email: email,
	// 	Password: password,
	// 	Action: "insert",
	// }
	//
	// sendData, err := exchData.toGob()
	// if err != nil {
	// 	return ExchData{}, fmt.Errorf("user did not gob %v", err)
	// }

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err //note we are not returning any words so we can check for the error
	}
	stmt := `INSERT INTO users (name, email, hashed_password, created) VALUES(?, ?, ?, UTC_TIMESTAMP())`

	_, err = m.dB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 &&
				strings.Contains(mySQLError.Message, "users_uc_email") {
				return errDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (m *userModel) authenticateUser(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password FROM users WHERE email = ? AND active = TRUE"
	row := m.dB.QueryRow(stmt, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errInvalidCredentials
		}
		return 0, err
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, errInvalidCredentials
		}
		return 0, err
	}
	return id, nil
}

func (m *userModel) getUser(id int) (*user, error) {
	var u = &user{}

	stmt := "SELECT id, name, email, created, active FROM users WHERE id = ?"
	row := m.dB.QueryRow(stmt, id)
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Created, &u.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errNoRecord
		}
		return nil, err
	}
	return u, nil
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
