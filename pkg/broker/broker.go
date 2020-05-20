package broker

import (
	"bytes"
	"encoding/gob"
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

//ExchData for data exchange in gob format
type ExchData struct {
	ID             int
	Name           string
	Email          string
	Password       string
	Created        time.Time
	Active         bool //true is active, false is inactive
	HashedPassword []byte
	Authenticated  bool
	Action         string //authenticate, insert, and getuser are permitted actions
	ErrType        errMsg
	Err            string
	Table          string
	Role           string   //one of superadmin, admin, agent
	People         []Person //returns admins or agents by status (e.g. active)
	RowCnt         int      //should be the same as legnth of people field
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

//ToGob gob encodes ExchData and returns a byte slice
func (e *ExchData) ToGob() ([]byte, error) {
	b := &bytes.Buffer{}
	enc := gob.NewEncoder(b)
	err := enc.Encode(*e)
	if err != nil {
		return []byte{}, fmt.Errorf("failed gob Encode %v", err)
	}
	return b.Bytes(), nil
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

//FromGob decodes a gob byte slice into ExchData
func (e *ExchData) FromGob(g []byte) error {
	b := &bytes.Buffer{}
	b.Write(g)
	dec := gob.NewDecoder(b)
	err := dec.Decode(e)
	if err != nil {
		return fmt.Errorf("failed screen gob decode %v", err)
	}
	return nil
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

//EncodeErr encodes the error to avoid having to ship error by encoding it as
//gob (per Google group error type does not work with gob)
// func (e *ExchData) EncodeErr(err error) {
// 	if err == nil {
// 		e.ErrType = NoErr
// 		return
// 	}
// 	if errors.Is(err, ErrNoRecord) {
// 		e.ErrType = NoRecord
// 		return
// 	}
// 	if errors.Is(err, ErrInvalidCredentials) {
// 		e.ErrType = InvalidCreds
// 		return
// 	}
// 	if errors.Is(err, ErrDuplicateEmail) {
// 		return
// 	}
// 	e.ErrType = ErrZero
// 	e.Err = fmt.Sprintf("%v", err)
// }

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

//DecodeErr tries to decode the error out of the gob encoding process to its
//original form as best as it can especially preserving the three errors defined
//above
// func (e *ExchData) DecodeErr() error {
// 	switch e.ErrType {
// 	case NoErr:
// 		return nil
// 	case ErrZero:
// 		if e.Err == "" {
// 			return ErrNoRecord
// 		}
// 		return fmt.Errorf(e.Err)
// 	case NoRecord:
// 		return ErrNoRecord
// 	case InvalidCreds:
// 		return ErrInvalidCredentials
// 	case DuplicateMail:
// 		return ErrDuplicateEmail
// 	}
// 	return fmt.Errorf("error decoder failed %d", int(e.ErrType))
// }

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
