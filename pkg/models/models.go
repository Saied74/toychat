package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/saied74/toychat/pkg/broker"
	"github.com/saied74/toychat/pkg/centerr"
)

//UserModel wraps the sql.DB connections
type UserModel struct {
	DB *sql.DB
}

//sends user name, eamil and password to the far end to be inserted into the db.
// TODO: encode password here instead of sending it in plaintext over the wire.
//note that since it is struct that is sent, it is gob encoded.
//Also, since error type does not gob encode, it is string encoded at the
//far end and deoded back into the error type at this end.
//The requested action is text encoded into the struct field as well.

//InsertUserR gob encodes exchData and sends it to the dbmgr over nats
// func InsertUserR(table, name, email, password string) error {
// 	exchData := broker.ExchData{
// 		Name:     name,
// 		Email:    email,
// 		Password: password,
// 		Action:   "insert",
// 		Table:    table,
// 	}
// 	sendData, err := exchData.ToGob()
// 	if err != nil {
// 		return err
// 	}
// 	answer := chatConnection(string(sendData), "forDB")
// 	exchData.FromGob(answer)
// 	return exchData.DecodeErr()
// }

//InsertAdminR inserts an administrator into admins table that includes a role
func InsertAdminR(table, role, name, email, password string) error {
	exchange := broker.Exchange{
		Table: table,
		Put:   []string{"name", "email", "hashed_password", "role"},
		Spec:  []string{},
		People: []broker.Person{
			broker.Person{
				Name:           name,
				Email:          email,
				HashedPassword: password,
				Role:           role,
			},
		},
		Action: "insert",
	}
	sendData, err := exchange.ToGob()
	if err != nil {
		return err
	}
	answer := chatConnection(string(sendData), "forDB")
	exchange.FromGob(answer)
	return exchange.DecodeErr()
}

//AuthenticateUserR gob encodes exchData and sends it to the dbmgr over nats
func AuthenticateUserR(table, role, email string) (*broker.Person, error) {
	exchange := broker.Exchange{
		Table: table,
		Put:   []string{},
		Spec:  []string{"role", "email"},
		People: []broker.Person{
			broker.Person{Role: role, Email: email},
		},
		Action: "get",
	}
	sendData, err := exchange.ToGob()
	if err != nil {
		return &broker.Person{}, err
	}
	answer := chatConnection(string(sendData), "forDB")
	exchange.FromGob(answer)
	length := len(exchange.People)
	if length == 1 {
		return &exchange.People[0], nil
	}
	return &broker.Person{}, fmt.Errorf("AuthenticateUserR brought back %d",
		length)
}

//GetUserR gob encodes exchData and sends it to the dbmgr over nats
func GetUserR(table string, id int) (*broker.Person, error) {
	exchange := broker.Exchange{
		Table: table,
		Put:   []string{},
		Spec:  []string{"id"},
		People: []broker.Person{
			broker.Person{ID: id},
		},
		Action: "get",
	}
	sendData, err := exchange.ToGob()
	if err != nil {
		return &broker.Person{}, err
	}
	answer := chatConnection(string(sendData), "forDB")
	exchange.FromGob(answer)
	length := len(exchange.People)
	if length == 1 {
		return &exchange.People[0], nil
	}
	return &broker.Person{}, fmt.Errorf("GetUserR brought back %d",
		length)
}

//GetByStatusR gets from the specified table a string agents by status (eg. active)
func GetByStatusR(table, role string, status bool) (*[]broker.Person, error) {
	exchange := broker.Exchange{
		Table: table,
		Put:   []string{},
		Spec:  []string{"role", "active"},
		People: []broker.Person{
			broker.Person{Active: status, Role: role},
		},
		Action: "get",
	}
	// centerr.InfoLog.Println("got to getByStatusR", exchData)
	sendData, err := exchange.ToGob()
	if err != nil {
		return nil, err
	}
	answer := chatConnection(string(sendData), "forDB")
	centerr.InfoLog.Println("got back from chatConnection")
	exchange.FromGob(answer)
	// centerr.InfoLog.Println("got to returning people", exchData.People)
	return &exchange.People, exchange.DecodeErr()
}

//ActivationR activates or deactivates agent or admin as requested.
func ActivationR(table, role string, people *[]broker.Person) error {
	exchange := broker.Exchange{
		Table:  table,
		Put:    []string{"active"},
		Spec:   []string{"id", "role"},
		People: *people,
		Action: "put",
	}
	sendData, err := exchange.ToGob()
	if err != nil {
		return err
	}
	answer := chatConnection(string(sendData), "forDB")
	exchange.FromGob(answer)
	// centerr.InfoLog.Println("got to returning people", exchData.People)
	return exchange.DecodeErr()
}

//ChgPwdR sends a request to the dbmgr to change the pawword for the specified email
func ChgPwdR(table, role, email, password string) error {
	exchange := broker.Exchange{
		Table: table,
		Put:   []string{"hashed_password"},
		Spec:  []string{"email", "role"},
		People: []broker.Person{
			broker.Person{HashedPassword: password, Email: email, Role: role},
		},
		Action: "put",
	}

	sendData, err := exchange.ToGob()
	if err != nil {
		return err
	}
	answer := chatConnection(string(sendData), "forDB")
	exchange.FromGob(answer)
	return exchange.DecodeErr()
}

//PutLine moves the agent offline and online
func PutLine(table, role string, id int, online bool) error {
	exchange := broker.Exchange{
		Table: table,
		Put:   []string{"online"},
		Spec:  []string{"id", "role"},
		People: []broker.Person{
			broker.Person{Online: online, ID: id, Role: role},
		},
		Action: "put",
	}
	sendData, err := exchange.ToGob()
	if err != nil {
		return err
	}
	answer := chatConnection(string(sendData), "forDB")
	exchange.FromGob(answer)
	return exchange.DecodeErr()

}

//sends string data to the far end, waits for the response and returns.
//for chat and mat, the data is string.  For dbmgr, the data is a struct.
//which is gob encoded before it is sent.  Gob encoder is in the broker pkg.
// TODO: find a way to build a nats connecton pool like the MySQL connection
//pool to speed up transactions.
func chatConnection(matValue, forCM string) []byte {
	var err error

	nc1, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		centerr.ErrorLog.Printf("in chatConnection connecting error %v", err)
	}
	defer nc1.Close()
	msg, err := nc1.Request(forCM, []byte(matValue), 2*time.Second)
	if err != nil {
		centerr.ErrorLog.Printf("in chatConnection %s request did not complete %v",
			forCM, err)
		return []byte{}
	}
	return msg.Data
}
