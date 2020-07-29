package broker

import (
	"errors"
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

//TableRow is a direct map of the database columns in exact the same order.
//data is extracted from the database into each of these fields.
//All "get" actions populate all fields.  For put actions, the fileds
//that need to be "put" must be populated.  The object Person is always
//used  in a slice as []People.
type TableRow struct {
	ID             int
	DialogID       int
	AgentID        int
	MessageID      int
	Dialog         int //number of dialogs an agent is handling
	Name           string
	Email          string
	Password       string //once the new API is implemented, this field comes out.
	HashedPassword string
	Created        time.Time
	Ended          time.Time
	Role           string
	Active         bool
	Online         bool
	Msg            string
}

//TableRows is a slice so multiple rows can be inserted and extracted
type TableRows []TableRow

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
	Tables TableRows
	//Person is the extration of the single People
	//The command for the far end, get,  put, or insert.
	Action string

	ErrType errMsg
	Err     string
}

//InsertXR inserts an administrator into admins table that includes a role
//X stands for user, agent, or admin
func InsertXR(table, role, name, email, password string) error {
	people := TableRows{
		TableRow{
			Name:           name,
			Email:          email,
			HashedPassword: password,
			Role:           role,
		},
	}
	exchange := Exchange{
		Table:  table,
		Put:    []string{"name", "email", "hashed_password", "created", "role"},
		Tables: people,
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
func AuthenticateXR(table, role, email string) (*TableRow, error) {
	people := TableRows{
		TableRow{Role: role, Email: email},
	}
	exchange := Exchange{
		Table:    table,
		Put:      []string{},
		SpecList: []string{"role", "email"},
		Get: []string{"id", "name", "email", "hashed_password", "created",
			"role", "active", "online"},
		Tables: people,
		Action: "get",
	}
	err := exchange.runGetExchange(people, exchange.SpecList)
	if err != nil {
		return &TableRow{}, err
	}
	person := exchange.Tables[0]
	// if err != nil {
	// 	return person, err
	// }
	return &person, nil
}

//GetXR gob encodes exchData and sends it to the dbmgr over nats
//X stands for user, agent, or admin
func GetXR(table string, id int) (*TableRow, error) {
	people := TableRows{TableRow{ID: id}}
	exchange := Exchange{
		Table:    table,
		Put:      []string{},
		SpecList: []string{"id"},
		Get: []string{"id", "name", "email", "hashed_password", "created",
			"role", "active", "online"},
		Tables: people,
		Action: "get",
	}
	err := exchange.runGetExchange(people, exchange.SpecList)
	if err != nil {
		return &TableRow{}, err
	}
	person := exchange.Tables[0]
	if err != nil {
		return &person, err
	}
	return &person, nil
}

//GetByStatusR gets from the specified table a string agents by status (eg. active)
func GetByStatusR(table, role string, status bool) (TableRows, error) {
	people := TableRows{TableRow{Active: status, Role: role}}
	exchange := Exchange{
		Table:    table,
		Put:      []string{},
		SpecList: []string{"role", "active"},
		Get: []string{"id", "name", "email", "hashed_password", "created",
			"role", "active", "online"},
		Tables: people,
		Action: "get",
	}
	err := exchange.runGetExchange(people, exchange.SpecList)
	if err != nil {
		return nil, err
	}
	return exchange.Tables, exchange.DecodeErr()
}

//ActivationR activates or deactivates agent or admin as requested.
func ActivationR(table, role string, people *TableRows) error {
	exchange := Exchange{
		Table:    table,
		Put:      []string{"active"},
		SpecList: []string{"id", "role"},
		Tables:   *people,
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
	people := TableRows{TableRow{HashedPassword: password, Email: email, Role: role}}
	exchange := Exchange{
		Table:    table,
		Put:      []string{"hashed_password"},
		SpecList: []string{"id", "role"},
		Tables:   people,
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
	people := TableRows{TableRow{Online: online, ID: id, Role: role}}
	exchange := Exchange{
		Table:    table,
		Put:      []string{"online"},
		SpecList: []string{"id", "role"},
		Tables:   people,
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
	people := TableRows{
		TableRow{
			Name:           name,
			Email:          email,
			HashedPassword: password,
		},
	}
	exchange := Exchange{
		Table:  table,
		Put:    []string{Name, Email, HashedPassword, Created},
		Tables: people,
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
func AuthenticateEUR(table, email string) (*TableRow, error) {
	people := TableRows{TableRow{Email: email}}
	exchange := Exchange{
		Table:    table,
		Put:      []string{},
		SpecList: []string{"email"},
		Get:      []string{iD, Name, Email, HashedPassword, Created, Active},
		Tables:   people,
		Action:   "get",
	}
	err := exchange.runGetExchange(people, exchange.SpecList)
	if err != nil {
		return &TableRow{}, err
	}
	person := exchange.Tables[0] //.PickZero()
	// if err != nil {
	// 	return person, err
	// }
	return &person, nil
}

//GetEUR gets the user infromation for the database (EU for end user)
func GetEUR(table string, id int) (*TableRow, error) {
	people := TableRows{TableRow{ID: id}}
	exchange := Exchange{
		Table:    table,
		Put:      []string{},
		SpecList: []string{"id"},
		Get: []string{iD, Name, Email, HashedPassword, Created,
			Active, Online},
		Tables: people,
		Action: "get",
	}
	err := exchange.runGetExchange(people, exchange.SpecList)
	if err != nil {
		return &TableRow{}, err
	}
	person := exchange.Tables[0] //.PickZero()
	// if err != nil {
	// 	return person, err
	// }
	return &person, nil
}
