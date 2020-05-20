package main

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"strings"

	"github.com/go-sql-driver/mysql"
	nats "github.com/nats-io/nats.go"
	"github.com/saied74/toychat/pkg/broker"
)

//these loggers will have to be moved to a package file.
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
	stmt := `INSERT INTO ` + e.Table +
		` (name, email, hashed_password, created, role) VALUES(?, ?, ?, UTC_TIMESTAMP(), ?)`
	for _, p := range e.People {
		_, err := m.dB.Exec(stmt, p.Name, p.Email, p.HashedPassword, p.Role)
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
	}
	return nil
}

func (m *userModel) get(e *broker.Exchange) error {
	cond := e.Specify()
	stmt := buildGetStmt(e.Table, e.Spec, allCol)
	log.Printf("Stmt: %s", stmt)
	e.People = []broker.Person{}
	for _, c := range cond {
		log.Printf("Condition: %v", c)
		rows, err := m.dB.Query(stmt, c...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var p = &broker.Person{}
			err := rows.Scan(&p.ID, &p.Name, &p.Email, &p.HashedPassword, &p.Created,
				&p.Role, &p.Active, &p.Online)
			if err != nil {
				return err
			}
			e.People = append(e.People, *p)
		}
	}
	log.Printf("exchange %v", e)
	return nil
}

func (m *userModel) put(e *broker.Exchange) error {
	cond := e.Specify()
	stmt := buildPutStmt(e.Table, e.Put, e.Spec)
	log.Printf("put statement: %s", stmt)
	for _, c := range cond {
		log.Printf("put condition: %v", cond)
		_, err := m.dB.Exec(stmt, c...)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildGetStmt(table string, give, get []string) string {
	stmt := "SELECT "
	getFields := strings.Join(get[:], ", ")
	stmt += getFields
	stmt += " FROM " + table + " WHERE "
	giveFields := strings.Join(give[:], " = ? AND ")
	stmt += giveFields
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
