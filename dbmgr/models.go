package main

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/saied74/toychat/pkg/broker"
	"golang.org/x/crypto/bcrypt"
)

type userModel struct {
	dB *sql.DB
}

func (m *userModel) insertUser(table, role, name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err //note we are not returning any words so we can check for the error
	}
	stmt := `INSERT INTO ` + table +
		` (name, email, hashed_password, created, role) VALUES(?, ?, ?, UTC_TIMESTAMP(), ?)`

	_, err = m.dB.Exec(stmt, name, email, string(hashedPassword), role)
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

func (m *userModel) authenticateUser(table, role, email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := `SELECT id, hashed_password FROM ` + table +
		` WHERE email = ? AND role = ? AND active = TRUE`
	// log.Printf("From select user %s", stmt)
	row := m.dB.QueryRow(stmt, email, role)
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

func (m *userModel) getUser(table string, id int) (*broker.ExchData, error) {
	var u = broker.ExchData{}

	stmt := `SELECT id, name, email, created, active FROM ` + table +
		` WHERE id = ?`
	row := m.dB.QueryRow(stmt, id)
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Created, &u.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, broker.ErrNoRecord
		}
		return &broker.ExchData{}, err
	}
	return &u, nil
}

func (m *userModel) getByStatus(table, role string, status bool) (*broker.ExchData, error) {
	// log.Println("got to getByStatus func", table, status)
	var u = &broker.ExchData{}

	stmt := `SELECT id, name, email, created, role FROM ` + table +
		` WHERE active = ? AND role = ?`
	rows, err := m.dB.Query(stmt, status, role)
	if err != nil {
		return &broker.ExchData{}, err
	}
	defer rows.Close()

	for rows.Next() {
		p := &broker.Person{}
		err = rows.Scan(&p.ID, &p.Name, &p.Email, &p.Created, &p.Role)
		if err != nil {
			return &broker.ExchData{}, err
		}
		// log.Println("got to people", *p)
		u.People = append(u.People, *p)
	}

	if err = rows.Err(); err != nil {
		return &broker.ExchData{}, err
	}
	return u, nil
}

func (m *userModel) activation(table string, people []broker.Person) error {
	stmt := `UPDATE ` + table + ` SET active = ? WHERE id= ?`
	for _, person := range people {
		log.Println("deep in there", person.Active, person.ID)
		_, err := m.dB.Exec(stmt, person.Active, person.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *userModel) chgPwd(table, role, email, password string) error {
	log.Println("in chgPwd", table, role, email, password)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err //note we are not returning any words so we can check for the error
	}
	stmt := `UPDATE ` + table + ` SET hashed_password = ? WHERE role = ? AND email= ? AND active = TRUE`

	_, err = m.dB.Exec(stmt, string(hashedPassword), role, email)
	if err != nil {
		return err
	}
	return nil
}
