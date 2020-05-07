package broker

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"time"
)

type errMsg int

//MpErr amd the rest of this block is used for encoding and decoding error types
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

//ExchData for data exchange in gob format
type ExchData struct {
	ID             int
	Name           string
	Email          string
	Password       string
	Created        time.Time
	Active         bool
	HashedPassword []byte
	Authenticated  bool
	Action         string //authenticate, insert, and getuser are permitted actions
	ErrType        errMsg
	Err            string
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

//EncodeErr encodes the error to avoid having to ship error by encoding it as
//gob (per Google group error type does not work with gob)
func (e *ExchData) EncodeErr(err error) {
	if err == nil {
		e.ErrType = NoErr
		return
	}
	if errors.Is(err, ErrNoRecord) {
		e.ErrType = NoRecord
		return
	}
	if errors.Is(err, ErrInvalidCredentials) {
		e.ErrType = InvalidCreds
		return
	}
	if errors.Is(err, ErrDuplicateEmail) {
		return
	}
	e.ErrType = ErrZero
	e.Err = fmt.Sprintf("%v", err)
}

//DecodeErr tries to decode the error out of the gob encoding process to its
//original form as best as it can especially preserving the three errors defined
//above
func (e *ExchData) DecodeErr() error {
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
