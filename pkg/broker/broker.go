package broker

import (
	"errors"
	"fmt"
	"time"
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

// TableProxy is to unify access to various database tables
type TableProxy interface {
	Length() int
	PickZero() (*Person, error)
	Pick() (*Dialog, error)
	AddToPeople(Person) People
	AddToDialog(Dialog) Dialogs
}

// TableEntry permits Person and Dialog to be processing in the same way.
type TableEntry interface {
	GetItems([]string) []interface{}
	GetBack([]string, []interface{}) error
}

//TableProxy is also of the type TableProxy
// type TableProxy []TableEntry

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

//People is a slice so multiple rows can be inserted and extracted
type People []Person

// Length is defined to handle indexing of TableProxy concrte type People
func (p People) Length() int {
	return len(p)
}

//PickZero accomodates the fact that you can't index an interface type
func (p People) PickZero() (*Person, error) {
	if p.Length() == 1 {
		return &p[0], nil
	}
	return &Person{}, fmt.Errorf("AuthenticateUserR brought back %d records",
		p.Length())
}

//Pick is a dummy method to satisfy TableProxy contract
func (p People) Pick() (*Dialog, error) {
	return nil, nil
}

// ReturnFirst is to handle indexing into TableProxy concrte type People
func (p People) ReturnFirst() Person {
	return p[0]
}

//AddToPeople provides the append functon for the interface
func (p People) AddToPeople(person Person) People {
	p = append(p, person)
	return p
}

//AddToDialog is a dummy method to satisfy the interface
func (p People) AddToDialog(dialog Dialog) Dialogs {
	return Dialogs{}
}

//Exchange is the new API interface to the dbmgr.  It is handed to the
//database get and put methods as is to be processed.
type Exchange struct {
	//Table is the name of the table to be processed, for now "admins in all cases."
	Table string
	//Put is list of column names to be put in the database after the SET verb
	//Put is only used by methods and functions that have "put" or "insert"
	//in the Action field
	Put []string
	//SpectList is a mirror image of Spec for building the SQL statement
	SpecList []string
	//Spec is the list of columns that come after the WHERE word
	//Spec will appear both when the Action is put or get but not insert
	Spec [][]interface{} //[]string
	//ScanSpec is specifically the specification for the rows.Scan statement
	ScanSpec [][]interface{}
	//Get is the list of the fields to be returned
	//Get is only used by methods or functions that set Action to "get"
	Get []string
	//See the notes on Person above.
	People TableProxy
	//Person is the extration of the single People
	//The command for the far end, get,  put, or insert.
	Action string

	ErrType errMsg
	Err     string
}

//InsertXR inserts an administrator into admins table that includes a role
//X stands for user, agent, or admin
func InsertXR(table, role, name, email, password string) error {
	people := People{
		Person{
			Name:           name,
			Email:          email,
			HashedPassword: password,
			Role:           role,
		},
	}
	exchange := Exchange{
		Table:  table,
		Put:    []string{"name", "email", "hashed_password", "created", "role"},
		People: people,
		Action: "insert",
	}
	for _, p := range people {
		c := p.BuildInsert(exchange.Put)
		exchange.Spec = append(exchange.Spec, c)
	}
	return exchange.runExchange()
}

//AuthenticateXR gob encodes exchData and sends it to the dbmgr over nats
//X stands for user, agent, or admin
func AuthenticateXR(table, role, email string) (*Person, error) {
	people := People{
		Person{Role: role, Email: email},
	}
	exchange := Exchange{
		Table:    table,
		Put:      []string{},
		SpecList: []string{"role", "email"},
		Get: []string{"id", "name", "email", "hashed_password", "created",
			"role", "active", "online"},
		People: people,
		Action: "get",
	}
	err := exchange.runGetExchange(people, exchange.SpecList)
	if err != nil {
		return &Person{}, err
	}
	person, err := exchange.People.PickZero()
	if err != nil {
		return person, err
	}
	return person, nil
}

//GetXR gob encodes exchData and sends it to the dbmgr over nats
//X stands for user, agent, or admin
func GetXR(table string, id int) (*Person, error) {
	people := People{Person{ID: id}}
	exchange := Exchange{
		Table:    table,
		Put:      []string{},
		SpecList: []string{"id"},
		Get: []string{"id", "name", "email", "hashed_password", "created",
			"role", "active", "online"},
		People: people,
		Action: "get",
	}
	err := exchange.runGetExchange(people, exchange.SpecList)
	if err != nil {
		return &Person{}, err
	}
	person, err := exchange.People.PickZero()
	if err != nil {
		return person, err
	}
	return person, nil
}

//GetByStatusR gets from the specified table a string agents by status (eg. active)
func GetByStatusR(table, role string, status bool) (People, error) {
	people := People{Person{Active: status, Role: role}}
	exchange := Exchange{
		Table:    table,
		Put:      []string{},
		SpecList: []string{"role", "active"},
		Get: []string{"id", "name", "email", "hashed_password", "created",
			"role", "active", "online"},
		People: people,
		Action: "get",
	}
	err := exchange.runGetExchange(people, exchange.SpecList)
	if err != nil {
		return nil, err
	}
	return exchange.People.(People), exchange.DecodeErr()
}

//ActivationR activates or deactivates agent or admin as requested.
func ActivationR(table, role string, people *People) error {
	exchange := Exchange{
		Table:    table,
		Put:      []string{"active"},
		SpecList: []string{"id", "role"},
		People:   *people,
		Action:   "put",
	}
	for _, p := range *people {
		c := p.Specify(exchange.Put, exchange.SpecList)
		exchange.Spec = append(exchange.Spec, c)
	}
	return exchange.runExchange()
}

//ChgPwdR sends a request to the dbmgr to change the pawword for the specified email
func ChgPwdR(table, role, email, password string) error {
	people := People{Person{HashedPassword: password, Email: email, Role: role}}
	exchange := Exchange{
		Table:    table,
		Put:      []string{"hashed_password"},
		SpecList: []string{"id", "role"},
		People:   people,
		Action:   "put",
	}
	for _, p := range people {
		c := p.Specify(exchange.Put, exchange.SpecList)
		exchange.Spec = append(exchange.Spec, c)
	}
	return exchange.runExchange()
}

//PutLine moves the agent offline and online
func PutLine(table, role string, id int, online bool) error {
	people := People{Person{Online: online, ID: id, Role: role}}
	exchange := Exchange{
		Table:    table,
		Put:      []string{"online"},
		SpecList: []string{"id", "role"},
		People:   people,
		Action:   "put",
	}
	for _, p := range people {
		c := p.Specify(exchange.Put, exchange.SpecList)
		exchange.Spec = append(exchange.Spec, c)
	}
	return exchange.runExchange()
}

//InsertEUR is for inserting end users (EU) from the front end
func InsertEUR(table, name, email, password string) error {
	people := People{
		Person{
			Name:           name,
			Email:          email,
			HashedPassword: password,
		},
	}
	exchange := Exchange{
		Table:  table,
		Put:    []string{Name, Email, HashedPassword, Created},
		People: people,
		Action: "insert",
	}
	for _, p := range people {
		c := p.BuildInsert(exchange.Put)
		exchange.Spec = append(exchange.Spec, c)
	}
	return exchange.runExchange()
}

//AuthenticateEUR gob encodes exchData and sends it to the dbmgr over nats
//EU stands for end user
func AuthenticateEUR(table, email string) (*Person, error) {
	people := People{Person{Email: email}}
	exchange := Exchange{
		Table:    table,
		Put:      []string{},
		SpecList: []string{"email"},
		Get:      []string{iD, Name, Email, HashedPassword, Created, Active},
		People:   people,
		Action:   "get",
	}
	err := exchange.runGetExchange(people, exchange.SpecList)
	if err != nil {
		return &Person{}, err
	}
	person, err := exchange.People.PickZero()
	if err != nil {
		return person, err
	}
	return person, nil
}

//GetEUR gets the user infromation for the database (EU for end user)
func GetEUR(table string, id int) (*Person, error) {
	people := People{Person{ID: id}}
	exchange := Exchange{
		Table:    table,
		Put:      []string{},
		SpecList: []string{"id"},
		Get: []string{iD, Name, Email, HashedPassword, Created,
			Active, Online},
		People: people,
		Action: "get",
	}
	err := exchange.runGetExchange(people, exchange.SpecList)
	if err != nil {
		return &Person{}, err
	}
	person, err := exchange.People.PickZero()
	if err != nil {
		return person, err
	}
	return person, nil
}
