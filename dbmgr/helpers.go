package main

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/go-sql-driver/mysql"
	nats "github.com/nats-io/nats.go"
	"github.com/saied74/toychat/pkg/broker"
	"github.com/saied74/toychat/pkg/centerr"
)

//The interface to the dbmgr is through broker.Exchange object.  There are
//three possible actions, put, get, and insert.  Once the gob data is decoded,
//the corresponding function is called and the result is gob encoded and
//returned through the nats mailbox.

func (app *App) processDBRequests(msg *nats.Msg, conn *nats.Conn) {
	var err error
	var exchange = &broker.Exchange{}
	err = exchange.FromGob(msg.Data)
	if err != nil {
		exchange.EncodeErr(err)
	}
	if err == nil {
		switch exchange.Action {
		case "get":
			err = app.users.get(exchange)
			exchange.EncodeErr(err)
		case "put":
			err = app.users.put(exchange)
			exchange.EncodeErr(err)
		case "insert":
			err = app.users.insert(exchange)
			exchange.EncodeErr(err)
		default:
			exchange.EncodeErr(err)
		}
	}
	g, err := exchange.ToGob()
	if err != nil {
		conn.Publish(msg.Reply, []byte{})
	} else {
		conn.Publish(msg.Reply, g)
	}
}

type userModel struct {
	dB *sql.DB
}

var allCol = []string{"id", "name", "email", "hashed_password", "created",
	"role", "active", "online"}

func (m *userModel) insert(e *broker.Exchange) error {
	stmt := buildInsertStmt(e.Table, e.Put)
	centerr.InfoLog.Printf("Insert Statement: %s", stmt)
	for _, p := range e.People {
		c := p.BuildInsert(e.Put)
		_, err := m.dB.Exec(stmt, c...)
		if err != nil {
			var mySQLError *mysql.MySQLError
			if errors.As(err, &mySQLError) {
				if mySQLError.Number == 1062 &&
					strings.Contains(mySQLError.Message, e.Table+"_uc_email") {
					return broker.ErrDuplicateEmail
				}
			}
			return err
		}
	}
	return nil
}

func (m *userModel) get(e *broker.Exchange) error {
	stmt := buildGetStmt(e.Table, e.Get, e.Spec)
	newPeople := []broker.Person{}
	for _, p := range e.People {
		c := p.GetSpec(e.Spec)
		person := broker.Person{}
		g := person.GetItems(e.Get)
		rows, err := m.dB.Query(stmt, c...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(g...)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return broker.ErrNoRecord
				}
				return err
			}
			person.GetBack(e.Get, g)
			newPeople = append(newPeople, person)
		}
	}
	e.People = newPeople
	return nil
}

func (m *userModel) put(e *broker.Exchange) error {
	stmt := buildPutStmt(e.Table, e.Put, e.Spec)
	log.Printf("put statement: %s", stmt)
	for _, p := range e.People {
		c := p.Specify(e.Put, e.Spec)
		log.Printf("put condition: %v", c)
		_, err := m.dB.Exec(stmt, c...)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildInsertStmt(table string, put []string) string {
	stmt := "INSERT INTO " + table + " ("
	putFields := strings.Join(put[:], ", ")
	stmt += putFields
	stmt += ") VALUES("
	for _, item := range put {
		switch item {
		case "created":
			stmt += "UTC_TIMESTAMP(), "
		default:
			stmt += "?, "
		}
	}
	stmt = strings.TrimSuffix(stmt, ", ")
	stmt += ")"
	return stmt
}

func buildGetStmt(table string, get, spec []string) string {
	stmt := "SELECT "
	getFields := strings.Join(get[:], ", ")
	stmt += getFields
	stmt += " FROM " + table + " WHERE "
	specFields := strings.Join(spec[:], " = ? AND ")
	stmt += specFields
	stmt += " = ?"
	return stmt
}

func buildPutStmt(table string, put, spec []string) string {
	stmt := "UPDATE " + table + " SET "
	putFields := strings.Join(put[:], " = ? AND ")
	stmt += putFields + " = ?"
	stmt += " WHERE "
	specFields := strings.Join(spec[:], " = ? AND ")
	stmt += specFields
	stmt += " = ?"
	return stmt

}
