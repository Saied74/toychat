package main

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/saied74/toychat/pkg/broker"
	"golang.org/x/crypto/bcrypt"
)

type userModel struct {
	dB *sql.DB
}

type user struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
	Active         bool
}

func (m *userModel) insertUser(name, email, password string) error {
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
				return broker.ErrDuplicateEmail
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
			return 0, broker.ErrInvalidCredentials
		}
		return 0, err
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, broker.ErrInvalidCredentials
		}
		return 0, err
	}
	return id, nil
}

func (m *userModel) getUser(id int) (*broker.ExchData, error) {
	var u = &broker.ExchData{}

	stmt := "SELECT id, name, email, created, active FROM users WHERE id = ?"
	row := m.dB.QueryRow(stmt, id)
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Created, &u.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, broker.ErrNoRecord
		}
		return nil, err
	}
	return u, nil
}
