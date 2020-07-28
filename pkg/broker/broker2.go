//this file contains the end user broker methods

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

const (
	iD             = "id"
	Name           = "name"
	Email          = "email"
	HashedPassword = "hashed_password"
	Created        = "created"
	Role           = "role"
	Active         = "active"
	Online         = "online"
	AgentID        = "agent_id"
	DialogID       = "dialog_id"
	Message        = "message"
	Started        = "started"
	Ended          = "ended"
	Open           = "open"
)

//BuildInsert uses the "put" slice pattern to build an empty interface
//slice to be used with the INSERT statement
func (p *Person) BuildInsert(put []string) []interface{} {
	var c = []interface{}{}
	for _, sp := range put {
		switch sp {
		case iD:
			c = append(c, p.ID)
		case Name:
			c = append(c, p.Name)
		case Email:
			c = append(c, p.Email)
		case HashedPassword:
			c = append(c, p.HashedPassword)
		case Created:
			continue
		case Active:
			c = append(c, p.Active)
		case Online:
			c = append(c, p.Online)
		case Role:
			c = append(c, p.Role)
		}
	}
	return c
}

//GetSpec builds the specifications to provide to the WHERE clause of SQL
func (p *Person) GetSpec(spec []string) []interface{} {
	var g = []interface{}{}
	for _, sp := range spec {
		switch sp {
		case iD:
			g = append(g, p.ID)
		case Name:
			g = append(g, p.Name)
		case Email:
			g = append(g, p.Email)
		case HashedPassword:
			g = append(g, p.HashedPassword)
		case Created:
			g = append(g, p.Created)
		case Active:
			g = append(g, p.Active)
		case Online:
			g = append(g, p.Online)
		case Role:
			g = append(g, p.Role)
		}
	}
	return g
}

//GetItems uses the get string to generate an interface to be passed to the
//sql.Execute statement for the INSERT sql command.
func (p *Person) GetItems(get []string) []interface{} {
	var g = []interface{}{}
	for _, sp := range get {
		switch sp {
		case iD:
			g = append(g, &p.ID)
		case Name:
			g = append(g, &p.Name)
		case Email:
			g = append(g, &p.Email)
		case HashedPassword:
			g = append(g, &p.HashedPassword)
		case Created:
			g = append(g, &p.Created)
		case Role:
			g = append(g, &p.Role)
		case Active:
			g = append(g, &p.Active)
		case Online:
			g = append(g, &p.Online)
		}
	}
	return g
}

//GetBack reverses the get item and takes the interface items and gets the
//underlying data back.
func (p *Person) GetBack(get []string, g []interface{}) error {
	for i, sp := range get {
		switch sp {
		case iD:
			xID, ok := g[i].(*int)
			if !ok {
				return fmt.Errorf("ID (int) type assertion failed")
			}
			p.ID = *xID
		case Name:
			xName, ok := g[i].(*string)
			if !ok {
				return fmt.Errorf("Name (string) type assertion failed")
			}
			p.Name = *xName
		case Email:
			xEmail, ok := g[i].(*string)
			if !ok {
				return fmt.Errorf("Email (string) type assertion failed")
			}
			p.Email = *xEmail
		case HashedPassword:
			xPass, ok := g[i].(*string)
			if !ok {
				return fmt.Errorf("Hashed Password (string) type assertion failed")
			}
			p.HashedPassword = *xPass
		case Created:
			xCreated, ok := g[i].(*time.Time)
			if !ok {
				return fmt.Errorf("Created (time.Time) type assertion failed")
			}
			p.Created = *xCreated
		case Role:
			xRole, ok := g[i].(*string)
			if !ok {
				return fmt.Errorf("Role (string) type assertion failed")
			}
			p.Role = *xRole
		case Active:
			xActive, ok := g[i].(*bool)
			if !ok {
				return fmt.Errorf("Active (bool) type assertion failed")
			}
			p.Active = *xActive
		case Online:
			xOnline, ok := g[i].(*bool)
			if !ok {
				return fmt.Errorf("Online (bool) type assertion failed")
			}
			p.Online = *xOnline
		}
	}
	return nil
}

//Specify builds inspec on the side of the gob decoding.  It uses people data
//and put and spec slices to build the items that need to be ither put into the
//database or are conditions of the database entry (WHERE condition.)
func (p *Person) Specify(put, spec []string) []interface{} {
	sp := []interface{}{}
	for _, pu := range put {
		switch pu {
		case iD:
			sp = append(sp, p.ID)
		case Name:
			sp = append(sp, p.Name)
		case Email:
			sp = append(sp, p.Email)
		case HashedPassword:
			sp = append(sp, p.HashedPassword)
		case Role:
			sp = append(sp, p.Role)
		case Active:
			sp = append(sp, p.Active)
		case Online:
			sp = append(sp, p.Online)
		}
	}
	for _, s := range spec {
		switch s {
		case iD:
			sp = append(sp, p.ID)
		case Name:
			sp = append(sp, p.Name)
		case Email:
			sp = append(sp, p.Email)
		case HashedPassword:
			sp = append(sp, p.HashedPassword)
		case Role:
			sp = append(sp, p.Role)
		case Active:
			sp = append(sp, p.Active)
		case Online:
			sp = append(sp, p.Online)
		}
	}
	return sp
}

//repeated code at the end of each send - recieve function.
func (e *Exchange) runExchange() error {
	gob.Register(time.Time{})
	gob.Register(People{})
	gob.Register(Dialogs{})
	sendData, err := e.ToGob()
	if err != nil {
		return err
	}
	answer := ChatConnection(sendData, "forDB")
	e.FromGob(answer)
	return e.DecodeErr()
}

func (e *Exchange) runGetExchange(people People, getSpec []string) error {
	for _, p := range people {
		c := p.GetSpec(getSpec)
		e.Spec = append(e.Spec, c)
	}
	err := e.runExchange()
	if err != nil {
		return err
	}
	return e.DecodeErr()
}

//ToGob encodes Exchange type data to be shipped over nats
func (e *Exchange) ToGob() ([]byte, error) {
	// start := time.Now()
	b := &bytes.Buffer{}
	enc := gob.NewEncoder(b)
	err := enc.Encode(*e)
	if err != nil {
		return []byte{}, fmt.Errorf("failed gob Encode %v", err)
	}
	// end := time.Now()
	// centerr.InfoLog.Printf("togob: time difference %v", end.Sub(start))
	return b.Bytes(), nil
}

//FromGob decides Exchange typed shipped over nats
func (e *Exchange) FromGob(g []byte) error {
	// start := time.Now()
	gob.Register(time.Time{})
	gob.Register(People{})
	gob.Register(Dialogs{})
	b := &bytes.Buffer{}
	b.Write(g)
	dec := gob.NewDecoder(b)
	err := dec.Decode(e)
	if err != nil {
		return fmt.Errorf("failed screen gob decode %v", err)
	}
	// end := time.Now()
	// centerr.InfoLog.Printf("fromgob: time difference %v", end.Sub(start))
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

//ChatConnection sends string data to the far end, waits for the response and returns.
//for chat and mat, the data is string.  For dbmgr, the data is a struct.
//which is gob encoded before it is sent.  Gob encoder is in the broker pkg.
// TODO: find a way to build a nats connecton pool like the MySQL connection
//pool to speed up transactions.
func ChatConnection(sendMsg []byte, target string) []byte {
	var err error

	nc1, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		centerr.ErrorLog.Printf("in chatConnection connecting error %v", err)
	}
	defer nc1.Close()
	msg, err := nc1.Request(target, sendMsg, 2*time.Second)
	if err != nil {
		centerr.ErrorLog.Printf("in chatConnection %s request did not complete %v",
			target, err)
		return []byte{}
	}
	return msg.Data
}
