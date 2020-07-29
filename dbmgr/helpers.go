package main

import (
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	nats "github.com/nats-io/nats.go"
	"github.com/saied74/toychat/pkg/broker"
)

func (app *App) processDBRequests(msg *nats.Msg, conn *nats.Conn) {
	gob.Register(broker.TableRows{})
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
		case "agent":
			err = app.users.getAgent(exchange)
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

func (m *userModel) insert(e *broker.Exchange) error {
	stmt := buildInsertStmt(e.Table, e.Put)
	for _, c := range e.Spec {
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
	var iter bool
	newPeople := broker.TableRows{}
	stmt := buildGetStmt(e.Table, e.Get, e.SpecList)
	for _, c := range e.Spec {
		rows, err := m.dB.Query(stmt, c...)
		if err != nil {
			return err
		}
		defer rows.Close()
		iter = false
		for rows.Next() {
			iter = true
			person := broker.TableRow{}
			g := person.GetItems(e.Get)
			err = rows.Scan(g...)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return broker.ErrNoRecord
				}
				return err
			}
			err = person.GetBack(e.Get, g)
			if err != nil {
				return fmt.Errorf("GetBack Error: %v", err)
			}
			newPeople = append(newPeople, person)
		}
	}
	if iter {
		e.Tables = newPeople
		return nil
	}
	e.Tables = broker.TableRows{}
	return broker.ErrNoRecord
}

func (m *userModel) put(e *broker.Exchange) error {
	stmt := buildPutStmt(e.Table, e.Put, e.SpecList)
	for _, c := range e.Spec {
		_, err := m.dB.Exec(stmt, c...)
		if err != nil {
			return err
		}
	}
	return nil
}

//getAgent is coded longhand without any abstraction since it is only one of its
//kind for now.  We will see what happens as the application develops.
func (m *userModel) getAgent(e *broker.Exchange) error {
	userMsgs := broker.TableRows{}
	stmt := "SELECT id, dialog  FROM admins WHERE role='agent' AND dialog < 3 ORDER BY dialog"
	tx, err := m.dB.Begin()
	if err != nil {
		return err
	}
	rows := tx.QueryRow(stmt)
	userMsg := broker.TableRow{}
	err = rows.Scan(&userMsg.AgentID, &userMsg.Dialog)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return broker.ErrNoRecord
		}
		return err
	}
	userMsg.Dialog++
	stmt = "UPDATE admins SET dialog = ? WHERE id = ?"
	_, err = m.dB.Exec(stmt, &userMsg.Dialog, &userMsg.Dialog)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	userMsgs = append(userMsgs, userMsg)
	e.Tables = userMsgs
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
		case "started":
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
