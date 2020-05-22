package broker

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/saied74/toychat/pkg/centerr"
)

type errMsg int

//MoErr and the rest of this block is used for encoding and decoding error types
//through the gob encoding and decoding (error type does not work)
const (
	NoErr         errMsg = iota //no error
	ErrZero                     //simple error
	NoRecord                    //errNoRecord
	InvalidCreds                //errInvalidCredentials
	DuplicateMail               //errDuplicateEmail
)

var (
	//ErrNoRecord incicates there was not a corresponding record in the DB
	ErrNoRecord = errors.New("models: no matching record found")
	//ErrInvalidCredentials indicates that user supplied an invalid email
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	// ErrDuplicateEmail indicates that email is alreday being used
	ErrDuplicateEmail = errors.New("models: duplicate email")
)

//Person is a direct map of the database columns in exact the same order.
//data is extracted from the database into each of these fields.
//All "get" actions populate all fields.  For put actions, the fileds
//that need to be "put" must be populated.  The object Person is always
//used  in a slice as []People.
type Person struct {
	ID             int
	Name           string
	Email          string
	Password       string //once the new API is implemented, this field comes out.
	HashedPassword string
	Created        time.Time
	Role           string
	Active         bool
	Online         bool
}

//Exchange is the new API interface to the dbmgr.  It is handed to the
//database get and put methods as is to be processed.
type Exchange struct {
	//Table is the name of the table to be processed, for now "admins in all cases."
	Table string
	//Put is list of column names to be put in the database after the SET verb
	Put []string
	//Spec is the list of columns that come after the WhARE word
	Spec []string
	//See the notes on Person above.
	People []Person
	//The command for the far end, get,  put, or insert.
	Action string

	ErrType errMsg
	Err     string
}

//Specify builds inspec on the side of the gob decoding.  It uses people data
//and put and spec slices to build the items that need to be ither put into the
//database or are conditions of the database entry (WHERE condition.)
func (e *Exchange) Specify() [][]interface{} {
	inspec := [][]interface{}{}
	for _, person := range e.People {
		sp := []interface{}{}
		for _, p := range e.Put {
			switch p {
			case "id":
				sp = append(sp, person.ID)
			case "name":
				sp = append(sp, person.Name)
			case "email":
				sp = append(sp, person.Email)
			case "hashed_password":
				sp = append(sp, person.HashedPassword)
			case "role":
				sp = append(sp, person.Role)
			case "active":
				sp = append(sp, person.Active)
			case "online":
				sp = append(sp, person.Online)
			}
		}
		for _, s := range e.Spec {
			switch s {
			case "id":
				sp = append(sp, person.ID)
			case "name":
				sp = append(sp, person.Name)
			case "email":
				sp = append(sp, person.Email)
			case "hashed_password":
				sp = append(sp, person.HashedPassword)
			case "role":
				sp = append(sp, person.Role)
			case "active":
				sp = append(sp, person.Active)
			case "online":
				sp = append(sp, person.Online)
			}
		}
		inspec = append(inspec, sp)

	}
	return inspec
}

//ToGob encodes Exchange type data to be shipped over nats
func (e *Exchange) ToGob() ([]byte, error) {
	b := &bytes.Buffer{}
	enc := gob.NewEncoder(b)
	err := enc.Encode(*e)
	if err != nil {
		return []byte{}, fmt.Errorf("failed gob Encode %v", err)
	}
	return b.Bytes(), nil
}

//FromGob decides Exchange typed shipped over nats
func (e *Exchange) FromGob(g []byte) error {
	b := &bytes.Buffer{}
	b.Write(g)
	dec := gob.NewDecoder(b)
	err := dec.Decode(e)
	if err != nil {
		return fmt.Errorf("failed screen gob decode %v", err)
	}
	return nil
}

//EncodeErr encodes err for transmission over gob encoded medium.
//errors don't gob encode.
func (e *Exchange) EncodeErr(err error) {
	switch {
	case err == nil:
		e.ErrType = NoErr
	case errors.Is(err, ErrNoRecord):
		e.ErrType = NoRecord
	case errors.Is(err, ErrInvalidCredentials):
		e.ErrType = InvalidCreds
	case errors.Is(err, ErrDuplicateEmail):
		e.ErrType = DuplicateMail
	default:
		e.ErrType = ErrZero
	}
	e.Err = fmt.Sprintf("%v", err)
}

//DecodeErr decodes error shipped over gob encoded medium.
func (e *Exchange) DecodeErr() error {
	switch e.ErrType {
	case NoErr:
		return nil
	case ErrZero:
		if e.Err == "" {
			return ErrNoRecord
		}
		return fmt.Errorf(e.Err)
	case NoRecord:
		return ErrNoRecord
	case InvalidCreds:
		return ErrInvalidCredentials
	case DuplicateMail:
		return ErrDuplicateEmail
	}
	return fmt.Errorf("error decoder failed %d", int(e.ErrType))
}

//InsertAdminR inserts an administrator into admins table that includes a role
func InsertAdminR(table, role, name, email, password string) error {
	exchange := Exchange{
		Table: table,
		Put:   []string{"name", "email", "hashed_password", "role"},
		Spec:  []string{},
		People: []Person{
			Person{
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
func AuthenticateUserR(table, role, email string) (*Person, error) {
	exchange := Exchange{
		Table: table,
		Put:   []string{},
		Spec:  []string{"role", "email"},
		People: []Person{
			Person{Role: role, Email: email},
		},
		Action: "get",
	}
	sendData, err := exchange.ToGob()
	if err != nil {
		return &Person{}, err
	}
	answer := chatConnection(string(sendData), "forDB")
	exchange.FromGob(answer)
	err = exchange.DecodeErr()
	if err != nil {
		return &Person{}, err
	}
	length := len(exchange.People)
	if length == 1 {
		return &exchange.People[0], nil
	}
	return &Person{}, fmt.Errorf("AuthenticateUserR brought back %d records",
		length)
}

//GetUserR gob encodes exchData and sends it to the dbmgr over nats
func GetUserR(table string, id int) (*Person, error) {
	exchange := Exchange{
		Table: table,
		Put:   []string{},
		Spec:  []string{"id"},
		People: []Person{
			Person{ID: id},
		},
		Action: "get",
	}
	sendData, err := exchange.ToGob()
	if err != nil {
		return &Person{}, err
	}
	answer := chatConnection(string(sendData), "forDB")
	exchange.FromGob(answer)
	length := len(exchange.People)
	if length == 1 {
		return &exchange.People[0], nil
	}
	return &Person{}, fmt.Errorf("GetUserR brought back %d",
		length)
}

//GetByStatusR gets from the specified table a string agents by status (eg. active)
func GetByStatusR(table, role string, status bool) (*[]Person, error) {
	exchange := Exchange{
		Table: table,
		Put:   []string{},
		Spec:  []string{"role", "active"},
		People: []Person{
			Person{Active: status, Role: role},
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
func ActivationR(table, role string, people *[]Person) error {
	exchange := Exchange{
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
	exchange := Exchange{
		Table: table,
		Put:   []string{"hashed_password"},
		Spec:  []string{"email", "role"},
		People: []Person{
			Person{HashedPassword: password, Email: email, Role: role},
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
	exchange := Exchange{
		Table: table,
		Put:   []string{"online"},
		Spec:  []string{"id", "role"},
		People: []Person{
			Person{Online: online, ID: id, Role: role},
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
